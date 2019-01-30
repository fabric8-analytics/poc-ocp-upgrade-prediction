package main

import (
	"path/filepath"

	"./gremlin"
)

// Runs the groovy script that creates indices.
func main() {
	gremlin.RunGroovyScript(filepath.Join("gremlin", "schema.groovy"))
}
