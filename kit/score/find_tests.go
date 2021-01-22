package score

import (
	"flag"
	"fmt"
	"go/ast"
	"log"
	"os"
	"reflect"
	"runtime"
	"strings"

	"golang.org/x/tools/go/packages"
)

func FindTests() {
	flag.Parse()
	cfg := &packages.Config{Mode: packages.NeedFiles | packages.NeedSyntax, Tests: true}
	pkgs, err := packages.Load(cfg, flag.Args()...)
	if err != nil {
		log.Fatal(err)
	}
	if packages.PrintErrors(pkgs) > 0 {
		os.Exit(1)
	}

	// Print the names of the source files
	// for each package listed on the command line.
	for _, pkg := range pkgs {
		// fmt.Println(pkg.ID, pkg.GoFiles)
		for _, f := range pkg.Syntax {
			// fmt.Println(f)
			for _, decl := range f.Decls {
				switch fn := decl.(type) {
				case *ast.FuncDecl:
					if fn.Recv != nil || !strings.HasPrefix(fn.Name.Name, "Test") {
						// Ignore non-test functions and functions with receivers.
						continue
					}
					fmt.Println("Found: ", fn.Name)
					checkTest(fn, "Test")
				}
			}
		}
	}
}

var scoreFuncs = []string{
	fnName(NewScore),
	fnName(NewScoreMax),
}

func fnName(x interface{}) string {
	name := runtime.FuncForPC(reflect.ValueOf(x).Pointer()).Name()
	return name[strings.LastIndex(name, ".")+1:]
}

func checkTest(fn *ast.FuncDecl, prefix string) {
	// Want functions with 0 results and 1 parameter.
	if fn.Type.Results != nil && len(fn.Type.Results.List) > 0 ||
		fn.Type.Params == nil ||
		len(fn.Type.Params.List) != 1 ||
		len(fn.Type.Params.List[0].Names) > 1 {
		return
	}

	// The param must look like a *testing.T or *testing.B.
	if !isTestParam(fn.Type.Params.List[0].Type, prefix[:1]) {
		return
	}

	ast.Inspect(fn, func(n ast.Node) bool {
		switch call := n.(type) {
		case *ast.CallExpr:
			if isScoreFunc(call) {
				for i, arg := range call.Args {
					processArg(i, arg)
				}
			}
		}
		return true
	})
}

func processArg(i int, arg ast.Expr) {
	ast.Inspect(arg, func(n ast.Node) bool {
		var s string
		switch x := n.(type) {
		case *ast.BasicLit:
			s = x.Value
		case *ast.Ident:
			s = x.Name
		}
		if s != "" {
			fmt.Printf("\t\targ[%d]: %s (%v)\n", i, s, arg)
		}
		return true
	})
}

func isScoreFunc(typ ast.Expr) (found bool) {
	ast.Inspect(typ, func(n ast.Node) bool {
		if name, ok := n.(*ast.Ident); ok {
			for _, scoreFn := range scoreFuncs {
				if name.Name == scoreFn {
					found = true
					fmt.Printf("\tname expression found: %v %s (%v)\n", name, name.Name, typ)
				}
			}
		}
		return true
	})
	return
}

func isTestParam(typ ast.Expr, wantType string) bool {
	ptr, ok := typ.(*ast.StarExpr)
	if !ok {
		// Not a pointer.
		return false
	}
	// No easy way of making sure it's a *testing.T or *testing.B:
	// ensure the name of the type matches.
	if name, ok := ptr.X.(*ast.Ident); ok {
		return name.Name == wantType
	}
	if sel, ok := ptr.X.(*ast.SelectorExpr); ok {
		return sel.Sel.Name == wantType
	}
	return false
}
