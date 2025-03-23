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
	&& docker build -t xrust_beze_user:latest -f cmd/user/Dockerfile . \
	&& docker-compose up -d

stop:
	docker-compose down

build-user:
	docker build -t xrust_beze_user:latest -f cmd/user/Dockerfile .

# Команда для запуска только микросервиса пользователей (опционально)
start-user-only: build-user
	docker-compose up -d mongo_db user_service
