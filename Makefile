users-db-sql-init = `cat init/init-users-db.sql`

dev-env:
	docker-compose up --build -d
	docker exec -it user-db cockroach sql --insecure --execute="$(users-sql-init)"
