#!/bin/bash

sql_init=$(cat ./internal/modules/session/init/init-db.sql)
docker exec -it session-db cockroach sql --insecure --execute="$sql_init"
chmod +x ./internal/modules/session/init/migrate.sh
./internal/modules/session/init/migrate.sh

docker exec -it 9ad16d9b184b cockroach sql --insecure --execute="$(cat ./internal/modules/user/init/init-db.sql)"