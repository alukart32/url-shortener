pg_name = postgres14
pg_user = postgres
pg_user_pass = postgres
pg_image = postgres:14-alpine
pg_uri = localhost:5432
db_name = shortener


.PHONY: help
help:
	@echo List of params:
	@echo   db_name                 - postgres docker container name (default: $(pg_name))
	@echo   pq_user                 - postgres root user (default: $(pg_user))
	@echo   pq_user_pass            - postgres root user password (default: $(pg_user_pass))
	@echo   db_image                - postgres docker image (default: $(pg_image))
	@echo   db_uri                  - postgres uri (default: $(pg_uri))
	@echo   db_name                 - postgres main db (default: $(db_name))
	@echo List of commands:
	@echo   postgres-up             - run postgres postgres docker container $(pg_name)
	@echo   postgres-up             - down postgres postgres docker container $(pg_name)
	@echo   create-db               - create db $(db_name)
	@echo   drop-db                 - drop db $(db_name)
	@echo   run-with-config         - run app with ./configs/config.json
	@echo   run-with-memstorage     - run app with memstorage
	@echo   run-with-filestorage    - run app with filestorage
	@echo   run-with-db             - run app with db
	@echo   test                    - run all tests
	@echo   test-cover              - show test coverage
	@echo   gen                     - gen resources
	@echo   help                    - help info
	@echo   clear                   - truncate resources
	@echo Usage:
	@echo                           make `cmd_name`

.PHONY: run-with-config
run-with-config:
	go run ./cmd/shortener/main.go -c "./configs/config.json"

.PHONY: run-with-memstorage
run-with-memstorage:
	go run ./cmd/shortener/main.go -a "localhost:8080" -b "http://localhost:8080" -t "192.168.1.0/24"

.PHONY: run-with-filestorage
run-with-filestorage:
	go run ./cmd/shortener/main.go -a "localhost:8080" -b "http://localhost:8080" -f "file.db" -t "192.168.1.0/24"

.PHONY: postgres-up
postgres-up:
	docker run --name $(pg_name) -e POSTGRES_USER=$(pg_user) -e POSTGRES_PASSWORD=$(pg_user_pass) -p 5432:5432 -d $(pg_image)

.PHONY: postgres-stop
postgres-stop:
	docker stop $(pg_name)

.PHONY: create-db
create-db:
	docker exec -it $(pg_name) createdb --username=$(pg_user) --owner=$(pg_user) $(db_name)

.PHONY: drop-db
drop-db:
	docker exec -it $(pg_name) dropdb --username=$(pg_user) $(db_name)

.PHONY: run-with-db
run-with-db:
	go run ./cmd/shortener/main.go -a "localhost:8080" -b "http://localhost:8080" -d "postgres://$(pg_user):$(pg_user_pass)@localhost:5432/$(db_name)" -t "192.168.1.0/24"

.PHONY: test
test:
	go test ./internal/... -coverprofile cover.out

.PHONY: test-cover
test-cover: test
	go tool cover -func cover.out

.PHONY: go-gen
go-gen:
	go generate ./...

.PHONY: grpc-gen-v1
grpc-gen-v1:
	protoc --go_out=. --go_opt=paths=import --go-grpc_out=. --go-grpc_opt=paths=import api/v1/proto/urls.proto
	protoc --go_out=. --go_opt=paths=import --go-grpc_out=. --go-grpc_opt=paths=import api/v1/proto/stat.proto
	protoc --go_out=. --go_opt=paths=import --go-grpc_out=. --go-grpc_opt=paths=import api/v1/proto/observ.proto

.PHONY: docker-clear
docker-clear: drop-db postgres-stop
	docker rm $(pg_name)

.PHONY: saticcheck
saticcheck:
	.\cmd\staticlint\staticlint.exe ./...