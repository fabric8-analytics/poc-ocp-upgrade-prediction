// Original Copyright: 2014 The Go Authors. All rights reserved.
// Original source:  https://raw.githubusercontent.com/golang/tools/master/cmd/callgraph/main.go
// Changes done by me, Avishkar Gupta <avgupta@redhat.com> according to my needs.
package serviceparser

import (
	"fmt"
	"go/token"
	"os"

	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/callgraph/static"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

// GetCompileTimeCalls returns a golang callgraph that contains all the edges we need to put between
// our functions that go into the callgraph.
func GetCompileTimeCalls(dir string, args []string) ([]CompileEdge, error) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "No main program/package in arguments.")
		return nil, nil
	}

	cfg := &packages.Config{
		Mode:  packages.LoadAllSyntax,
		Tests: false,
		Dir:   dir,
	}

	initial, err := packages.Load(cfg, args...)
	if err != nil {
		return nil, err
	}
	if packages.PrintErrors(initial) > 0 {
		return nil, fmt.Errorf("packages contain errors")
	}

	// Create and build SSA-form program representation.
	prog, _ := ssautil.AllPackages(initial, 0)
	prog.Build()

	// callgraph construction
	var cg *callgraph.Graph = static.CallGraph(prog)
	cg.DeleteSyntheticNodes()

	// Allocate these once, outside the traversal.
	var edges []CompileEdge
	if err := callgraph.GraphVisitEdges(cg, func(edge *callgraph.Edge) error {
		data := CompileEdge{fset: prog.Fset}
		data.position.Offset = -1
		data.edge = edge
		data.Caller = edge.Caller.Func
		data.Callee = edge.Callee.Func
		// TODO: Correct logic to return this struct and use it to create compile time paths in gremlin.
		edges = append(edges, data)
		return nil
	}); err != nil {
		return nil, err
	}
	return edges, nil
}

// CompileEdge for us represents a compile time edge.
type CompileEdge struct {
	Caller *ssa.Function
	Callee *ssa.Function

	edge     *callgraph.Edge
	fset     *token.FileSet
	position token.Position // initialized lazily
}

func (e *CompileEdge) pos() *token.Position {
	if e.position.Offset == -1 {
		e.position = e.fset.Position(e.edge.Pos()) // called lazily
	}
	return &e.position
}

// Filename gives the filename in which the call was made.
func (e *CompileEdge) Filename() string { return e.pos().Filename }

// Line gives the Line where the call was made.
func (e *CompileEdge) Line() int { return e.pos().Line }

// Description is a method that returns a description for the edge from the underlying callgraph edge.
func (e *CompileEdge) Description() string { return e.edge.Description() }
