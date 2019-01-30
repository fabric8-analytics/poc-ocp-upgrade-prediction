#! /bin/bash
set +ex

git submodule init
git submodule update

cd ./dynamodb-janusgraph-storage-backend
mvn install
./src/test/resources/install-gremlin-server.sh
sed -i '' 's/WebSocketChannelizer/WsAndHttpChannelizer/g' ./server/dynamodb-janusgraph-storage-backend-1.2.0/conf/gremlin-server/gremlin-server-local.yaml

mvn test -Pstart-dynamodb-local &
./server/dynamodb-janusgraph-storage-backend-1.2.0/bin/gremlin-server.sh ./server/dynamodb-janusgraph-storage-backend-1.2.0/conf/gremlin-server/gremlin-server-local.yaml &
cd ..

export GOPATH=$GOPATH:`pwd`
go get github.com/tidwall/gjson
export GREMLIN_REST_URL="http://localhost:8182"
