package ergocheck_test

import (
	"testing"

	"github.com/gostaticanalysis/testutil"
	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/newmo-oss/ergo/ergocheck"
)

// TestAnalyzer is a test for Analyzer.
func TestAnalyzer(t *testing.T) {
	t.Parallel()
	modfile := testutil.ModFile(t, ".", nil)
	testdata := testutil.WithModules(t, analysistest.TestData(), modfile)

	if err := ergocheck.Analyzer.Flags.Set("packages", ".+/a$"); err != nil {
		t.Fatal("failed to set packages to ergocheck.Analyzer")
	}

	if err := ergocheck.Analyzer.Flags.Set("excludes", ".+/exclue$"); err != nil {
		t.Fatal("failed to set excludes to ergocheck.Analyzer")
	}

	// these packages for test are in testdata/src
	pkgs := []string{
		"github.com/newmo-oss/a",
		"github.com/newmo-oss/notarget",
		"github.com/newmo-oss/exclude",
	}
	analysistest.Run(t, testdata, ergocheck.Analyzer, pkgs...)
}
