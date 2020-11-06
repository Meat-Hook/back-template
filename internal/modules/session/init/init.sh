#!/bin/bash

sql_init=$(cat ./internal/modules/session/init/init-db.sql)
docker exec -it session-db cockroach sql --insecure --execute="$sql_init"
chmod +x ./internal/modules/session/init/migrate.sh
./internal/modules/session/init/migrate.sh
