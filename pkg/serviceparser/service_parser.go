package serviceparser

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"sourcegraph.com/sourcegraph/go-diff/diff"
)

var logger, _ = zap.NewProduction()
var sugarLogger = logger.Sugar()

// ImportContainer is a type to contain the import declaration, similar to *ast.ImportSpec
type ImportContainer struct {
	LocalName    string `json:"local_name"`
	ImportPath   string `json:"import_path"`
	DependentPkg string `json:"dependent_pkg"`
}

// AllPkgFunc variable contains all the services mapped to their corresponding functions.
var AllPkgFunc = make(map[string]map[string][]string)

// AllPkgImports contains all the external dependencies.
var AllPkgImports = make(map[string]map[string]interface{})

// AllCompileTimeFlows contains all the function calls identified at compile time.
var AllCompileTimeFlows = make(map[string]map[string]interface{})

// AllDeclaredPackages contains all the packages declared in this service.
var AllDeclaredPackages map[string]bool

// FilePackageMap is a mapping that tell you which package is in which file.
var FilePackageMap map[string]string

// ParseService parses a service and dumps all its functions to a JSON
func ParseService(serviceName string, root string, destdir string) {
	sugarLogger.Info("Walking: ", root)
	AllDeclaredPackages = make(map[string]bool)
	AllPkgFunc[serviceName] = make(map[string][]string)
	AllPkgImports[serviceName] = make(map[string]interface{})
	AllCompileTimeFlows[serviceName] = make(map[string]interface{})
	FilePackageMap = make(map[string]string)
	err := filepath.Walk(root, func(path string, f os.FileInfo, err error) error {
		// Do not visit git dir.
		if f.IsDir() && (f.Name() == ".git" || f.Name() == "vendor" || strings.Contains(f.Name(), "generated")) {
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
			sugarLogger.Fatal(err)
		}

		for pkg, pkgast := range node {
			pkgFiles := pkgast.Files
			for filename := range pkgFiles {
				// I think this will always be unique so not doing on a per-service basis.
				FilePackageMap[filename] = pkgast.Name
			}
			AllDeclaredPackages[pkg] = true
			pkgFunctions, pkgImports := parseServiceAST(pkgast, fset, pkg)
			AllCompileTimeFlows[serviceName][pkg] = ParseTreePaths(pkg, pkgast)
			AllPkgFunc[serviceName][pkg] = pkgFunctions
			AllPkgImports[serviceName][pkg] = pkgImports
		}
		return nil
	})
	if err != nil {
		sugarLogger.Fatal(err)
	}
	packageJSON, err := json.Marshal(AllPkgFunc[serviceName])
	if err != nil {
		sugarLogger.Fatal(err)
	}
	err = ioutil.WriteFile(filepath.Join(destdir, serviceName+".json"), packageJSON, 0644)
	if err != nil {
		panic(err)
	}
}

func parseImportNode(imp *ast.ImportSpec, pkg string) ImportContainer {
	var impName string
	if imp.Name != nil {
		impName = imp.Name.Name
	} else {
		_, impName = filepath.Split(imp.Path.Value)
	}
	ic := ImportContainer{
		LocalName:    strings.Trim(impName, "\""),
		ImportPath:   strings.Trim(imp.Path.Value, "\""),
		DependentPkg: pkg,
	}
	sugarLogger.Infof("%v\n", ic)
	return ic
}

func parseServiceAST(node ast.Node, fset *token.FileSet, pkg string) ([]string, []ImportContainer) {
	var functions []string
	var imports []ImportContainer
	ast.Inspect(node, func(n ast.Node) bool {

		// Find Functions
		switch fnOrImp := n.(type) {
		case *ast.FuncDecl:
			functions = append(functions, fnOrImp.Name.Name)
		case *ast.ImportSpec:
			imports = append(imports, parseImportNode(fnOrImp, pkg))
		}
		return true
	})

	return functions, imports
}

// ParseDiff parses a git commit diff set.
func ParseDiff(diffstr string) ([]*diff.FileDiff, error) {
	fdiff, err := diff.ParseMultiFileDiff([]byte(diffstr))
	if err != nil {
		return nil, err
	}
	return fdiff, nil
}
