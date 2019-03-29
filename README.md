# Proof of concept - Predict OCP cluster upgrade failures

## Setup & running the project

- Clone this repo OUTSIDE your gopath as with: `git clone https://github.com/fabric8-analytics/poc-ocp-upgrade-prediction`
- Init submodules as with: `git submodule init && git submodule update`
- Run [automation script](./scripts/automation.sh to setup and start your gremlin)
- Make sure gremlin server is running.
- Set the following environment variables: 
```json
            "env": {
                "GREMLIN_REST_URL": "http://localhost:8182",
                "GOPATH": "GOPATH",
                "GH_TOKEN": "YOUR_GH_TOKEN",
                "KUBECONFIG": "PATH_TO_KUBECONFIG",
                "KUBERNETES_SERVICE_PORT": 6443,
                "KUBERNETES_SERVICE_HOST": "PATH_TO_YOUR_DEV_CLUSTER_API"
            }
```
- Build the project with: `make build`
- Install both the binaries as with: `make install`
- Now create JanusGraph indices for faster node creation, as with `go run scripts/populate_janusgraph_schema.go`
- First create the compile time paths using the clustergraph flow, if you you are using cluster version  as with: `$GOPATH/bin/clustergraph [`PATH_TO_FOLDER_CONTAINING_CLUSTER_VERSION_FILE`] [DIR_FOR_CLONING_REPOS]`
- Sample `cluster_version.json` is located inside this repo. `PATH_TO_FOLDER_CONTAINING_CLUSTER_VERSION_FILE` can be the path where this source is cloned.
- Then run the REST API as with: `$GOPATH/bin/api.go`
- Send a get request to the REST API, the PR and repo are hardcoded for now runtimepaths will be created. Here's a sample request:
```bash
curl -X GET \
  http://localhost:8080 \
  -H 'Content-Type: application/json' \
  -H 'Postman-Token: b9c125ed-8d5f-4481-9251-1ee42d44a723' \
  -H 'cache-control: no-cache' \
  -d '{
    "pr_id": 482,
    "repo_url": "openshift/machine-config-operator/"
}'
```

## Current Limitations

* Does not parse dynamic function calls such as anonymous functions from a map because they are mapped at runtime and were too much work for POC [like _bindata here](https://github.com/openshift/machine-config-operator/blob/master/pkg/operator/assets/bindata.go#L1195).

## License and contributing

See [LICENSE](LICENSE)
Just send a PR.