APP_NAME = redditclone2

.DEFAULT_GOAL := run

build:
	export GOWORK=off && \
	go build -mod=vendor -v -o ./bin/${APP_NAME} ./cmd/${APP_NAME}

run: build
	./bin/${APP_NAME}

test_handlers:
	go test ./internal/handlers -coverprofile=./internal/handlers/cover.out
	go tool cover -html=./internal/handlers/cover.out -o ./internal/handlers/cover.html

test_post:
	go test ./internal/post -coverprofile=./internal/post/cover.out
	go tool cover -html=./internal/post/cover.out -o ./internal/post/cover.html

test_user:
	go test ./internal/user -coverprofile=./internal/user/cover.out
	go tool cover -html=./internal/user/cover.out -o ./internal/user/cover.html

test:
	go test -v -coverpkg=./... -coverprofile=cover.out ./...
	go tool cover -html=cover.out -o cover.html

.PHONY: clean
clean:
	go clean
	rm -rf ./bin/${APP_NAME}

compose_up:
	docker compose up -d

compose_stop:
	docker compose stop

compose_rm:
	docker rm $$(docker ps -a -q) -v
	docker volume prune -f

mysql_access:
	mysql -h 127.0.0.1 -P 3306 -u root -padmin

mongo_access:
	docker exec -it redditclone2-mongodb-1 mongosh "mongodb://localhost:27017"