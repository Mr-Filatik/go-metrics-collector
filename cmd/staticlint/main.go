package main

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"

	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"

	"github.com/Mr-Filatik/go-metrics-collector/cmd/staticlint/analyzer/osexit"
	"github.com/kisielk/errcheck/errcheck"
)

func main() {
	var analyzers []*analysis.Analyzer

	analyzers = addStandartAnalyzers(analyzers)

	analyzers = addStaticCheckSAAnalyzers(analyzers)

	analyzers = addStaticCheckOtherAnalyzers(analyzers)

	analyzers = addThirdPartAnalyzers(analyzers)

	analyzers = addCustomAnalyzers(analyzers)

	multichecker.Main(analyzers...)
}

func addStandartAnalyzers(slice []*analysis.Analyzer) []*analysis.Analyzer {
	slice = append(slice,
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		bools.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		httpresponse.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		printf.Analyzer,
		shift.Analyzer,
		stdmethods.Analyzer,
		stringintconv.Analyzer,
		structtag.Analyzer,
		testinggoroutine.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,
	)

	return slice
}

func addStaticCheckSAAnalyzers(slice []*analysis.Analyzer) []*analysis.Analyzer {
	for i := range staticcheck.Analyzers {
		slice = append(slice, staticcheck.Analyzers[i].Analyzer)
	}

	return slice
}

func addStaticCheckOtherAnalyzers(slice []*analysis.Analyzer) []*analysis.Analyzer {
	checks := map[string]bool{
		"ST1000": true,
		"ST1023": true,
	}

	for i := range stylecheck.Analyzers {
		if checks[stylecheck.Analyzers[i].Analyzer.Name] {
			slice = append(slice, stylecheck.Analyzers[i].Analyzer)
		}
	}

	return slice
}

func addThirdPartAnalyzers(slice []*analysis.Analyzer) []*analysis.Analyzer {
	for i := range simple.Analyzers {
		slice = append(slice, simple.Analyzers[i].Analyzer)
	}

	slice = append(slice, errcheck.Analyzer)

	return slice
}

func addCustomAnalyzers(slice []*analysis.Analyzer) []*analysis.Analyzer {
	slice = append(slice, osexit.Analyzer)

	return slice
}
