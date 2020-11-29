export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64

build:
	chmod +x ./builds.sh && ./builds.sh

init:
	chmod +x ./builds.sh && ./builds.sh
	consul agent -dev -background

init2:
	chmod +x ./builds.sh && ./builds.sh
	docker-compose up --build -d
	sleep 5
	chmod +x ./init.sh && ./init.sh
	chmod +x ./migrate.sh && ./migrate.sh

dev-env-up:
	./builds.sh
	docker-compose up --build -d
	sleep 3
	chmod +x ./migrate.sh
	./migrate.sh

dev-env-stop:
	docker-compose stop

dev-env-restart:
	docker-compose stop
	docker-compose up -d

clear:
	docker-compose down --volumes
	rm -rf volume/