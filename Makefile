export GOOS=linux

init:
	chmod +x ./builds.sh && ./builds.sh
	docker-compose up --build -d
	sleep 5
	chmod +x ./init.sh && ./init.sh
	chmod +x ./migrate.sh && ./migrate.sh

dev-env-up:
	docker-compose up --build -d
	sleep 3
	chmod +x ./migrate.sh
	./migrate.sh

dev-env-stop:
	docker-compose stop

dev-env-restart:
	docker-compose stop
	docker-compose up -d
