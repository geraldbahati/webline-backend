#!/usr/bin/env bash

set -e

host="$1"
user="$2"
shift 2

until pg_isready -h "$host" -U "$user"; do
  >&2 echo "Postgres is unavailable - sleeping"
  sleep 1
done

>&2 echo "Postgres is up - executing command"
exec "$@"
