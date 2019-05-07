package serviceparser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

var logger, _ = zap.NewDevelopment()
var sugarLogger = logger.Sugar()

// ParseService parses a service and dumps all its functions to a JSON
func (components *ServiceComponents) ParseService(serviceName string, root string) {
	sugarLogger.Debugf("Parsing service: %v\n", root)

	err := filepath.Walk(root, func(path string, f os.FileInfo, err error) error {
		// Do not visit git dir, vendor, generated, bindata etc.
		if f.IsDir() && (f.Name() == ".git" || f.Name() == "vendor" || strings.Contains(f.Name(), "generated") || strings.Contains(f.Name(), "third_party") || strings.Contains(f.Name(), "test") || (strings.Contains(f.Name(), "bindata"))) {
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
			pkgDir, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			pkg = filepath.Join(filepath.Dir(pkgDir), pkg)
			pkgFiles := pkgast.Files
			for filename := range pkgFiles {
				// I think this will always be unique so not doing on a per-service basis.
				components.FilePackageMap[filename] = pkgast.Name
			}
			components.AllDeclaredPackages[pkg] = true
			pkgFunctions, pkgImports := parseServiceAST(pkgast, fset, pkg)
			components.AllPkgFunc[pkg] = pkgFunctions
			components.AllPkgImports[pkg] = pkgImports
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
			if fnOrImp.Name.Name != "" {
				functions = append(functions, fnOrImp.Name.Name)
			}
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
							if len(valSpec.Values) == 0 {
								continue
							}
							_, isFnLit := valSpec.Values[0].(*ast.FuncLit)
							if isFnLit && valSpec.Names[0].Name != "" {
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
