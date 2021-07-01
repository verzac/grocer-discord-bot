set -x
rm main
set -e
go build -o main
./main &
set +e
make e2e
exit_code=$?
kill $!
rm main
exit $exit_code