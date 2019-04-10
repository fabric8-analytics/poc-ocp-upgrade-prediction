package serviceparser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"regexp"
	"strings"

	gdf "sourcegraph.com/sourcegraph/go-diff/diff"
)

// MetaRepo contains all the fields that are required to clone something.
type MetaRepo struct {
	Branch    string
	Revision  string
	URL       string
	LocalPath string
}

// Struct Touchpoints defines all the touchpoints of a PR
type TouchPoints struct {
	functionsChanged map[string][]string
	functionsDeleted map[string][]string
	functionsAdded   map[string][]string
}

// ParseDiff parses a git commit diff set.
func ParseDiff(diffstr string) ([]*gdf.FileDiff, error) {
	fdiff, err := gdf.ParseMultiFileDiff([]byte(diffstr))
	if err != nil {
		return nil, err
	}
	return fdiff, nil
}

// GetTouchPointsOfPR is used to get the functions that are affected by a certain PR.
//(Go source code changes.)
func GetTouchPointsOfPR(allDiffs []*gdf.FileDiff, branchDetails []MetaRepo) *TouchPoints {
	var filesChanged []*gdf.FileDiff
	var filesAdded []*gdf.FileDiff
	diffMap := make(map[string]*gdf.FileDiff)
	astFilesAdded := make(map[string]*ast.File)
	astFilesChanged := make(map[string]*ast.File)

	for _, diff := range allDiffs {
		if diff.OrigName == "/dev/null" {
			filesAdded = append(filesAdded, diff)
			fileAst, err := parser.ParseFile(token.NewFileSet(),
				filepath.Join(branchDetails[0].LocalPath, diff.NewName), nil, parser.ParseComments)
			if err != nil {
				sugarLogger.Errorf("%v\n", err)
			}
			astFilesAdded[diff.NewName] = fileAst
		} else {
			filesChanged = append(filesChanged, diff)
			fileAst, err := parser.ParseFile(token.NewFileSet(),
				filepath.Join(branchDetails[0].LocalPath, diff.OrigName), nil, parser.ParseComments)
			if err != nil {
				sugarLogger.Errorf("%v\n", err)
			}
			astFilesChanged[diff.NewName] = fileAst
		}
	}

	// Now get all the changed functions from the ASTs
	funcsTouched := make(map[string][]string)
	funcsDeleted := make(map[string][]string)
	funcsAdded := make(map[string][]string)

	// All functions in newly added files will be a part of this.
	for filename, fileAst := range astFilesAdded {
		ast.Inspect(fileAst, func(n ast.Node) bool {
			if fdecl, isfdecl := n.(*ast.FuncDecl); isfdecl {
				funcsTouched[filename] = append(funcsTouched[filename], fdecl.Name.Name)
			}
			return true
		})
	}

	// The diffs are a bit complicated- if it's in a (+) line it's a touchpoint, else if it's only in
	// a (-) line it's actually been removed.
	for filename, fileAst := range astFilesChanged {
		ast.Inspect(fileAst, func(n ast.Node) bool {

			if fdecl, isfdecl := n.(*ast.FuncDecl); isfdecl {
				for _, hunk := range diffMap[filename].Hunks {
					linesAdded := ""
					linesDeleted := ""

					linesChanged := strings.Split(hunk.String(), "\n")
					for _, change := range linesChanged {
						if strings.HasPrefix(change, "+") {
							linesAdded += change
						} else {
							linesDeleted += change
						}
					}
					funcDeclre := fmt.Sprintf("func.*%s\\s*\\(.*", fdecl.Name.Name)

					isDeleted, err := regexp.MatchString(funcDeclre, linesDeleted)
					if err != nil {
						sugarLogger.Errorf("%v\n", err)
					}

					isAdded, err := regexp.MatchString(funcDeclre, linesAdded)
					if err != nil {
						sugarLogger.Errorf("%v\n", err)
					}

					if isDeleted && !isAdded {
						funcsDeleted[filename] = append(funcsTouched[filename], fdecl.Name.Name)
					} else if isDeleted && isAdded {
						funcsTouched[filename] = append(funcsTouched[filename], fdecl.Name.Name)
					} else {
						funcsAdded[filename] = append(funcsTouched[filename], fdecl.Name.Name)
					}
				}
			}

			return true
		})
	}

	// Check which functions are in the current master branch functions and also in the filesChanged
	// thing to know which diff was changed.
	return &TouchPoints{
		functionsAdded:   funcsAdded,
		functionsDeleted: funcsDeleted,
		functionsChanged: funcsTouched,
	}
}
