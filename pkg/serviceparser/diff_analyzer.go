package serviceparser

import (
	"github.com/fabric8-analytics/poc-ocp-upgrade-prediction/pkg/utils"
	"path/filepath"
	"regexp"
	"strings"

	gdf "sourcegraph.com/sourcegraph/go-diff/diff"
)

// ParseDiff parses a git commit diff set.
func ParseDiff(diffstr string) ([]*gdf.FileDiff, error) {
	fdiff, err := gdf.ParseMultiFileDiff([]byte(diffstr))
	if err != nil {
		return nil, err
	}
	return fdiff, nil
}

func getAddedFunctions(diffContent string, filename string, functionsMatch []SimpleFunctionRepresentation) []SimpleFunctionRepresentation {
	funcAddedRe := regexp.MustCompile(`\+\s*func\s*(?P<structdefn>\(.*\))?\s*(?P<fname>[a-zA-Z0-9_]*)\(`)
	matches := funcAddedRe.FindAllStringSubmatch(diffContent, -1)
	//sugarLogger.Infof("%v %+v\n", diffContent, matches)
	for _, match := range matches {
		for i, k := range funcAddedRe.SubexpNames() {
			if k == "fname" {
				if len(match) < 3 || match[2] == "" {
					break
				}
				functionsMatch = append(functionsMatch, SimpleFunctionRepresentation{
					Fun: match[i],
					Pkg: filepath.Dir(filename),
					DeclFile: filename,
				})
			}
		}
	}
	return functionsMatch
}

func getDeletedFunctions(diffContent string, filename string, functionsMatch []SimpleFunctionRepresentation) []SimpleFunctionRepresentation {
	funcAddedRe := regexp.MustCompile(`\-\s*func\s*(?P<structdefn>\(.*\))?\s*(?P<fname>[a-zA-Z0-9_]*)\(`)
	matches := funcAddedRe.FindAllStringSubmatch(diffContent, -1)
	for _, match := range matches {
		for i, k := range funcAddedRe.SubexpNames() {
			if k == "fname" {
				if len(match) < 3 || match[2] == "" {
					break
				}
				functionsMatch = append(functionsMatch, SimpleFunctionRepresentation{
					Fun: match[i],
					Pkg: filepath.Dir(filename),
					DeclFile: filename,
				})
			}
		}
	}
	return functionsMatch
}

func getModifiedFunctions(funcsAdded []SimpleFunctionRepresentation, funcsDeleted []SimpleFunctionRepresentation, modifiedFuncs []SimpleFunctionRepresentation) []SimpleFunctionRepresentation {
	addedFuncsMap := make(map[SimpleFunctionRepresentation]bool)
	for _, frep := range funcsAdded {
		addedFuncsMap[frep] = true
	}
	for _, frep := range funcsDeleted {
		if addedFuncsMap[frep] == true {
			// TODO: Make this more efficient.
			alreadyPresent := 0
			for _, fun := range modifiedFuncs {
				if fun == frep {
					alreadyPresent = 1
					break
				}
			}
			if alreadyPresent == 0 {
				modifiedFuncs = append(modifiedFuncs, frep)
			}
		}
	}
	return modifiedFuncs
}

// Parse the section header for a function name.
func parseSectionHeader(sectionHeader string, filename string) SimpleFunctionRepresentation {
	funcModifiedRe := regexp.MustCompile(`func\s*(?P<structdefn>\(.*\))?\s*(?P<fname>[a-zA-Z0-9_]*)\(`)
	modifiedMap := utils.ReSubMatchMap(funcModifiedRe, sectionHeader)
	return SimpleFunctionRepresentation{Fun: modifiedMap["fname"], Pkg: filepath.Dir(filename), DeclFile: filename}
}

// GetTouchPointsOfPR is used to get the functions that are affected by a certain PR.
//(Go source code changes.)
func GetTouchPointsOfPR(allDiffs []*gdf.FileDiff, branchDetails []MetaRepo) *TouchPoints {
	funcsAdded := make([]SimpleFunctionRepresentation, 0)
	funcsDeleted := make([]SimpleFunctionRepresentation, 0)
	funcsChanged := make([]SimpleFunctionRepresentation, 0)

	for _, diff := range allDiffs {
		// Ignore all the test files
		if strings.Contains(diff.OrigName, "_test.go") || strings.Contains(diff.NewName, "_test.go") {
			continue
		}
		diffContent := ""
		for _, hunk := range diff.Hunks {
			diffContent += strings.Trim(string(hunk.Body), "\t\n")
			// Use the section header for modification was done somewhere within a function body.
			sectionFunc := parseSectionHeader(hunk.Section, removeAB(diff.NewName))
			if sectionFunc.Fun != "" {
				// TODO: Make this efficient by way of lookup.
				alreadyChanged := 0
				for _, fun := range funcsChanged {
					if fun == sectionFunc {
						alreadyChanged = 1
						break
					}
				}
				if alreadyChanged == 0 {
					funcsChanged = append(funcsChanged, sectionFunc)
				}
			}
		}
		// First get all the changes where the declaration itself was modified.
		funcsAdded = getAddedFunctions(diffContent, removeAB(diff.NewName), funcsAdded)
		funcsDeleted = getDeletedFunctions(diffContent, removeAB(diff.OrigName), funcsDeleted)
		funcsChanged = getModifiedFunctions(funcsAdded, funcsDeleted, funcsChanged)
	}

	// Check which functions are in the current master branch functions and also in the filesChanged
	// thing to know which diff was changed.
	funcsAdded, funcsDeleted = removeModifiedFromAddedDeleted(funcsAdded, funcsDeleted)

	return &TouchPoints{
		FunctionsAdded:   funcsAdded,
		FunctionsDeleted: funcsDeleted,
		FunctionsChanged: funcsChanged,
	}
}

func removeAB(filename string) string {
	if strings.HasPrefix(filename, "a/") {
		return strings.TrimPrefix(filename, "a/")
	}
	if strings.HasPrefix(filename, "b/") {
		return strings.TrimPrefix(filename, "b/")
	}
	return filename
}

func removeModifiedFromAddedDeleted(added []SimpleFunctionRepresentation, deleted []SimpleFunctionRepresentation) ([]SimpleFunctionRepresentation, []SimpleFunctionRepresentation) {
	found := make(map[SimpleFunctionRepresentation]int)
	for idx, fun := range added {
		found[fun] = idx
	}
	deletionQAdded := make([]int, 0)
	deletionQDeleted := make([]int, 0)
	for idx, fun := range deleted {
		if idxAdded, exists := found[fun]; exists {
			deletionQAdded = append(deletionQAdded, idxAdded)
			deletionQDeleted = append(deletionQDeleted, idx)
		}
	}
	added = removeFunctionsFromSlice(added, deletionQAdded)
	deleted = removeFunctionsFromSlice(deleted, deletionQDeleted)
	return added, deleted
}

func removeFunctionsFromSlice(haystack []SimpleFunctionRepresentation, idxs []int) []SimpleFunctionRepresentation {
	var trimmed []SimpleFunctionRepresentation
	toDelete := make(map[int]bool)
	for _, idx := range idxs {
		toDelete[idx] = true
	}
	for i, fn := range haystack {
		if _, exists := toDelete[i]; !exists {
			trimmed = append(trimmed, fn)
		}
	}
	return trimmed
}