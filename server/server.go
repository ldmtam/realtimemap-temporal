package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"go.temporal.io/sdk/client"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Server struct {
	ctx context.Context
	srv *http.Server
}

func NewHttpServer(ctx context.Context, redisCli *redis.Client, temporalClient client.Client) *Server {
	router := gin.Default()

	serveAPI(router, redisCli, temporalClient)

	srv := &http.Server{
		Addr:    ":12345",
		Handler: router,
	}

	return &Server{
		ctx: ctx,
		srv: srv,
	}
}

func (s *Server) ListenAndServe() <-chan bool {
	done := make(chan bool)
	go s.listenAndServe(done)
	return done
}

func (s *Server) listenAndServe(done chan<- bool) {
	go func() {
		defer func() {
			done <- true
		}()
		if err := s.srv.ListenAndServe(); err != nil {
			return
		}
	}()

	<-s.ctx.Done()
	slog.Info("Shutting down http server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.srv.Shutdown(shutdownCtx)
}
