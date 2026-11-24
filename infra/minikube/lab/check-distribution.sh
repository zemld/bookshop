#!/bin/sh

for request in $(seq 1 500); do
  curl -sS -o /dev/null "${1:-http://127.0.0.1:8088/books}"
done
