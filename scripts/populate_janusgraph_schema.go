package main

import (
	"path/filepath"

	"github.com/fabric8-analytics/poc-ocp-upgrade-prediction/pkg/gremlin"
)

// Runs the groovy script that creates indices.
func main() {
	gremlin.RunGroovyScript(filepath.Join("gremlin", "schema.groovy"))
}
