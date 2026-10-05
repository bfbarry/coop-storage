#!/bin/bash
set -e

docker restart metadata-server
docker restart osd-server

# run this first
# cd && bash dev_create_tables.sh


sleep 2  # give containers time to fully start

go run e2e.go
