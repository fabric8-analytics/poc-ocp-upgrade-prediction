# Proof of concept - Predict OCP cluster upgrade failures

## Initial Setup and configuration

### Setting up a local graph instance
- Clone this repo OUTSIDE your gopath as with: `git clone --recursive-submodules -j8 https://github.com/fabric8-analytics/poc-ocp-upgrade-prediction`
- Run [Installation script](./scripts/install-graph.sh) to setup your gremlin (one time)
- Run [automation script](./scripts/run_graph.sh) to run your gremlin every time you want to bring up the graph to run this code.
- Now create JanusGraph indices for faster node creation, as with `go run scripts/populate_janusgraph_schema.go`

### Environment Variables
- Make sure gremlin server is running.
- Need to set the following environment variables: 
```json5
            {
                "GREMLIN_REST_URL": "http://localhost:8182", // The API endpoint for the Gremlin server.
                "GOPATH": "GOPATH", // The Gopath on the current machine that you're working off of.
                "GH_TOKEN": "YOUR_GH_TOKEN", // Github token, should contain all repo permissions. This is required to fork the projects into a namespace for you
                "KUBECONFIG": "PATH_TO_KUBECONFIG", // Path to a 
                "KUBERNETES_SERVICE_PORT": 6443,
                "KUBERNETES_SERVICE_HOST": "PATH_TO_YOUR_DEV_CLUSTER_API",
                "REMOTE_SERVER_URL": "" // This is the path to the layer running the
            }
```

### Build instructions

- You need a fairly recent version of Go (1.11.x, 1.12.x)
- This project uses the new vgo(go modules) dependency management system, simply running `make build` with fetch all the dependencies
- Install all the Go binaries with: `make install`

## Artifacts

### Compile time flow creation: clustergraph
- First create the compile time paths using the clustergraph flow, if you you are using cluster version  as with: `$GOPATH/bin/clustergraph --cluster-version=`
- Make sure the index creation outlined in the final step of the [first phase](#### Setting up a local graph instance) has been done otherwise this would be painfully slow.
- Currently, in order for this to work for just one service there's a `break` statement at the end of the control block. Remove it to create the graph for the entire payload.

### Component end to end test node creation flow for a PR: api 
- Then run the REST API as with: `$GOPATH/bin/api.go`
- Send a get request to the REST API, the PR and repo are hardcoded for now runtimepaths will be created. Here's a sample request:
```bash
curl -X GET \
  http://localhost:8080/api/v1/createprnode \
  -H 'Content-Type: application/json' \
  -H 'Postman-Token: b9c125ed-8d5f-4481-9251-1ee42d44a723' \
  -H 'cache-control: no-cache' \
  -d '{
    "pr_id": 482,
    "repo_url": "openshift/machine-config-operator/"
}'
```

### Payload creation for running the end to end tests: custompayload-creator

* Follow the installation procedure to install all the binaries from above
* Make sure to login to registry.svc.ci.openshift.org(your ~/.docker/config.json should have a token for registry.svc.ci.openshift.org)
* This binary will create a custom payload based off an already existing payload for an OCP release. Sample usage:

```bash
$ $GOPATH/bin/custompayload-create --cluster-version=4.0.0-0.ci-2019-04-15-000954 # This version won't work, it's outdated. Pick one from the ocp releases page.
```
This will, in you current working directory create a directory inside which all the services will be cloned and patched. It is alternatively possible to give it a different destination dir with the `--destdir` argument.

## Current Limitations

* Does not parse dynamic function calls such as anonymous functions from a map because they are mapped at runtime and were too much work for POC [like _bindata here](https://github.com/openshift/machine-config-operator/blob/master/pkg/operator/assets/bindata.go#L1195).


## License and contributing

See [LICENSE](LICENSE)  

To contribute to this project Just send a PR.
