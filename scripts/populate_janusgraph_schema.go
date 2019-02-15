package scripts

import (
	"path/filepath"

	"../pkg/gremlin"
)

// Runs the groovy script that creates indices.
func main() {
	gremlin.RunGroovyScript(filepath.Join("gremlin", "schema.groovy"))
}
