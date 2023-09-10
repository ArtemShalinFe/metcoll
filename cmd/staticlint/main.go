// The package is used to combine analyzers.
// Built on the basis of multichecker golang.org/x/tools/go/analysis/multichecker.
//
// The package includes the following analyzers from the package golang.org/x/tools/go/analysis:
//   - assign (defines an Analyzer that detects useless assignments)
//   - atomic (defines an Analyzer that checks for common mistakes using the sync/atomic package)
//   - bools (defines an Analyzer that detects common mistakes involving boolean operators)
//   - defers (defines an Analyzer that checks for common mistakes in defer statements)
//   - errorsas (defines an Analyzer that checks that the second argument to errors.As is a pointer
//     to a type implementing error)
//   - fieldalignment (defines an Analyzer that detects structs that would use less memory if their fields were sorted)
//   - loopclosure (defines an Analyzer that checks for references to enclosing loop variables
//     from within nested functions)
//   - printf (defines an Analyzer that checks consistency of Printf format strings and arguments)
//   - stringintconv (defines an Analyzer that flags type conversions from integers to strings)
//   - structtag (defines an Analyzer that checks struct field tags are well formed)
//
// The package includes the following analyzers from the package honnef.co/go/tools/:
//   - simple (contains analyzes that simplify code)
//   - staticcheck (contains analyzes that find bugs and performance issues)
//   - stylecheck (contains analyzes that enforce style rules)
//
// The package includes analyzer from the package "github.com/tommy-muehle/go-mnd/v2", that detect "magic numbers".
//
// The package includes analyzer "exitchecker",
//
//	that checks a direct call to `os.Exit` in the main function of the main package.
package main

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/defers"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"

	mnd "github.com/tommy-muehle/go-mnd/v2"

	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"

	"github.com/ArtemShalinFe/metcoll/internal/exitchecker"
)

func main() {
	var checks []*analysis.Analyzer

	for _, v := range staticcheck.Analyzers {
		checks = append(checks, v.Analyzer)
	}

	sc := map[string]bool{
		"S1000": true,
		"S1003": true,
	}
	for _, v := range simple.Analyzers {
		if sc[v.Analyzer.Name] {
			checks = append(checks, v.Analyzer)
		}
	}

	stc := map[string]bool{
		"ST1000": true,
		"ST1003": true,
	}
	for _, v := range stylecheck.Analyzers {
		if stc[v.Analyzer.Name] {
			checks = append(checks, v.Analyzer)
		}
	}

	checks = append(checks,
		assign.Analyzer,
		atomic.Analyzer,
		bools.Analyzer,
		defers.Analyzer,
		errorsas.Analyzer,
		fieldalignment.Analyzer,
		loopclosure.Analyzer,
		stringintconv.Analyzer,
		printf.Analyzer,
		structtag.Analyzer,
		mnd.Analyzer,
		exitchecker.NewAnalyzer())

	multichecker.Main(
		checks...,
	)
}
