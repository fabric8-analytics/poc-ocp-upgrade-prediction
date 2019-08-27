package serviceparser

// ImportContainer is a type to contain the import declaration, similar to *ast.ImportSpec
type ImportContainer struct {
	LocalName    string `json:"local_name"`
	ImportPath   string `json:"import_path"`
	DependentPkg string `json:"dependent_pkg"`
}

// ServiceComponents is the result of calling ParseService on a repository(folder).
type ServiceComponents struct {
	// The name of this service
	Servicename string `json:"servicename,omitempty"`
	// AllPkgFunc variable contains all the packages of the service and their corresponding functions.
	AllPkgFunc map[string][]string `json:"all_pkg_func,omitempty"`
	// AllPkgImports contains all the external dependencies, again mapped to packages.
	AllPkgImports map[string]interface{} `json:"all_pkg_imports,omitempty"`
	// TODO: Add a field for compile time flows for all the services here?
	// AllDeclaredPackages contains all the packages declared in this service, in a map because there's no set in this language.
	AllDeclaredPackages map[string]bool `json:"all_declared_packages,omitempty"`
	// FilePackageMap is a mapping that tell you which package is in which file.
	// Contains absolute path of file(relative to root of current package package ex. pkg/a/b.go)
	FilePackageMap map[string]string `json:"file_package_map,omitempty"`
}

// Creates a new service components structure for serviceparser's use.
func NewServiceComponents(servicename string) *ServiceComponents {
	return &ServiceComponents{
		Servicename:         servicename,
		AllPkgFunc:          make(map[string][]string),
		AllPkgImports:       make(map[string]interface{}),
		AllDeclaredPackages: make(map[string]bool),
		FilePackageMap:      make(map[string]string),
	}
}

// A barebones function representation
type SimpleFunctionRepresentation struct {
	Fun      string `json:"fun"`
	Pkg      string `json:"pkg"`
	DeclFile string `json:"decl_file"`
}

// MetaRepo contains all the fields that are required to clone something.
type MetaRepo struct {
	Branch    string `json:"branch"`
	Revision  string `json:"revision"`
	URL       string `json:"url"`
	LocalPath string `json:"local_path"`
}

// Struct Touchpoints defines all the touchpoints of a PR
type TouchPoints struct {
	FunctionsChanged []SimpleFunctionRepresentation `json:"functions_changed"`
	FunctionsDeleted []SimpleFunctionRepresentation `json:"functions_deleted"`
	FunctionsAdded   []SimpleFunctionRepresentation `json:"functions_added"`
}

func (t *TouchPoints) Flatten() []SimpleFunctionRepresentation {
	var retval []SimpleFunctionRepresentation
	retval = append(retval, t.FunctionsAdded...)
	retval = append(retval, t.FunctionsChanged...)
	retval = append(retval, t.FunctionsDeleted...)
	return retval
}