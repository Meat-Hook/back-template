#!/bin/bash

migrate run --db-port 26257 --dir ./internal/modules/user/migrate --operation up --db-pass "root" --db-user root --db-name postgres
