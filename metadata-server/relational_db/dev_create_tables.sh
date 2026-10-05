#!/bin/bash

docker exec -i -e PGPASSWORD=password psql psql -U user -d appdb < schema.sql
