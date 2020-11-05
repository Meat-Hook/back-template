#!/bin/sh

sql_init=$(cat ./internal/modules/user/init/init-db.sql)
docker exec -it user-db cockroach sql --insecure --execute="$sql_init"
chmod +x ./internal/modules/user/init/migrate.sh
./internal/modules/user/init/migrate.sh
