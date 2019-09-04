package gremlin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/fabric8-analytics/poc-ocp-upgrade-prediction/pkg/serviceparser"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var logger, _ = zap.NewProduction()
var sugarLogger = logger.Sugar()

/*
// RunQuery runs the specified gremling query and returns its result.
func RunQuery(query string) map[string]interface{} {
	payload := map[string]interface{}{
		"gremlin": query,
	}
	payloadJSON, _ := json.Marshal(payload)
	response, err := http.Post(os.Getenv("GREMLIN_REST_URL"), "application/json",  bytes.NewBuffer(payloadJSON))
	//sugarLogger.Infof("Response : %s\n", response)

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
*/

var timeout = time.Duration(120 * time.Second)
var client = http.Client{
	Timeout: timeout,
}

func RunQuery(query string) map[string]interface{} {

	var result map[string]interface{}

	httpcontentlengthstring := "1073741824"
	httpcontentlengthstring, _ = os.LookupEnv("GREMLIN_HTTP_CONTENT_LENGTH")

	payload := map[string]interface{}{
		"gremlin": query,
	}
	payloadJSON, _ := json.Marshal(payload)

	fmt.Println(query)
	if os.Getenv("GREMLIN_REST_URL") != "" {
		request, err := http.NewRequest("POST", os.Getenv("GREMLIN_REST_URL"), bytes.NewBuffer(payloadJSON))
		request.Header.Set("content-type", "application/json")
		request.Header.Set("content-length", httpcontentlengthstring)

		if err != nil {
			fmt.Print(err)
			return result
		}
		response, err := client.Do(request)
		if err != nil {
			fmt.Print(err)
			return result
		} else {
			err = json.NewDecoder(response.Body).Decode(&result)
			if err != nil {
				sugarLogger.Errorf("Failed to decode JSON: %v\n", err)
			}
			defer response.Body.Close()
		}
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
		clusterVersion = g.V().has('vertex_label', 'clusterVersion').has('cluster_version', '%s');
		serviceVersion = g.addV('service_version').property('vertex_label', 'service_version').property('name', '%s').property('version', '%s');
		if (clusterVersion.hasNext()) { if (serviceVersion.hasNext()) { clusterVersion.next().addEdge('contains_service', serviceVersion.next()) } };`, clusterVersion, serviceName, version)
	RunQuery(query)
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
	response := RunQuery(query)
	sugarLogger.Debugf(" the clusterversion response %s ", response)
}

// RunGroovyScript takes the path to a groovy script and runs it at the Gremlin console.
func RunGroovyScript(scriptPath string) {
	scriptContent := ReadFile(scriptPath)
	gremlinResponse := RunQuery(scriptContent)
	sugarLogger.Info(gremlinResponse)
}

func CreateDependencyNodes(serviceName, serviceVersion string, ic []serviceparser.ImportContainer) {
	var batchcount int
	batchsize := 20
	queryBase := fmt.Sprintf(
		`serviceNode = g.V().has('vertex_label', 'service_version').has('name', '%s').has('version', '%s');`, serviceName, serviceVersion)
	query := queryBase

	for _, imported := range ic {
		query += fmt.Sprintf(`importNode = g.addV('dependency').property('vertex_label', 'dependency').property('local_name', '%s').property('importpath', '%s');
				  if (serviceNode.hasNext()) {
					if (importNode.hasNext()) {
						copyimportNode = importNode.next();
					  	serviceNode.next()addEdge('depends_on', copyimportNode);
					  	packageNode = g.V().has('vertex_label', 'package').has('name', '%s');
					  	if (packageNode.hasNext()) {
							copyimportNode.addEdge('affects_package', packageNode.next());
					  	}
					}
				} 
				`, imported.LocalName, imported.ImportPath, imported.DependentPkg)

		batchsizestring, _ := os.LookupEnv("BATCH_SIZE_CREATE_DEPENDENCY_NODES")
		batchsize, _ = strconv.Atoi(batchsizestring)
		if batchcount > batchsize {
			RunQuery(query)
			batchcount = 0
			query = queryBase
		} else {
			batchcount = batchcount + 1
		}
	}
	// Any remainders
	if query != queryBase {
		RunQuery(query)
	}
}

// AddPackageFunctionNodesToGraph as advertised, adds a package node and its corresponding functions to the graph.
func AddPackageFunctionNodesToGraph(serviceVersion string, components *serviceparser.ServiceComponents) {
	gremlinQuery := ""
	for pkg, functions := range components.AllPkgFunc {

		serviceName := getServiceName(pkg)
		gremlinQuery += NewPackageNodeQuery(serviceName, serviceVersion, pkg)
		gremlinQuery += CreateFunctionNodes(functions)
		RunQuery(gremlinQuery)
		gremlinQuery = ""
	}
}

func getServiceName(packageName string) string {
	serviceName := "hypershift"
	if strings.Index(packageName, "kubernetes") != -1 {
		serviceName = "hyperkube"
	}
	return serviceName
}

// CreateCompileTimePaths creates compile time paths from the callgraph output.
func CreateCompileTimePaths(edges []serviceparser.CompileEdge, serviceName, serviceVersion string) {
	queryString := ""
	batchcount := 0
	batchsize := 50

	for _, edge := range edges {
		callerFn := edge.Caller.Name()
		callerPkg := fmt.Sprintf("%v", edge.Caller.Package())
		callerPkg = strings.TrimPrefix(callerPkg, "package ")

		// Only consider itself and kubernetes for now.

		calleeFn := edge.Callee.Name()
		calleePkg := fmt.Sprintf("%v", edge.Callee.Package())
		calleePkg = strings.TrimPrefix(calleePkg, "package ")
		callerPkg = sanitize(strings.Trim(callerPkg, "()*"))
		calleePkg = sanitize(strings.Trim(calleePkg, "()*"))

		batchsizestring, _ := os.LookupEnv("BATCH_SIZE_CREATE_COMPILE_TIME_PATHS")
		batchsize, _ = strconv.Atoi(batchsizestring)
		if batchcount < batchsize {
			queryString += fmt.Sprintf(`from = g.V().has('vertex_label', 'package').has('name', '%s').out().has('vertex_label', 'function').has('name', '%s');to = g.V().has('vertex_label', 'package').has('name', '%s').out().has('vertex_label', 'function').has('name', '%s'); if (from.hasNext()) { if (to.hasNext()) { from.next().addEdge('compile_time_call', to.next()).property('edge_label', 'compile_time_call'); }};		
			`, callerPkg, callerFn, calleePkg, calleeFn)
			batchcount = batchcount + 1
		} else {
			RunQuery(queryString)
			batchcount = 0
			queryString = ""
		}
	}
	if queryString != "" {
		RunQuery(queryString)
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
	var confScore float64
	for _, point := range points.Flatten() {
		q := fmt.Sprintf(`g.V().has('vertex_label', 'package').has('name', '%s').out().has('vertex_label', 'function').has('name', '%s').bothE().has('edge_label', 'compile_time_call').dedup().count();`, point.Pkg, point.Fun)
		response := RunQueryUnMarshaled(q)
		thisFnScore := gjson.Get(response, "result.data").Array()[0].Int()
		countCompileTime += thisFnScore
		q = fmt.Sprintf(`g.V().has('vertex_label', 'package').has('name', '%s').out().has('vertex_label', 'function').has('name', '%s').bothE().has('edge_label', 'run_time_call').dedup().count();`, point.Pkg, point.Fun)
		response = RunQueryUnMarshaled(q)
		thisRunScore := gjson.Get(response, "result.data").Array()[0].Int()
		countRunTime += thisRunScore
		sugarLogger.Infof("Score for %v.%v: %d\n", point.Pkg, point.Fun, thisFnScore)
		sugarLogger.Infof("Score for %v.%v: %d\n", point.Pkg, point.Fun, thisRunScore)

	}
	if countCompileTime > 0 {
		confScore = float64(countRunTime+10) / float64(countCompileTime)
	} else {
		confScore = -1
	}
	conf := PrConfidence{
		ConfidenceScore: confScore,
	}
	return conf
}

func GetCompileTimePathsAffectedByPR(points *serviceparser.TouchPoints) []map[string]interface{} {
	var compilePaths []map[string]interface{}
	for _, point := range points.Flatten() {
		pathsIn := fmt.Sprintf(`g.V().has('vertex_label', 'package').has('name', '%s').out().has('vertex_label', 'function').has('name', '%s').repeat(inE('compile_time_call').outV().dedup()).until(inE('compile_time_call').count().is(0)).path()`, point.Pkg, point.Fun)
		sugarLogger.Infof("%v %v\n", "Running query: ", pathsIn)
		pathsOut := fmt.Sprintf(`g.V().has('vertex_label', 'package').has('name', '%s').out().has('vertex_label', 'function').has('name', '%s').repeat(outE('compile_time_call').inV().dedup()).until(outE('compile_time_call').count().is(0)).path()`, point.Pkg, point.Fun)
		responseIn := RunQueryUnMarshaled(pathsIn)
		responseOut := RunQueryUnMarshaled(pathsOut)
		respArr := gjson.Get(responseIn, "result.data").Array()
		respArr = append(respArr, gjson.Get(responseOut, "result.data").Array()...)
		if len(respArr) == 0 {
			continue
		}
		for _, rep := range respArr {
			repMap := rep.Value().(map[string]interface{})
			compilePaths = append(compilePaths, repMap)
		}
	}
	return compilePaths
}
