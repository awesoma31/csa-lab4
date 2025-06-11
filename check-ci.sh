#!/bin/bash

echo "--- Running Go fmt/imports check ---"
# Проверяем, есть ли неформатированные файлы или неорганизованные импорты
if ! goimports -l .; then
  echo "ERROR: Go files not formatted with goimports or imports are unorganized."
  echo "Please run 'goimports -w .' on your local machine to fix."
  exit 1
fi
echo "Go fmt/imports check passed."

echo "--- Running golangci-lint ---"
# Запускаем golangci-lint. Если у вас есть .golangci.yml, он будет использован автоматически.
if ! golangci-lint run; then
  echo "ERROR: golangci-lint found issues."
  echo "Please fix the linting errors."
  exit 1
fi
echo "golangci-lint passed."

echo "--- Running Go Tests ---"
if ! go test ./... -v; then
  echo "ERROR: Go tests failed."
  exit 1
fi
echo "Go Tests passed."

echo "--- All local CI checks passed successfully! ---"
