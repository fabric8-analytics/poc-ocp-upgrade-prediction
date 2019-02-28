package gremlin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/fabric8-analytics/poc-ocp-upgrade-prediction/pkg/serviceparser"
	"io/ioutil"
	"net/http"
	"os"

	"go.uber.org/zap"
)

var logger, _ = zap.NewProduction()
var sugarLogger = logger.Sugar()

// RunQuery runs the specified gremling query and returns its result.
func RunQuery(query string) map[string]interface{} {
	payload := map[string]interface{}{
		"gremlin": query,
	}
	payloadJSON, _ := json.Marshal(payload)
	response, err := http.Post(os.Getenv("GREMLIN_REST_URL"), "application/json", bytes.NewBuffer(payloadJSON))

	if err != nil {
		sugarLogger.Fatal(err)
	}

	var result map[string]interface{}

	json.NewDecoder(response.Body).Decode(&result)
	return result
}

// ReadJSON reads the contents of a JSON and returns it as a map[string]interface{}
func ReadJSON(jsonFilepath string) string {
	b, err := ioutil.ReadFile(jsonFilepath) // just pass the file name
	if err != nil {
		sugarLogger.Fatal(err)
	}
	return string(b)
}

//ReadFile reads the contents of a text file and return it as a string
func ReadFile(filepath string) string {
	b, err := ioutil.ReadFile(filepath)
	if err != nil {
		sugarLogger.Fatal(err)
	}
	return string(b)
}

// CreateNewServiceVersionNode creates a new service node for a codebase. DO NOT CALL THIS FUNCTION
// WITHOUT A CLUSTER VERSION NODE IN CONTEXT
func CreateNewServiceVersionNode(serviceName string, version string) string {
	query := fmt.Sprintf(`
		serviceVersion = g.addV('service_version').property('name', '%s').property('version', '%s').next();
		clusterVersion.addEdge('contains_service_at_version', serviceVersion);`, serviceName, version)
	return query
}

// CreateNewPackageNode creates a new package node and joins it using an edge
// to the parent service node.
func CreateNewPackageNode(packagename string) string {
	query := fmt.Sprintf(`packageNode = g.addV('package').property('name', '%s').next();
	serviceVersion.addEdge('contains_package', packageNode);`, packagename)
	return query
}

// CreateFunctionNodes adds function nodes to the graph and an edge between it and it's
// parent service and it's package
// DO NOT CALL CreateNewPackageNode BEFORE YOU'VE ENTERED ALL THE NODES FOR A SERVICE
func CreateFunctionNodes(functionNames []string) string {
	var fullQuery string
	for _, fn := range functionNames {
		query := fmt.Sprintf(`functionNode = g.addV('function').property('name', '%s').next();
							  packageNode.addEdge('has_fn', functionNode);`, fn)
		fullQuery += query
	}
	return fullQuery
}

// CreateClusterVerisonNode creates the top level cluster version node
// CALL THIS JUST ONCE PER RUN OF THIS SCRIPT, THAT IS HOW THIS CODE IS DESIGNED.
func CreateClusterVerisonNode(clusterVersion string) string {
	query := fmt.Sprintf(`
		clusterVersion = g.addV('clusterVersion').property('cluter_version', '%s').next()`, clusterVersion)
	return query
}

// RunGroovyScript takes the path to a groovy script and runs it at the Gremlin console.
func RunGroovyScript(scriptPath string) {
	scriptContent := ReadFile(scriptPath)
	gremlinResponse := RunQuery(scriptContent)
	sugarLogger.Info(gremlinResponse)
}

// CreateDependencyNode creates the nodes that contain the external dependency information for the
// service and connects it to the packages as well as the functions directly.
func CreateDependencyNodes(clusterVersion string, serviceName string, serviceVersion string, ic []serviceparser.ImportContainer) {
	query := fmt.Sprintf(
		`serviceNode = g.V().hasLabel('service_version').has('name', '%s').has('version', '%s').next();`, serviceName, serviceVersion)

	for _, imported := range ic {
		query += fmt.Sprintf(`importNode = g.addV('dependency').property('local_name', '%s').property('importpath', '%s').next();
				  servicenode.addEdge(importNode, 'depends_on');`, imported.LocalName, imported.ImportPath)
	}

	sugarLogger.Infof("%v\n", RunQuery(query))
}
