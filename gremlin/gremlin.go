package gremlin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

// RunQuery runs the specified gremling query and returns its result.
func RunQuery(query string) map[string]interface{} {
	payload := map[string]interface{}{
		"gremlin": query,
	}
	payloadJSON, _ := json.Marshal(payload)
	response, err := http.Post(os.Getenv("GREMLIN_REST_URL"), "application/json", bytes.NewBuffer(payloadJSON))

	if err != nil {
		log.Fatal(err)
	}

	var result map[string]interface{}

	json.NewDecoder(response.Body).Decode(&result)
	return result
}

// ReadJSON reads the contents of a JSON and returns it as a map[string]interface{}
func ReadJSON(jsonFilepath string) string {
	b, err := ioutil.ReadFile(jsonFilepath) // just pass the file name
	if err != nil {
		log.Fatal(err)
	}
	return string(b)
}

//ReadFile reads the contents of a text file and return it as a string
func ReadFile(filepath string) string {
	b, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Fatal(err)
	}
	return string(b)
}

// CreateNewServiceNode creates a new service node for a codebase
func CreateNewServiceNode(serviceName string, version string) {
	query := fmt.Sprintf(`
		graph.addVertex(label,'service',
						'name', '%s',
						'version', '%s')
					  `, serviceName, version)
	gremlinResponse := RunQuery(query)
	log.Print(gremlinResponse)
}

// CreateNewFunctionNode adds a new function node to the graph and an edge between it and it's
// parent service
func CreateNewFunctionNode(functionName string) {
}

// CreateClusterVerisonNode creates the top level cluster version node
func CreateClusterVerisonNode(clusterVersion string) {
	query := fmt.Sprintf(`
		graph.addVertex(label, 'clusterVersion',
						'cluter_version', '%s')
	`, clusterVersion)
	gremlinResponse := RunQuery(query)
	log.Print(gremlinResponse)
}

// RunGroovyScript takes the path to a groovy script and runs it at the Gremlin console.
func RunGroovyScript(scriptPath string) {
	scriptContent := ReadFile(scriptPath)
	gremlinResponse := RunQuery(scriptContent)
	log.Print(gremlinResponse)
}
