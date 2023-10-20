start-redis:
	docker-compose -f redis.yml up

stop-redis:
	docker-compose -f redis.yml down