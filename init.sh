#!/bin/sh

# init user service
sql_user_init=$(cat internal/modules/user/init/init-users-db.sql)
docker exec -it user-db cockroach sql --insecure --execute="$sql_user_init"
cfg_json=$(cat internal/modules/user/init/init-cfg.json)
curl -X PUT -H "Content-Type: application/json" -d "$cfg_json" http://localhost:8500/v1/kv/config/user
