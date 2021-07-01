set -x
rm main
set -e
go build -o main
./main &
set +e
make e2e
kill $!
rm main