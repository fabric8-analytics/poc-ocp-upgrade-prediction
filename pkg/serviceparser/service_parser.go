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

type importContainer struct {
	LocalName    string `json:"local_name"`
	ImportPath   string `json:"import_path"`
	DependentPkg string `json:"dependent_pkg"`
}

// AllPkgFunc variable contains all the services mapped to their corresponding functions.
var AllPkgFunc = make(map[string]map[string][]string)

// AllPkgImports contains all the external dependencies.
var AllPkgImports = make(map[string]interface{})

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
			pkgFunctions, pkgImports := parseServiceAST(ast, fset, pkg)
			AllPkgFunc[serviceName][pkg] = pkgFunctions
			AllPkgImports[serviceName] = pkgImports
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

func parseImportNode(imp *ast.ImportSpec, pkg string) *importContainer {
	ic := importContainer{
		LocalName:    imp.Name.Name,
		ImportPath:   imp.Path.Value,
		DependentPkg: pkg,
	}
	return &ic
}

func parseServiceAST(node ast.Node, fset *token.FileSet, pkg string) ([]string, []*importContainer) {
	var functions []string
	var imports []*importContainer
	ast.Inspect(node, func(n ast.Node) bool {

		// Find Functions
		switch fnOrImp := n.(type) {
		case *ast.FuncDecl:
			functions = append(functions, fnOrImp.Name.Name)
		case *ast.ImportSpec:
			// TODO: add the logic to get imports
			imports = append(imports, parseImportNode(fnOrImp, pkg))
		}
		return true
	})

	return functions, imports
}
