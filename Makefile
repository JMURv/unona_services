dev:
	docker-compose up

pb:
	protoc --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative api/pb/services.proto

dev:
	docker-compose -f docker-compose.yml up

dev-db:
	docker run --rm --name postgres_notifications_db \
	-p 5432:5432 \
	-e POSTGRES_PASSWORD=794613825Zx! \
	-e POSTGRES_DB=notifications_db \
	-e POSTGRES_USER=postgres \
	-v pg_data_users:/var/lib/postgresql/data \
	postgres:15.0-alpine

kafka:
	docker run --rm --name=kafka -p 9092:9092 apache/kafka:3.7.0

redis:
	docker run --rm --name=redis -p 6379:6379 redis:alpine

jaeger:
	docker run --rm --name jaeger \
      -e COLLECTOR_ZIPKIN_HOST_PORT=:9411 \
      -p 6831:6831/udp \
      -p 6832:6832/udp \
      -p 5778:5778 \
      -p 16686:16686 \
      -p 4317:4317 \
      -p 4318:4318 \
      -p 14250:14250 \
      -p 14268:14268 \
      -p 14269:14269 \
      -p 9411:9411 \
      jaegertracing/all-in-one:latest

prometheus:
	docker run --rm --name prometheus \
		-p 9090:9090 \
		-v $PWD/prometheus.yml:/etc/prometheus/prometheus.yml \
		-v prometheus-data:/prometheus \
 		prom/prometheus:latest

node-exp:
	docker run --rm --name node-exp -p 9100:9100 prom/node-exporter

grafana:
	docker run --rm --name grafana \
		-p 3000:3000 \
		-v grafana-data:/var/lib/grafana \
		grafana/grafana:latest