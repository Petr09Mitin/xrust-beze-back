dev-init:
	#docker-compose up -d
#	sleep 4 # wait for init all services
#	docker-compose exec mongo_db sh -c "mongo < /scripts/mongo-init.js"
	#sleep 4
	#docker-compose exec mongo_db sh -c "mongo < /scripts/init-db.js"
	docker-compose exec mongo_db sh -c "mongo < /scripts/init-indexes.js"

start:
	docker build -t xrust_beze_chat:latest -f cmd/chat/Dockerfile . \
	&& docker build -t ml_explanator:latest -f ml_explanator/Dockerfile . \
	&& docker build -t ml_check:latest -f ml_check/Dockerfile . \
	&& docker build -t xrust_beze_user:latest -f cmd/user/Dockerfile . \
	&& docker build -t xrust_beze_file:latest -f cmd/file/Dockerfile . \
	&& docker-compose up

stop:
	docker-compose down

proto:
	protoc --go_out=. --go-grpc_out=. proto/user/user.proto && protoc --go_out=. --go-grpc_out=. proto/file/file.proto

# микросервис user
build-user:
	docker build -t xrust_beze_user:latest -f cmd/user/Dockerfile .

start-user-only: build-user
	docker-compose up -d mongo_db user_service
