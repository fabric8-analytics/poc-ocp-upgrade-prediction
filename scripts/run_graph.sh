#!/bin/bash
set +ex

echo "Starting dynamodb and Janus"
cd dynamodb-janusgraph-storage-backend
mvn test -Pstart-dynamodb-local > /dev/null &

echo "Waiting..."
sleep 10
echo "Starting gremlin"
cd server/dynamodb-janusgraph-storage-backend-1.2.0
echo "We are in dir: `pwd`"
bin/gremlin-server.sh conf/gremlin-server/gremlin-server-local.yaml > /dev/null &

echo "Started gremlin"
export GOPATH=$GOPATH:`pwd`
export GREMLIN_REST_URL="http://localhost:8182"
