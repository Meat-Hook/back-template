#!/bin/bash

migrate run --db-port 25555 --dir ./internal/modules/session/migrate --operation up --db-pass "root" --db-user root --db-name postgres
