dev-init:
	docker-compose up -d
#	sleep 4 # wait for init all services
#	docker-compose exec mongo_db sh -c "mongo < /scripts/mongo-init.js"
	sleep 4
	docker-compose exec mongo_db sh -c "mongo < /scripts/init-db.js"

start:
	docker build -t xrust_beze_chat:latest -f cmd/chat/Dockerfile . \
	&& docker build -t ml_explanator:latest -f ml_explanator/Dockerfile . \
	&& docker build -t ml_check:latest -f ml_check/Dockerfile . \
	&& docker-compose up

stop:
	docker-compose down
