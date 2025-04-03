#!/bin/sh
set -e

# Проверяем, что передана команда
if [ "$#" -eq 0 ]; then
  echo "Usage: docker-compose run goose [goose-command]"
  echo "Example: docker-compose run goose up"
  exit 1
fi

# Выполняем команду goose с переданными аргументами
exec goose "$@"