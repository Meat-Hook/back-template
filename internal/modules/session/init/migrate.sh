#!/bin/bash

cfg_json=$(cat ./internal/modules/session/init/init-cfg.json)
curl -X PUT -H "Content-Type: application/json" -d "$cfg_json" http://localhost:8500/v1/kv/config/session
migrate run --db-port 25555 --dir ./internal/modules/session/migrate --operation up --db-pass "" --db-user session_service --db-name session_db
