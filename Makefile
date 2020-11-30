export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64

init:
	chmod +x ./builds.sh && ./builds.sh
	docker-compose up --build -d

dev-env-up:
	docker-compose up --build -d

dev-env-down:
	docker-compose down

dev-env-restart: dev-env-down dev-env-up
