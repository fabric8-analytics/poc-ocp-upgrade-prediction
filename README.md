# Proof of concept - Predict OCP cluster upgrade failures

### Setup & running the project

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
- Send a get request to the REST API as with: `http://localhost:8080/v1/api/prcoverage`, the PR and repo are hardcoded for now runtimepaths will be created.


### Current Limitations

* Does not parse dynamic function calls such as anonymous functions from a map because they are mapped at runtime and were too much work for POC [like _bindata here](https://github.com/openshift/machine-config-operator/blob/master/pkg/operator/assets/bindata.go#L1195).
