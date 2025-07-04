dev-init:
	#docker-compose up -d
#	sleep 4 # wait for init all services
#	docker-compose exec mongo_db sh -c "mongo < /scripts/mongo-init.js"
	#sleep 4
	#docker-compose exec mongo_db sh -c "mongo < /scripts/init-db.js"
	docker-compose exec mongo_db sh -c "mongo < /scripts/init-indexes.js"

clear-mongo:
	docker compose exec mongo_db sh -c "mongo < /scripts/clear-db.js"

start:
	docker build -t petr09mitin/xrust_beze_chat:latest -f cmd/chat/Dockerfile . \
	&& docker build -t petr09mitin/ml_explanator:latest -f ml_explanator/Dockerfile . \
	&& docker build -t petr09mitin/ml_image_moderator:latest -f ml_image_moderator/Dockerfile . \
	&& docker build -t petr09mitin/ml_moderator:latest -f ml_moderator/Dockerfile . \
	&& docker build -t petr09mitin/xrust_beze_user:latest -f cmd/user/Dockerfile . \
	&& docker build -t petr09mitin/xrust_beze_file:latest -f cmd/file/Dockerfile . \
	&& docker build -t petr09mitin/xrust_beze_auth:latest -f cmd/auth/Dockerfile . \
	&& docker build -t petr09mitin/xrust_beze_study_material:latest -f cmd/study_material/Dockerfile . \
	&& docker build -t petr09mitin/xrust_beze_studymateriald:latest -f cmd/studymateriald/Dockerfile . \
	&& docker build -t petr09mitin/xrust_beze_voicerecognitiond:latest -f cmd/voicerecognitiond/Dockerfile . \
	&& docker-compose up

stop:
	docker-compose down

PROTO_FILES := proto/user/user.proto proto/file/file.proto proto/auth/auth.proto proto/study_material/study_material.proto
proto: $(PROTO_FILES)
	protoc --go_out=. --go-grpc_out=. proto/user/user.proto
	protoc --go_out=. --go-grpc_out=. proto/file/file.proto
	protoc --go_out=. --go-grpc_out=. proto/auth/auth.proto
	protoc --go_out=. --go-grpc_out=. proto/study_material/study_material.proto

build-user:
	docker build -t petr09mitin/xrust_beze_user:latest -f cmd/user/Dockerfile .

build-file:
	docker build -t petr09mitin/xrust_beze_file:latest -f cmd/file/Dockerfile .

build-moderator:
	docker build -t petr09mitin/ml_moderator:latest -f ml_moderator/Dockerfile .

build-image-moderator:
	docker build -t petr09mitin/ml_image_moderator:latest -f ml_image_moderator/Dockerfile .

build-ai-tags:
	docker build -t petr09mitin/ai_tags:latest -f ai_tags/Dockerfile .

start-ai-tags:
	docker-compose up ai_tags minio-xb

build-RAG:
	docker build -t petr09mitin/rag_service:latest -f RAG_service/Dockerfile .

start-RAG:
	docker-compose up RAG_service minio-xb ml_explanator mongo_db

build-explane:
	docker build -t petr09mitin/ml_explanator:latest -f ml_explanator/Dockerfile .

build-transcript:
	docker build -t petr09mitin/transcript:latest -f transcript/Dockerfile .

start-transcript:
	docker-compose up transcript minio-xb

start-user-only: build-user build-file
	docker-compose up -d mongo_db user_service auth_service study_material

build-auth:
	docker build -t petr09mitin/xrust_beze_auth:latest -f cmd/auth/Dockerfile .

start-auth-only: build-user build-auth build-moderator
	docker-compose up -d redis_xb auth_service ml_moderator

build-study-material:
	docker build -t petr09mitin/xrust_beze_study_material:latest -f cmd/study_material/Dockerfile .

start-study-material-only: build-study-material # build-user build-file build-auth 
	docker-compose up -d mongo_db auth_service user_service file_service study_material
