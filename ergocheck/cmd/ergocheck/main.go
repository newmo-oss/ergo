package main

import (
	"github.com/newmo-oss/ergo/ergocheck"
	"golang.org/x/tools/go/analysis/unitchecker"
)

func main() {
	unitchecker.Main(ergocheck.Analyzer)
}
