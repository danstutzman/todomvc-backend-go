#!/bin/bash -ex
go run backend.go handlers.go \
  -in_memory_db \
  -socket_path '/tmp/echo.sock'
