#!/bin/sh
# helper script to run E2E tests against a local build of the bot
set -x
rm main
rm db/gorm.db
set -e
go build -o main
./main &
set +e
make e2e
exit_code=$?
if [ "$exit_code" -eq 0 ]; then
  sqlite3 db/gorm.db \
    'INSERT INTO api_clients (created_at, updated_at, created_by_id, client_id, client_secret, scope) VALUES (CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, "e2e-api", "e2e-api", "$2y$10$.Qp5rC5CVatOexCAW5atju5KILAM.ZK/ZYgvBF6yhmGa6QZVpMAZi", "guild:e2e-api");'
  E2E_API_HEADER="Basic ZTJlLWFwaTplMmUtYXBpLXNlY3JldA==" make e2e_api
  exit_code=$?
fi
kill $!
rm main
exit $exit_code
