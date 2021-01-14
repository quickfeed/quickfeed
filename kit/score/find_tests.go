package score

import (
	"fmt"
	"go/ast"
	"go/build"
	"log"
	"strings"

	"golang.org/x/tools/go/analysis"
)

const Doc = `FIXME: check for common mistaken usages of tests and examples
The tests checker walks Test, Benchmark and Example functions checking
malformed names, wrong signatures and examples documenting non-existent
identifiers.
Please see the documentation for package testing in golang.org/pkg/testing
for the conventions that are enforced for Tests, Benchmarks, and Examples.`

var Analyzer = &analysis.Analyzer{
	Name: "scores",
	Doc:  Doc,
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, f := range pass.Files {
		if !strings.HasSuffix(pass.Fset.File(f.Pos()).Name(), "_test.go") {
			continue
		}
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil {
				// Ignore non-functions or functions with receivers.
				continue
			}

			switch {
			// case strings.HasPrefix(fn.Name.Name, "Example"):
			// checkExample(pass, fn)
			case strings.HasPrefix(fn.Name.Name, "Test"):
				fmt.Println("Found: ", fn.Name.Name)
				// checkTest(pass, fn, "Test")
				// case strings.HasPrefix(fn.Name.Name, "Benchmark"):
				// checkTest(pass, fn, "Benchmark")
			}
		}
	}
	return nil, nil
}

func FindTests() []string {
	pkg, err := build.ImportDir("/Users/meling/work/quickfeed/kit/score", build.ImportMode(build.FindOnly))
	if err != nil {
		log.Fatal(err)
	}
	return pkg.TestGoFiles
}
