dev-init:
	docker-compose up -d
#	sleep 4 # wait for init all services
#	docker-compose exec mongo_db sh -c "mongo < /scripts/mongo-init.js"
	sleep 4
	docker-compose exec mongo_db sh -c "mongo < /scripts/init-db.js"

start:
	docker build -t xrust_beze:latest -f cmd/chat/Dockerfile . \
	&& docker-compose up
#	&& docker build -t ml_summarizer:latest -f ml_summarizer/Dockerfile . \

stop:
	docker-compose down
