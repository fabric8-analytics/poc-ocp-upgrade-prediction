# Proof of concept - Predict OCP cluster upgrade failures

### Modules
* service_parser: Parse the source code of a service (supplied as git source
    via command line argument) and return the functions mapped to individual
    packages in a JSON.

* utils: Only contains a utility to extract git commit stuff for now
* gremlin: Contains the gremlin queries for interaction with JanusGraph
* ghpr
* traceappend

### TODO

* Enhancements to the graph schema
* Make TODO
