dev-init:
	docker-compose up -d --force-recreate
	sleep 4 # wait for init all services
	docker-compose exec mongo_db sh -c "mongo < /scripts/mongo-init.js"
	sleep 4
	docker-compose exec mongo_db sh -c "mongo < /scripts/init-db.js"
	# init kafka
	docker-compose exec kafka sh -c "sh /scripts/setup.sh"

# скиллы можно добавить вот так:
# docker-compose exec user_service sh -c \
# 	"[ $$(mongo xrust_beze --quiet --eval 'db.skills.countDocuments()') -eq 0 ] && \
# 	 go run scripts/load_skills_to_mongo.go || echo 'Skills already loaded'"

# или так: загрузка навыков в mongo (только в случае, если коллекция skills пуста)
load-skills:
	go run scripts/load_skills_to_mongo.go

start:
	docker build -t xrust_beze:latest -f cmd/chat/Dockerfile . \
	&& docker build -t xrust_beze_user:latest -f cmd/user/Dockerfile . \
	&& docker-compose up -d

stop:
	docker-compose down

proto:
	protoc --go_out=. --go-grpc_out=. proto/user/user.proto

# микросервис user
build-user:
	docker build -t xrust_beze_user:latest -f cmd/user/Dockerfile .

start-user-only: build-user
	docker-compose up -d mongo_db user_service
