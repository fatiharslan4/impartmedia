#!/bin/bash


docker-compose down -v
docker-compose up -d bootstrap-mysql
docker wait backend_bootstrap-mysql_1 > /dev/null
script/run_migrations.sh up
yes | script/run_migrations.sh down #check bidirection
script/run_migrations.sh up
script/generate_db_code.sh