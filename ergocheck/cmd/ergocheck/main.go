package main

import (
	"github.com/newmo-oss/ergo/ergocheck"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(ergocheck.Analyzer)
}
