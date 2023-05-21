APP_NAME = asperitas

.DEFAULT_GOAL := run

.PHONY: build
build:
	go build -mod=vendor -v -o ./bin/${APP_NAME} ./cmd/${APP_NAME}

.PHONY: run
run: build
	./bin/${APP_NAME}

.PHONY: test_handlers
test_handlers:
	go test ./internal/handlers -coverprofile=./internal/handlers/cover.out
	go tool cover -html=./internal/handlers/cover.out -o ./internal/handlers/cover.html

.PHONY: test_post
test_post:
	go test ./internal/post -coverprofile=./internal/post/cover.out
	go tool cover -html=./internal/post/cover.out -o ./internal/post/cover.html

.PHONY: test_user
test_user:
	go test ./internal/user -coverprofile=./internal/user/cover.out
	go tool cover -html=./internal/user/cover.out -o ./internal/user/cover.html

.PHONY: test
test:
	go test -v -coverpkg=./... -coverprofile=cover.out ./...
	go tool cover -html=cover.out -o cover.html

.PHONY: clean
clean:
	go clean
	rm -rf ./bin/${APP_NAME}

.PHONY: compose_up
compose_up:
	docker compose up -d

.PHONY: compose_stop
compose_stop:
	docker compose stop

.PHONY: compose_rm
compose_rm:
	docker rm $$(docker ps -a -q) -v
	docker volume prune -f

.PHONY: mysql_access
mysql_access:
	mysql -h 127.0.0.1 -P 3306 -u root -padmin

.PHONY: mongo_access
mongo_access:
	docker exec -it asperitas-mongodb-1 mongosh "mongodb://localhost:27017"
