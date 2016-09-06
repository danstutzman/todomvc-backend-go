#!/bin/bash -ex
go install
$GOPATH/bin/todomvc-backend-go -in_memory_db -socket_path '/tmp/echo.sock'
