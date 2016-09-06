#!/bin/bash -ex
go install
$GOPATH/bin/todomvc-backend-go \
  -postgres_credentials_path postgres_credentials.dev.json
