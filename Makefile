export GOOS=linux

init:
	chmod +x ./builds.sh && ./builds.sh
	docker-compose up --build -d
	sleep 5
	chmod +x ./init.sh && ./init.sh

dev-env:
	docker-compose up --build -d

restart:
	docker-compose stop
	docker-compose up -d
