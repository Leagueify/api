build:
	docker build -t leagueify-api .

build-dev:
	docker build --target dev -t leagueify-api-dev .

clean: stop
	docker compose down -v

clean-dev: stop-dev
	docker compose --profile dev down -v

format:
	go fmt ./...

init:
	go get .

prep: format vet

start: build
	docker compose --profile prod up

start-detached: build
	docker compose --profile prod up -d

start-dev: build-dev
	docker compose --profile dev up

start-dev-detached: build-dev
	docker compose --profile dev up -d

stop:
	docker compose --profile prod down -v

stop-dev:
	docker compose --profile dev down -v

test:
	mkdir -p testCoverage
	go test ./... -cover -coverprofile=testCoverage/report.out
	go tool cover -html=testCoverage/report.out -o testCoverage/report.html

vet:
	go vet ./...
