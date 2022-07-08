SRC=./src
E2E=./e2e
OUT=./bin/server

.PHONY: all
all: test e2e build

.PHONY: build
build:
	go build -o "${OUT}" "${SRC}"

.PHONY: clean
clean:
	go clean
	rm -f "${OUT}"

.PHONY: test
test: _require_dotenv
	godotenv -f ./env/.env.test,./env/.env.default go test "${SRC}/..."

.PHONY: e2e
e2e: _require_dotenv
	godotenv -f ./env/.env.test,./env/.env.default go test "${E2E}/..."

.PHONY: run
run: _require_dotenv
	make build
	godotenv -f ./env/.env.local,./env/.env.default "./${OUT}"

.PHONY: dev
dev: _require_gow
	make -j _dev_containers _dev_watch

.PHONY: _require_dotenv
_require_dotenv:
	if [ -z `which godotenv` ]; then echo "installing godotenv..."; go install github.com/joho/godotenv/cmd/godotenv@latest; echo "done!"; fi

.PHONY: _require_gow
_require_gow:
	if [ -z `which gow` ]; then echo "installing gow..."; go install github.com/mitranim/gow@latest; echo "done!"; fi

.PHONY: _dev_containers
_dev_containers:
	godotenv -f ./env/.env.local,./env/.env.default docker compose -f dev/docker-compose.yml up

.PHONY: _dev_watch
_dev_watch:
	godotenv -f ./env/.env.local,./env/.env.default gow -e="go,mod,html" run ./src
