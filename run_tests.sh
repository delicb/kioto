#!/bin/sh

set -e

KIOTO_INTEGRATION_TEST=true

echo "mode: count" > profile.cov
for d in $(go list ./...); do
  go test -coverprofile=./coverage.out -covermode=count -v "$d"
  tail -n +2 ./coverage.out >> profile.cov
  rm ./coverage.out;
done
go tool cover -func profile.cov
go tool cover -html profile.cov -o coverage.html
