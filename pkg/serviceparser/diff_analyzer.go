package serviceparser

import (
	"go/ast"
	"go/token"

	gdf "sourcegraph.com/sourcegraph/go-diff/diff"
)

// Block represents the information about a basic block to be recorded in the analysis.
type Block struct {
	startByte token.Pos
	endByte   token.Pos
}

// MetaRepo contains all the fields that are required to clone something.
type MetaRepo struct {
	Branch   string
	Revision string
	URL      string
}

// File is a wrapper for the state of a file used in the parser.
type File struct {
	fset    *token.FileSet
	name    string // Name of file.
	astFile *ast.File
	blocks  []Block
	pkgName string
}

// Struct Touchpoints defines all the touchpoints of a PR
type TouchPoints struct {
}

// ParseDiff parses a git commit diff set.
func ParseDiff(diffstr string) ([]*gdf.FileDiff, error) {
	fdiff, err := gdf.ParseMultiFileDiff([]byte(diffstr))
	if err != nil {
		return nil, err
	}
	return fdiff, nil
}

func GetTouchPointsOfPR(allHunks [][]*gdf.Hunk, branchDetails []MetaRepo) TouchPoints {
	// TODO
	return TouchPoints{}
}
