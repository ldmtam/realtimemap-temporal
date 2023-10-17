package ingress

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"log/slog"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	brokerURL = "ssl://mqtt.hsl.fi:8883"
	clientID  = "realtimemap-temporal"
)

type Payload struct {
	Longitude *float64   `json:"long"`
	Latitude  *float64   `json:"lat"`
	Heading   *int32     `json:"hdg"`
	DoorState *int32     `json:"drst"`
	Timestamp *time.Time `json:"tst"`
	Speed     *float64   `json:"spd"`
}

func (p *Payload) HasValidPosition() bool {
	return p != nil &&
		p.Latitude != nil && p.Longitude != nil &&
		p.Heading != nil && p.Timestamp != nil &&
		p.Speed != nil && p.DoorState != nil
}

type Event struct {
	VehiclePosition *Payload `json:"VP"`
	DoorOpen        *Payload `json:"DOO"`
	DoorClosed      *Payload `json:"DOC"`
	VehicleId       string
	OperatorId      string
}

func ConsumeVehicleEvents(onEvent func(*Event), ctx context.Context) <-chan bool {
	done := make(chan bool)
	go func() {
		var f mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
			event := &Event{}

			//0/1       /2        /3             /4              /5           /6               /7            /8               /9         /10            /11        /12          /13         /14             /15       /16
			// /<prefix>/<version>/<journey_type>/<temporal_type>/<event_type>/<transport_mode>/<operator_id>/<vehicle_number>/<route_id>/<direction_id>/<headsign>/<start_time>/<next_stop>/<geohash_level>/<geohash>/<sid>/#
			topicParts := strings.Split(msg.Topic(), "/")

			if err := json.Unmarshal(msg.Payload(), event); err != nil {
				slog.Error("Error unmarshalling json", err)
			} else {
				event.OperatorId = topicParts[7]
				event.VehicleId = topicParts[7] + "." + topicParts[8]
				onEvent(event)
			}
		}

		opts := mqtt.NewClientOptions().
			AddBroker(brokerURL).
			SetClientID(clientID).
			SetDefaultPublishHandler(f).
			SetKeepAlive(2 * time.Second).
			SetPingTimeout(1 * time.Second).
			SetCleanSession(true)

		client := mqtt.NewClient(opts)
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
		slog.Info("CONNECTED")

		if token := client.Subscribe("/hfp/v2/journey/ongoing/vp/bus/#", 0, nil); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
		slog.Info("SUBSCRIBED")

		<-ctx.Done()

		if token := client.Unsubscribe("/hfp/v2/journey/ongoing/vp/bus/#"); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
		slog.Info("UNSUBSCRIBED")

		client.Disconnect(250)
		slog.Info("DISCONNECTED")

		done <- true
	}()

	return done
}
