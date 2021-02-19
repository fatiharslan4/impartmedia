#!/usr/bin/env bash

set -e
docker-compose down -v --remove-orphans
docker-compose rm -f -v
docker-compose up
docker-compose down -v --remove-orphans
docker-compose rm -f -v