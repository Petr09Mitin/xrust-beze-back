dev-init:
	docker-compose up -d --force-recreate
	sleep 4 # wait for init all services
	docker-compose exec mongo_db sh -c "mongo < /scripts/mongo-init.js"
	sleep 4
	docker-compose exec mongo_db sh -c "mongo < /scripts/init-db.js"
	# init kafka
	docker-compose exec kafka sh -c "sh /scripts/setup.sh"

start:
	docker build -t xrust_beze:latest -f cmd/chat/Dockerfile . \
	&& docker-compose up

stop:
	docker-compose down
