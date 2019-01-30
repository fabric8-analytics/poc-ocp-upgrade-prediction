package serviceparser

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

// AllPkgFunc variable contains all the services mapped to their corresponding functions.
var AllPkgFunc = make(map[string]map[string][]string)

// ParseService parses a service and dumps all its functions to a JSON
func ParseService(serviceName string, root string, destdir string) {
	log.Print("Walking: ", root)
	AllPkgFunc[serviceName] = make(map[string][]string)
	err := filepath.Walk(root, func(path string, f os.FileInfo, err error) error {
		// Do not visit git dir.
		if f.IsDir() && (f.Name() == ".git" || f.Name() == "vendor") {
			return filepath.SkipDir
		}
		// Our logic is not for files.
		if !f.IsDir() {
			return nil
		}

		fset := token.NewFileSet()
		node, err := parser.ParseDir(fset,
			path,
			nil, parser.ParseComments)
		if err != nil {
			log.Fatal(err)
		}

		for pkg, ast := range node {
			pkgFunctions := parseServiceAST(ast, fset)
			AllPkgFunc[serviceName][pkg] = pkgFunctions
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	packageJSON, err := json.Marshal(AllPkgFunc[serviceName])
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile(filepath.Join(destdir, serviceName+".json"), packageJSON, 0644)
	if err != nil {
		panic(err)
	}
}

func parseServiceAST(node ast.Node, fset *token.FileSet) []string {
	var functions []string
	ast.Inspect(node, func(n ast.Node) bool {

		// Find Functions
		fn, ok := n.(*ast.FuncDecl)
		if ok {
			functions = append(functions, fn.Name.Name)
			return true
		}
		return true
	})

	return functions
}
