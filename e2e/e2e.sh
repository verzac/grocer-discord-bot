set -x
rm main
rm db/gorm.db
set -e
go build -o main
./main &
set +e
make e2e
exit_code=$?
kill $!
rm main
exit $exit_code