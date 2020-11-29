#!/bin/bash

cfg_json=$(cat ./internal/modules/user/init/init-cfg.json)
curl -X PUT -H "Content-Type: application/json" -d "$cfg_json" http://localhost:8500/v1/kv/config/user
migrate run --db-port 26257 --dir ./internal/modules/user/migrate --operation up --db-pass "" --db-user user_service --db-name user_db