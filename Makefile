dev-build:
	docker build --target dev -t leagueify-api-dev .

dev-clean: dev-stop
	docker image rm leagueify-api-dev

dev-start: dev-build
	docker compose --profile dev up

dev-stop:
	docker compose --profile dev down -v

format:
	go fmt ./...

init:
	go get .

prod-build:
	docker build -t leagueify-api .

prod-clean: prod-stop
	docker image rm leagueify-api

prod-start: prod-build
	docker compose --profile prod up

prod-stop:
	docker compose --profile prod down -v

test:
	mkdir -p testCoverage
	go test ./... -cover -coverprofile=testCoverage/report.out
	go tool cover -html=testCoverage/report.out -o testCoverage/report.html

vet:
	go vet ./...
