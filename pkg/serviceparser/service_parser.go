package serviceparser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"sourcegraph.com/sourcegraph/go-diff/diff"
)

var logger, _ = zap.NewDevelopment()
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
		pkgs, err := parser.ParseDir(fset,
			path,
			nil, parser.ParseComments)
		if err != nil {
			sugarLogger.Fatal(err)
		}

		for pkg, pkgast := range pkgs {
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
	sugarLogger.Debugf("%v\n", ic)
	return ic
}

func parseServiceAST(pkgast *ast.Package, fset *token.FileSet, pkg string) ([]string, []ImportContainer) {
	var functions []string
	var imports []ImportContainer

	ast.Inspect(pkgast, func(n ast.Node) bool {

		// Find Functions
		switch fnOrImp := n.(type) {
		case *ast.FuncDecl:
			functions = append(functions, fnOrImp.Name.Name)
		case *ast.ImportSpec:
			imports = append(imports, parseImportNode(fnOrImp, pkg))
		}
		return true
	})

	var functionLit []string

	for _, fileast := range pkgast.Files {
		for _, decl := range fileast.Decls {
			if declBody, ok := decl.(*ast.GenDecl); ok {
				if declBody.Tok == token.VAR {
					for _, specBody := range declBody.Specs {
						valSpec, isvalSpec := specBody.(*ast.ValueSpec)
						if isvalSpec {
							sugarLogger.Debug("Found a valuespec.")
							if len(valSpec.Values) == 0 {
								continue
							}
							_, isFnLit := valSpec.Values[0].(*ast.FuncLit)
							if isFnLit {
								functionLit = append(functionLit, valSpec.Names[0].Name)
							}
						}
					}
				}
			}
		}
	}
	functions = append(functions, functionLit...)
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
