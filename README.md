# Proof of concept - Predict OCP cluster upgrade failures

## Initial Setup and configuration

### Setting up a local graph instance
- Clone this repo OUTSIDE your gopath as with: `git clone —recursive -submodules -j8 https://github.com/fabric8-analytics/poc-ocp-upgrade-prediction`
- Run [Installation script](./scripts/install-graph.sh) to setup your gremlin (one time).
- Run [automation script](./scripts/run_graph.sh) to run your gremlin every time you want to bring up the graph to run this code.
- Now create JanusGraph indices for faster node creation, as with `go run scripts/populate_janusgraph_schema.go`
- Alternatively just give a remote gremlin instance to `GREMLIN_REST_URL`

### Environment Variables
- Make sure gremlin server is running.
- Need to set the following environment variables: 
```json5
            {
                "GREMLIN_REST_URL": "http://localhost:8182", // The API endpoint for the Gremlin server.
                "GOPATH": "GOPATH", // The Gopath on the current machine that you're working off of.
                "GH_TOKEN": "YOUR_GH_TOKEN", // Github token, should contain all repo permissions. This is required to fork the projects into a namespace for you
                "KUBECONFIG": "PATH_TO_KUBECONFIG", // Path to a kubeconfig, openshift-install binary should have generated this for you under "auth/" folder wherever you ran the cluster installation.
                "KUBERNETES_SERVICE_PORT": 6443, // Port on which your Kubernetes cluster API is running, this is generally 6443 AFAIK.
                "KUBERNETES_SERVICE_HOST": "PATH_TO_YOUR_DEV_CLUSTER_API", // Path to a running Kubernetes cluster that we need to run the end to end tests/service end to end tests. Is of the form api.*.devcluster.openshift.com
                "REMOTE_SERVER_URL": "" // This is the path to the layer running the origin end-to-end test wrapper.
            }
```

### Build instructions

- You need a fairly recent version of Go (1.11.x, 1.12.x)
- You'll need to increase your git max buffer size: `git config http.postBuffer 524288000`
- For the python bits, Python 3.6 is required. You would additionally need `Flask` and `requests`. Install these with `pip install flask requests`
- This project uses the new vgo(go modules) dependency management system, simply running `make build` with fetch all the dependencies
- Install all the Go binaries with: `make install`

## Artifacts: Go

### Compile time flow creation: clustergraph
- First create the compile time paths using the clustergraph flow, if you you are using cluster version  as with: `$GOPATH/bin/clustergraph --cluster-version=4.0.0-0.ci-2019-04-23-214213`
- Make sure the index creation outlined in the final step of the [first phase](#setting-up-a-local-graph-instance) has been done otherwise this would be painfully slow.
- Currently, in order for this to work for just one service there's a `break` statement at the end of the control block [here](https://github.com/fabric8-analytics/poc-ocp-upgrade-prediction/blob/master/cmd/clustergraph/clustergraph.go#L67). Remove it to create the graph for the entire payload.

### Component end to end test node creation flow for a PR: api 
- This spins up a REST API as with: `$GOPATH/bin/api`
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
* Optionally specify the `--no-images` flag so that it doesn't bother with docker image creation.
* This binary will create a custom payload based off an already existing payload for an OCP release. Sample usage:

```bash
$ $GOPATH/bin/custompayload-create --cluster-version=4.0.0-0.ci-2019-04-15-000954 --destdir=/tmp --no-images # This version won't work, it's outdated. Pick one from the ocp releases page.
```
* This will, in you `destdir` or your current working directory create a directory inside which all the services will be cloned and patched.
* This'll take a long time.

### Patched openshift tests to run serially: openshift-tests

* Fork this, you need to clone: `https://github.com/rootavish/origin` (not included as a submodule here)
* Compile the binary, as with `make WHAT=cmd/openshift-tests`
* The binary would be available under `_output/bin/{linux/darwin}/amd64/` and can be run with `./openshift-tests`

## Artifacts: Python

### Fork all payload repositories to a namespace(org/user): scripts/github_forker.py

This script forks all the missing repositories that are mentioned in the payload to a namespace that you own, this is required for pushing changes to them so they can be picked up by the CI operator. Currently only works with the `--org` flag set and requires you to create an organization on Github. Sample usage:

```bash
$ python github_forker.py --namespace poc-ocp-upgrades --cluster-version=4.0.0-0.ci-2019-04-22-163416 --org=true
```
### Push all changes for tracing source code to fork: scripts/commit_sources.py

This script will commit all the changes made by `clusterpatcher` script and push them to github to our forks so that it can be picked by the ci-operator. Sample usage is almost the same as github_forker with the exception that `--no-verify=true` needs to be set if using global `git-secrets` as some repositories contain dummy AWS keys for running tests.
```bash
$python commit_sources.py --namespace poc-ocp-upgrades --cluster-version=4.0.0-0.ci-2019-04-22-163416 --org=true --no-verify=true
```

### Wrapper over gremlin for e2e product runtime node creation: e2e_logger_api.py

Run this with `python e2e_logger_api.py`, should start the flask development server at port `5001`.

## Current Limitations

* Does not parse dynamic function calls such as anonymous functions from a map because they are mapped at runtime and were too much work for POC [like _bindata here](https://github.com/openshift/machine-config-operator/blob/master/pkg/operator/assets/bindata.go#L1195).


## License and contributing

See [LICENSE](LICENSE)  

To contribute to this project Just send a PR.
