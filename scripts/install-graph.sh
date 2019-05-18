#!/bin/bash
set +ex

git submodule init
git submodule update

SCRIPTPATH=$(dirname "$SCRIPT")
echo "Scriptpath: ${SCRIPTPATH}"
echo "Running mvn install"
mvn install -f $SCRIPTPATH/dynamodb-janusgraph-storage-backend/pom.xml > /dev/null

echo "Installing gremlin server"
cd $SCRIPTPATH/dynamodb-janusgraph-storage-backend/
echo "We are in dir: `pwd`"
./src/test/resources/install-gremlin-server.sh

echo "Changing channelizer for gremlin server"
sed -i  's/WebSocketChannelizer/WsAndHttpChannelizer/g' server/dynamodb-janusgraph-storage-backend-1.2.0/conf/gremlin-server/gremlin-server-local.yaml
