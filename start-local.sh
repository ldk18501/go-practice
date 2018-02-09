#!/bin/bash

docker run -d \
  -v /data/mongodb/go:/data/db:rw \
  --name go-mongo mongo:3.4.10

go_mongo_ip=$(docker inspect --format '{{ .NetworkSettings.IPAddress }}' go-mongo)

docker run -d \
  -e MONGO_URL=$go_mongo_ip:27017 \
  --name go-server \
  -p 8888:8888 \
  go-test
