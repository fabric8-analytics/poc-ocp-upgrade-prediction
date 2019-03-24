package serviceparser

import (
	"go/ast"
	"go/token"

	"sourcegraph.com/sourcegraph/go-diff/diff"
)

// Block represents the information about a basic block to be recorded in the analysis.
type Block struct {
	startByte token.Pos
	endByte   token.Pos
}

// File is a wrapper for the state of a file used in the parser.
type File struct {
	fset    *token.FileSet
	name    string // Name of file.
	astFile *ast.File
	blocks  []Block
	pkgName string
}

// ParseDiff parses a git commit diff set.
func ParseDiff(diffstr string) ([]*diff.FileDiff, error) {
	fdiff, err := diff.ParseMultiFileDiff([]byte(diffstr))
	if err != nil {
		return nil, err
	}
	return fdiff, nil
}
