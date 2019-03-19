package gremlin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/fabric8-analytics/poc-ocp-upgrade-prediction/pkg/serviceparser"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var logger, _ = zap.NewDevelopment()
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

	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		sugarLogger.Errorf("Failed to decode JSON: %v\n", err)
	}
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
func CreateNewServiceVersionNode(clusterVersion, serviceName, version string)  {
	query := fmt.Sprintf(`
		clusterVersion = g.V().hasLabel('clusterVersion').has('cluster_version', '%s').next();
		serviceVersion = g.addV('service_version').property('name', '%s').property('version', '%s').next();
		clusterVersion.addEdge('contains_service', serviceVersion);`, clusterVersion, serviceName, version)
	sugarLogger.Debug(query)
	sugarLogger.Debugf("%v\n", RunQuery(query))
}

// NewPackageNodeQuery creates a new package node and joins it using an edge
// to the parent service node.
func NewPackageNodeQuery(serviceName, serviceVersion, packagename string) string {
	query := fmt.Sprintf(`
	serviceVersion = g.V().hasLabel('service_version').has('name', '%s').has('version', '%s').next();
	packageNode = g.addV('package').property('name', '%s').next();
	serviceVersion.addEdge('contains_package', packageNode);`, serviceName, serviceVersion, packagename)
	return query
}

// CreateFunctionNodes adds function nodes to the graph and an edge between it and it's
// parent service and it's package
// DO NOT CALL NewPackageNodeQuery BEFORE YOU'VE ENTERED ALL THE NODES FOR A PACKAGE
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
func CreateClusterVerisonNode(clusterVersion string) {
	query := fmt.Sprintf(`
		clusterVersion = g.addV('clusterVersion').property('cluster_version', '%s').next()`, clusterVersion)
	sugarLogger.Debugf("%v\n%v\n", query, RunQuery(query))
}

// RunGroovyScript takes the path to a groovy script and runs it at the Gremlin console.
func RunGroovyScript(scriptPath string) {
	scriptContent := ReadFile(scriptPath)
	gremlinResponse := RunQuery(scriptContent)
	sugarLogger.Info(gremlinResponse)
}

// CreateDependencyNodes creates the nodes that contain the external dependency information for the
// service and connects it to the packages as well as the functions directly.
func CreateDependencyNodes(clusterVersion, serviceName, serviceVersion string, ic []serviceparser.ImportContainer) {
	queryBase := fmt.Sprintf(
		`serviceNode = g.V().has('cluster_version', '%s').out().hasLabel('service_version').has('name', '%s').has('version', '%s').next();`, clusterVersion, serviceName, serviceVersion)

	query := queryBase
	for idx, imported := range ic {
		query += fmt.Sprintf(`importNode = g.addV('dependency').property('local_name', '%s').property('importpath', '%s').next();
				  serviceNode.addEdge('depends_on', importNode);
				  packageNode = g.V().hasLabel('package').has('name', '%s').next();
				  importNode.addEdge('affects_package', packageNode);`, imported.LocalName, imported.ImportPath, imported.DependentPkg)

		// Running this query in batches.
		if idx % 30 == 0 {
			gremlinResponse := RunQuery(query)
			sugarLogger.Debugf("%v\n%v\n", query, gremlinResponse)
			query = queryBase
		}

	}

	// Any remainders
	if query != queryBase {
		gremlinResponse := RunQuery(query)
		sugarLogger.Debugf("%v\n%v\n", query, gremlinResponse)
	}
}

// CreateCompileTimeFlows adds all the compile time code flow edges to the graph.
func CreateCompileTimeFlows(clusterVersion, serviceName, serviceVersion string, paths map[string]interface{}) {
	for _, pathStruct := range paths {
		pathArr, ok := pathStruct.([]serviceparser.CodePath)
		if !ok {
			sugarLogger.Fatalf("Did not get a valid codepath.")
			os.Exit(-1)
		}
		for _, path := range pathArr {
			// First find the service
			query := fmt.Sprintf(
				`serviceNode = g.V().has('cluster_version', '%s').out().hasLabel('service_version').has('name', '%s').has('version', '%s').next();`, clusterVersion, serviceName, serviceVersion)
			// The "From" function will always be a part of the service.
			query = query + fmt.Sprintf(`fromFunc = g.V(serviceNode).out().hasLabel('package').has('name', '%s').out().hasLabel('function').has('name', '%s').next();`, path.ContainerPackage, path.From)
			// If there is no selector for the called function function is most assumed to be defined in same package.
			_, selectorName := filepath.Split(path.SelectorCallee)

			if path.SelectorCallee == "" {
				query += fmt.Sprintf(`functionNode = g.V(serviceNode).out().hasLabel('package').has('name', '%s').out().hasLabel('function').has('name', '%s').next();
										 fromFunc.addEdge('compile_time_call', functionNode);`, path.ContainerPackage, path.To)
				sugarLogger.Debug(query)
				sugarLogger.Debugf("First case: %v\n", RunQuery(query))
				continue
			}
			// Check if selector is one of our packages by mapping it to localName.
			if _, ok := serviceparser.AllDeclaredPackages[selectorName]; ok {
				query += fmt.Sprintf(`functionNode = g.V(serviceNode).out().hasLabel('package').has('name', '%s').out().hasLabel('function').has('name', '%s').next();
										  fromFunc.addEdge('compile_time_call', functionNode);`, selectorName, path.To)
				sugarLogger.Debug(query)
				sugarLogger.Debugf("Second case: %v\n", RunQuery(query))
				continue
			}

			sugarLogger.Info(path.SelectorCallee)
			selectorParts := strings.Split(path.SelectorCallee, ",")
			// Else create a new function node and link to external dependency
			query += fmt.Sprintf(`functionNode = g.V().addV('function').property('name', '%s').next();
									 importNode = g.V(serviceNode).out().hasLabel('dependency').has('local_name', '%s');
 									 exists = importNode.hasNext();
									 if (exists) {
										importNode.next().addEdge('provides', functionNode);
 								     }
									 fromFunc.addEdge('compile_time_call', functionNode);`, strings.Join(selectorParts[0:len(selectorParts) - 1], ".") + "." + path.To, selectorParts[len(selectorParts) - 1])
			sugarLogger.Debug(query)
			sugarLogger.Debugf("Third case: %v\n", RunQuery(query))
		}
	}
}

func AddPackageFunctionNodesToGraph(serviceName string, serviceVersion string) {
	for pkg, functions := range serviceparser.AllPkgFunc[serviceName] {
		gremlinQuery := NewPackageNodeQuery(serviceName, serviceVersion, pkg)
		gremlinQuery += CreateFunctionNodes(functions)

		sugarLogger.Infof("Executing package node creation gremlin query for service: %s, package: %s\n", serviceName, pkg)
		gremlinResponse := RunQuery(gremlinQuery)
		sugarLogger.Debugf("%v\n%v\n", gremlinQuery, gremlinResponse)
	}
}


func AddRuntimePathsToGraph(clusterVersion, serviceName, serviceVersion string, runtimePaths []serviceparser.CodePath) {
	for _, runtimePath := range runtimePaths {
		// First find the service
		query := fmt.Sprintf(
			`serviceNode = g.V().has('cluster_version', '%s').out().has('name', '%s').has('version', '%s').next();`, clusterVersion, serviceName, serviceVersion)

		query += fmt.Sprintf(`g.V(serviceNode).out().hasLabel('package').has('name', '%s')`, runtimePath.ContainerPackage)
		// TODO
	}
}