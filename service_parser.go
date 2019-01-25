package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

var allPkgFunc = make(map[string][]string)

func visit(path string, f os.FileInfo, err error) error {
	// Do not visit git dir.
	if f.IsDir() && (f.Name() == ".git" || f.Name() == "vendor") {
		return filepath.SkipDir
	}
	// Our logic is not for files.
	if !f.IsDir() {
		return nil
	}

	fset := token.NewFileSet()
	log.Print("Inside: ", path)
	node, err := parser.ParseDir(fset,
		path,
		nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	for pkg, ast := range node {
		pkgFunctions := parseServiceAST(ast, fset)
		allPkgFunc[pkg] = pkgFunctions
	}
	return nil
}

func main() {
	root := os.Args[1]
	fmt.Println("Walking: ", root)
	err := filepath.Walk(root, visit)
	if err != nil {
		log.Fatal(err)
	}
	packageJSON, err := json.Marshal(allPkgFunc)
	if err != nil {
		log.Fatal(err)
	}
	_, fileName := filepath.Split(root)
	err = ioutil.WriteFile(fileName+".json", packageJSON, 0644)
	if err != nil {
		panic(err)
	}
}

func parseServiceAST(node ast.Node, fset *token.FileSet) []string {
	functions := make([]string, 1)
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
