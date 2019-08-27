package gremlin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/fabric8-analytics/poc-ocp-upgrade-prediction/pkg/serviceparser"
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
		sugarLogger.Error(err)
		return make(map[string]interface{})
	}

	var result map[string]interface{}

	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		sugarLogger.Errorf("Failed to decode JSON: %v\n", err)
	}
	return result
}

// RunQuery runs the specified gremling query and returns its result unmarshaled.
func RunQueryUnMarshaled(query string) string {
	payload := map[string]interface{}{
		"gremlin": query,
	}
	payloadJSON, _ := json.Marshal(payload)
	response, err := http.Post(os.Getenv("GREMLIN_REST_URL"), "application/json", bytes.NewBuffer(payloadJSON))

	if err != nil {
		sugarLogger.Error(err)
		return ""
	}

	var buf bytes.Buffer
	_, err = buf.ReadFrom(response.Body)

	if err != nil {
		sugarLogger.Error(err)
	}
	return buf.String()
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
func CreateNewServiceVersionNode(clusterVersion, serviceName, version string) {
	query := fmt.Sprintf(`
		clusterVersion = g.V().has('vertex_label', 'clusterVersion').has('cluster_version', '%s').next();
		serviceVersion = g.addV('service_version').property('vertex_label', 'service_version').property('name', '%s').property('version', '%s').next();
		clusterVersion.addEdge('contains_service', serviceVersion);`, clusterVersion, serviceName, version)
	sugarLogger.Debug(query)
	sugarLogger.Debugf("%v\n", RunQuery(query))
}

// NewPackageNodeQuery creates a new package node and joins it using an edge
// to the parent service node.
func NewPackageNodeQuery(serviceName, serviceVersion, packagename string) string {
	query := fmt.Sprintf(`
	serviceVersion = g.V().has('vertex_label', 'service_version').has('name', '%s').has('version', '%s').next();
	packageNode = g.addV('package').property('vertex_label', 'package').property('name', '%s').next();
	serviceVersion.addEdge('contains_package', packageNode);`, serviceName, serviceVersion, packagename)
	return query
}

// CreateFunctionNodes adds function nodes to the graph and an edge between it and it's
// parent service and it's package
// DO NOT CALL NewPackageNodeQuery BEFORE YOU'VE ENTERED ALL THE NODES FOR A PACKAGE
func CreateFunctionNodes(functionNames []string) string {
	var fullQuery string
	for _, fn := range functionNames {
		query := fmt.Sprintf(`functionNode = g.addV('function').property('vertex_label', 'function').property('name', '%s').next();
							  packageNode.addEdge('has_fn', functionNode);`, fn)
		fullQuery += query
	}
	return fullQuery
}

// CreateClusterVerisonNode creates the top level cluster version node
// CALL THIS JUST ONCE PER RUN OF THIS SCRIPT, THAT IS HOW THIS CODE IS DESIGNED.
func CreateClusterVerisonNode(clusterVersion string) {
	query := fmt.Sprintf(`
		clusterVersion = g.addV('clusterVersion').property('vertex_label', 'clusterVersion').property('cluster_version', '%s').next()`, clusterVersion)
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
func CreateDependencyNodes(serviceName, serviceVersion string, ic []serviceparser.ImportContainer) {
	queryBase := fmt.Sprintf(
		`serviceNode = g.V().has('vertex_label', 'service_version').has('name', '%s').has('version', '%s').next();`, serviceName, serviceVersion)

	query := queryBase
	for idx, imported := range ic {
		query += fmt.Sprintf(`importNode = g.addV('dependency').property('vertex_label', 'dependency').property('local_name', '%s').property('importpath', '%s').next();
				  serviceNode.addEdge('depends_on', importNode);
				  packageNode = g.V().has('vertex_label', 'package').has('name', '%s').next();
				  importNode.addEdge('affects_package', packageNode);`, imported.LocalName, imported.ImportPath, imported.DependentPkg)

		// Running this query in batches.
		if idx%30 == 0 {
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

// AddPackageFunctionNodesToGraph as advertised, adds a package node and its corresponding functions to the graph.
func AddPackageFunctionNodesToGraph(serviceName string, serviceVersion string, components *serviceparser.ServiceComponents) {
	for pkg, functions := range components.AllPkgFunc {
		gremlinQuery := NewPackageNodeQuery(serviceName, serviceVersion, pkg)
		gremlinQuery += CreateFunctionNodes(functions)

		sugarLogger.Infof("Executing package node creation gremlin query for service: %s, package: %s", serviceName, pkg)
		gremlinResponse := RunQuery(gremlinQuery)
		sugarLogger.Debugf("%v\n%v\n", gremlinQuery, gremlinResponse)
	}
}

// CreateCompileTimePaths creates compile time paths from the callgraph output.
func CreateCompileTimePaths(edges []serviceparser.CompileEdge, serviceName, serviceVersion string) {
	buffer := 1000
	queryString := ""
	for _, edge := range edges {
		callerFn := edge.Caller.Name()
		callerPkg := fmt.Sprintf("%v", edge.Caller.Package())
		callerPkg = strings.TrimPrefix(callerPkg, "package ")

		// Only consider itself and kubernetes for now.

		calleeFn := edge.Callee.Name()
		calleePkg := fmt.Sprintf("%v", edge.Callee.Package())
		calleePkg = strings.TrimPrefix(calleePkg, "package ")

		sugarLogger.Debugf("%v\n", calleePkg)

		callerPkg = sanitize(strings.Trim(callerPkg, "()*"))
		calleePkg = sanitize(strings.Trim(calleePkg, "()*"))

		var serviceNodeFrom, serviceNodeTo string
		if strings.HasPrefix(callerPkg, "kubernetes") {
			serviceNodeFrom = "hyperkube"
			callerPkg = strings.TrimPrefix(callerPkg, "vendor/k8s.io/kubernetes/")
		} else {
			serviceNodeFrom = serviceName
		}
		if strings.HasPrefix(calleePkg, "kubernetes") {
			serviceNodeTo = "hyperkube"
			calleePkg = strings.TrimPrefix(calleePkg, "vendor/k8s.io/kubernetes/")
		} else {
			serviceNodeTo = serviceName
		}

		serviceFinder := fmt.Sprintf(`serviceNodeFrom = g.V().has('vertex_label', 'service_version').has('name', '%s').has('version', '%s').next();
			serviceNodeTo = g.V().has('vertex_label', 'service_version').has('name', '%s').has('version', '%s').next();`,
			serviceNodeFrom, serviceVersion, serviceNodeTo, serviceVersion)

		gremlin := fmt.Sprintf(`from = g.V(serviceNodeFrom).out().has('vertex_label', 'package').has('name', '%s').out().has('vertex_label', 'function').has('name', '%s');
			to = g.V(serviceNodeTo).out().has('vertex_label', 'package').has('name', '%s').out().has('vertex_label', 'function').has('name', '%s');
			if (from.hasNext()) {
				if (to.hasNext()) {
					fromNode = from.Next()
					fromNode.addEdge('compile_time_call', to.next()).property('edge_label', 'compile_time_call');
				}
			}	
		`, callerPkg, callerFn, calleePkg, calleeFn)
		if len(queryString)+len(gremlin)+len(serviceFinder) < buffer {
			queryString += serviceFinder + gremlin
		} else {
			response := RunQuery(queryString)
			sugarLogger.Infof("%v\n", queryString)
			sugarLogger.Infof("Got response: %v from gremlin", response)
			queryString = serviceFinder + gremlin
		}
	}
	if queryString != "" {
		response := RunQuery(queryString)
		sugarLogger.Infof("Got response: %v from gremlin", response)
		queryString = ""
	}
}

func sanitize(s string) string {
	s = strings.TrimPrefix(s, "github.com/openshift/origin/")
	s = strings.TrimPrefix(s, "github.com/openshift/origin/vendor/k8s.io/")
	return s
}

func GetPRConfidenceScore(points *serviceparser.TouchPoints) PrConfidence {
	countCompileTime := int64(0)
	countRunTime := int64(0)
	var confScore int64
	for _, point := range points.Flatten() {
		q := fmt.Sprintf(`g.V().has('vertex_label', 'package').has('name', '%s').out().has('vertex_label', 'function').has('name', '%s').in().has('edge_label', 'compile_time_path').count();`, point.Pkg, point.Fun)
		response := RunQueryUnMarshaled(q)
		thisFnScore := gjson.Get(response, "result.data").Array()[0].Int()
		countCompileTime += thisFnScore
		q = fmt.Sprintf(`g.V().has('vertex_label', 'package').has('name', '%s').out().has('vertex_label', 'function').has('name', '%s').in().has('edge_label', 'run_time_path').count();`, point.Pkg, point.Fun)
		response = RunQueryUnMarshaled(q)
		thisRunScore := gjson.Get(response, "result.data").Array()[0].Int()
		countRunTime += thisRunScore
		sugarLogger.Infof("Score for %v.%v: %d\n", point.Pkg, point.Fun, thisFnScore)
	}
	if countCompileTime > 0 {
		confScore = countRunTime / countCompileTime
	} else {
		confScore = -1
	}
	conf := PrConfidence{
		ConfidenceScore: confScore,
	}
	return conf
}

func GetCompileTimePathsAffectedByPR(points *serviceparser.TouchPoints) []string {
	var compilePaths []string
	for _, point := range points.Flatten() {
		curPaths := fmt.Sprintf(`g.V().has('vertex_label', 'package').has('name', '%s').out().has('vertex_label', 'function').has('name', '%s').inE().has('edge_label', 'compile_time_call').outV().path()`, point.Pkg, point.Fun)
		response := RunQueryUnMarshaled(curPaths)
		compilePaths = append(compilePaths, gjson.Get(response, "result.data").String())
	}
	return compilePaths
}