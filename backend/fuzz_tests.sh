#!/bin/bash

# Запуск фаззинг-тестов для API
echo "Running API fuzzing tests..."
go test -fuzz=FuzzRegisterHandler -fuzztime=10s ./internal/api
go test -fuzz=FuzzLoginHandler -fuzztime=10s ./internal/api
go test -fuzz=FuzzAuthMiddleware -fuzztime=10s ./internal/api
go test -fuzz=FuzzGetUserLogs -fuzztime=10s ./internal/api
go test -fuzz=FuzzLogEntryJSON -fuzztime=10s ./internal/api

# Запуск фаззинг-тестов для Auth
echo "Running Auth fuzzing tests..."
go test -fuzz=FuzzCreateUser -fuzztime=10s ./internal/auth
go test -fuzz=FuzzValidateUser -fuzztime=10s ./internal/auth
go test -fuzz=FuzzAuthMiddleware -fuzztime=10s ./internal/auth
go test -fuzz=FuzzAddLog -fuzztime=10s ./internal/auth
go test -fuzz=FuzzGetLogs -fuzztime=10s ./internal/auth
go test -fuzz=FuzzGetLogsByTimeRange -fuzztime=10s ./internal/auth
go test -fuzz=FuzzClearLogs -fuzztime=10s ./internal/auth
go test -fuzz=FuzzGetUsernameFromContext -fuzztime=10s ./internal/auth

echo "All fuzzing tests completed!" 