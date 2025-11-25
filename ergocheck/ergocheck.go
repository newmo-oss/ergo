package ergocheck

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"iter"
	"log/slog"
	"regexp"
	"slices"

	"github.com/gostaticanalysis/analysisutil"
	"github.com/gostaticanalysis/ssainspect"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/ssa"

	"github.com/newmo-oss/ergo"
)

const doc = `ergocheck detects misuse usage as follows
* calling errors.New and fmt.Errorf
* calling ergo.New in package variable initializations
`

var Analyzer = &analysis.Analyzer{
	Name: "ergocheck",
	Doc:  doc,
	Run: func(pass *analysis.Pass) (any, error) {
		return new(runner).run(pass)
	},
	Requires: []*analysis.Analyzer{
		buildssa.Analyzer,
	},
}

var (
	flagPackages string
	flagExclues  string
)

func init() {
	Analyzer.Flags.StringVar(&flagPackages, "packages", "", "target pacakges import path (regexp)")
	Analyzer.Flags.StringVar(&flagExclues, "excludes", "", "excluded pacakges import path (regexp)")
}

type libFunc struct {
	pkg      string
	funcname string
}

type runner struct {
	targetPackgeRegexp   *regexp.Regexp
	excludePackageRegexp *regexp.Regexp
	pass                 *analysis.Pass
	libFuncs             map[string]*types.Func
	ssa                  *buildssa.SSA
}

func (r *runner) init(pass *analysis.Pass) error {
	if flagPackages != "" {
		targetPackgeRegexp, err := regexp.Compile(flagPackages)
		if err != nil {
			return ergo.Wrap(err, "failed to compile target packages import path regexp", slog.String("regexp", flagPackages))
		}
		r.targetPackgeRegexp = targetPackgeRegexp
	}

	if flagExclues != "" {
		excludePackageRegexp, err := regexp.Compile(flagExclues)
		if err != nil {
			return ergo.Wrap(err, "failed to compile excluded packages import path regexp", slog.String("regexp", flagExclues))
		}
		r.excludePackageRegexp = excludePackageRegexp
	}

	r.pass = pass
	builtSSA, ok := pass.ResultOf[buildssa.Analyzer].(*buildssa.SSA)
	if ok {
		r.ssa = builtSSA
	}

	r.libFuncs = r.getFuncs([]libFunc{
		{pkg: "errors", funcname: "New"},
		{pkg: "fmt", funcname: "Errorf"},
		{pkg: "github.com/newmo-oss/ergo", funcname: "New"},
		{pkg: "github.com/newmo-oss/ergo", funcname: "Wrap"},
		{pkg: "github.com/newmo-oss/ergo", funcname: "WithCode"},
	})

	return nil
}

func (r *runner) isTargetPkg(pkgpath string) bool {
	return (r.targetPackgeRegexp == nil || r.targetPackgeRegexp.MatchString(pkgpath)) &&
		(r.excludePackageRegexp == nil || !r.excludePackageRegexp.MatchString(pkgpath))
}

func (r *runner) run(pass *analysis.Pass) (any, error) {
	if err := r.init(pass); err != nil {
		return nil, err
	}

	if r.ssa == nil {
		// skip
		return nil, nil
	}

	pkgpath := pass.Pkg.Path()
	if !r.isTargetPkg(pkgpath) {
		// skip
		return nil, nil
	}

	funcs := slices.Clone(r.ssa.SrcFuncs)

	// Add dummy functions that correspond to variable initialization,
	// because SrcFunc does not have it.
	for _, m := range r.ssa.Pkg.Members {
		f, ok := m.(*ssa.Function)
		// exclude package functions and methods
		if ok && !slices.Contains(funcs, f) {
			funcs = append(funcs, f)

			// Add function literals in pacakge variable initilization.
			// These functions belong to the dummy function as the AnonFuncs field.
			//
			// var _ = func() {
			// 	fmt.Errorf("error: %w", err)
			// }
			//
			for _, anon := range f.AnonFuncs {
				if !slices.Contains(funcs, anon) {
					funcs = append(funcs, anon)
				}
			}
		}
	}

	for cur := range ssainspect.All(funcs) {
		r.checkDeprecatedFunc(cur.Instr)
		r.checkFormatString(cur.Instr)
		r.checkNilErr(cur.Instr)
	}

	r.checkVarInit()

	return nil, nil
}

func (r *runner) getFuncs(libFuncs []libFunc) map[string]*types.Func {
	m := make(map[string]*types.Func)
	for _, libFunc := range libFuncs {
		f, ok := analysisutil.ObjectOf(r.pass, libFunc.pkg, libFunc.funcname).(*types.Func)
		if ok {
			m[libFunc.pkg+"."+libFunc.funcname] = f
		}
	}
	return m
}

type deprecatedFunc struct {
	obj     *types.Func
	suggest string
}

func (r *runner) checkDeprecatedFunc(instr ssa.Instruction) {
	deprecatedFuncs := []deprecatedFunc{
		{obj: r.libFuncs["errors.New"], suggest: "ergo.New"},
		{obj: r.libFuncs["fmt.Errorf"], suggest: "ergo.Wrap"},
	}

	for _, deprecated := range deprecatedFuncs {
		if deprecated.obj == nil {
			continue
		}
		if analysisutil.Called(instr, nil, deprecated.obj) {
			r.pass.Reportf(instr.Pos(), "%s must not be used in the %s package, it should be replaced by %s", deprecated.obj.FullName(), r.pass.Pkg.Path(), deprecated.suggest)
		}
	}
}

var formatRegexp = regexp.MustCompile(`%[+#]?[vsdT]`)

type formatFunc struct {
	obj *types.Func
	arg int
}

func (r *runner) checkFormatString(instr ssa.Instruction) {
	funcs := []formatFunc{
		{obj: r.libFuncs["github.com/newmo-oss/ergo.New"], arg: 0},
		{obj: r.libFuncs["github.com/newmo-oss/ergo.Wrap"], arg: 1},
	}

	for _, f := range funcs {
		if f.obj == nil {
			continue
		}

		if !analysisutil.Called(instr, nil, f.obj) {
			continue
		}

		call, ok := instr.(*ssa.Call)
		if !ok {
			continue
		}

		if f.arg > len(call.Call.Args)-1 {
			continue
		}

		msgarg, ok := call.Call.Args[f.arg].(*ssa.Const)
		if !ok || !types.Identical(msgarg.Type().Underlying(), types.Typ[types.String]) {
			continue
		}

		msg := constant.StringVal(msgarg.Value)
		if formatRegexp.MatchString(msg) {
			r.pass.Reportf(instr.Pos(), `the message of %s must not be format string such as "xxxx %%s": %q`, f.obj.FullName(), msg)
		}
	}
}

type wrapFunc struct {
	obj *types.Func
	arg int
}

func (r *runner) checkNilErr(instr ssa.Instruction) {
	funcs := []wrapFunc{
		{obj: r.libFuncs["github.com/newmo-oss/ergo.WithCode"], arg: 0},
		{obj: r.libFuncs["github.com/newmo-oss/ergo.Wrap"], arg: 0},
	}

	for _, f := range funcs {
		if f.obj == nil {
			continue
		}

		if !analysisutil.Called(instr, nil, f.obj) {
			continue
		}

		call, ok := instr.(*ssa.Call)
		if !ok {
			continue
		}

		if f.arg > len(call.Call.Args)-1 {
			continue
		}

		errarg := call.Call.Args[f.arg]
		if isNil(call.Block(), errarg) {
			r.pass.Reportf(instr.Pos(), `The %s argument of %s must not be nil`, ordinalNumber(f.arg+1), f.obj.FullName())
		}
	}
}

func ordinalNumber(n int) string {
	switch n {
	case 1:
		return "1st"
	case 2:
		return "2nd"
	case 3:
		return "3rd"
	}
	return fmt.Sprintf("%dth", n)
}

func isNil(b *ssa.BasicBlock, v ssa.Value) bool {
	switch v := v.(type) {
	case *ssa.Const:
		return v.IsNil()
	case *ssa.Phi:
		if hasNilGuard(b, v) {
			return false
		}

		if slices.ContainsFunc(v.Edges, func(v ssa.Value) bool {
			return isNil(b, v)
		}) {
			return true
		}
	}
	return false
}

// hasNilGuard checks whether a value is guaranteed to be non-nil by a conditional branch such as if err != nil.
// For example, the value of first argument of ergo.Wrap in the following function f, is guaranteed to be non-nil by if err != nil.
//
//	func f() error {
//		// (0)
//		var err error
//		if cond {
//			// --> (1)
//			err := ergo.New("error")
//		}
//		// --> (2)*
//
//		if err != nil {
//			// --> ((3))
//			retrn ergo.Wrap(err, "wrap")
//		}
//
//		// --> ((4))
//		return nil
//	}
//
// The control flow graph of function f can be represented as follows.
// A node of the graph means a basic block of SSA form.
// The node numbers correspond to the comments in the function f.
// There are three types of node, start node, terminate node and star node.
// A start node is a first block of a function.
// A terminate node is the block which has a return instruction (*ssa.Return).
// A star node is the block which has an if instruction (*ssa.If) such as if err != nil.
// Here we call if instructions such as if err != nil, "nil guards".
//
//	(0)
//	 |
//	 | if cond then
//	 +---->(1)
//	 |      |
//	 | else | if err != nil then
//	 +---->(2)*----->((3))
//	        |
//	        | else
//	        v
//	       ((4))
//
//	(0)  : Start node (first block)
//	((n)): Terminate node (return block)
//	(n)* : Star node (the block has if err != nil instructions)
//
// The algorithms of hasNilGuard are divided into the following:
//
//  1. Finding nil guards (a star node) in referrers of the given ssa.Value.
//  2. Removing the star nodes from the control flow graph.
//     2.1 In fact, the star node will be ignored in backtracing.
//  3. Backtracing the control flow graph from the given ssa.BasicBlock (node) to the start node.
//
// If it reached to the start node, the control flow graph has one or more pathes which are not including star nodes (nil guards), between from the given ssa.Block to the start node.
func hasNilGuard(b *ssa.BasicBlock, v ssa.Value) bool {
	if _, ok := v.(ssa.Instruction); !ok {
		return false
	}

	refs := v.Referrers()
	if refs == nil {
		return false
	}

	guards := make(map[ssa.Instruction]bool)
	for _, ref := range *refs {
		if guard := toNilGuard(v, ref); guard != nil {
			guards[guard] = true
		}
	}

	if len(guards) == 0 {
		return false
	}

	if backtrace(b, make(map[*ssa.BasicBlock]bool), func(b *ssa.BasicBlock) bool {
		n := len(b.Instrs)
		if n == 0 {
			return false
		}
		return guards[b.Instrs[n-1]]
	}, func(b *ssa.BasicBlock) bool {
		return b.Index == 0
	}) {
		return false
	}

	return true
}

func toNilGuard(v ssa.Value, instr ssa.Instruction) *ssa.If {
	cond, ok := instr.(*ssa.BinOp)
	if !ok {
		return nil
	}

	guard := getIf(cond)
	if guard == nil {
		return nil
	}

	if equalValue(cond.X, v) && isNil(instr.Block(), cond.Y) {
		return guard
	}

	if equalValue(cond.Y, v) && isNil(instr.Block(), cond.X) {
		return guard
	}

	return nil
}

func getIf(cond *ssa.BinOp) *ssa.If {
	refs := cond.Referrers()
	if refs == nil {
		return nil
	}

	for _, ref := range *refs {
		guard, ok := ref.(*ssa.If)
		if ok {
			return guard
		}
	}

	return nil
}

func equalValue(x, y ssa.Value) bool {
	if x == y {
		return true
	}

	for x := range phiValues(x) {
		if x == y {
			return true
		}
	}

	for y := range phiValues(y) {
		if x == y {
			return true
		}
	}

	return false
}

func phiValues(v ssa.Value) iter.Seq[ssa.Value] {
	return func(yield func(ssa.Value) bool) {
		phi, ok := v.(*ssa.Phi)
		if !ok {
			yield(v)
			return
		}

		for _, v := range phi.Edges {
			for v := range phiValues(v) {
				if !yield(v) {
					return
				}
			}
		}
	}
}

func backtrace(b *ssa.BasicBlock, done map[*ssa.BasicBlock]bool, drop, terminate func(b *ssa.BasicBlock) bool) bool {
	if done[b] {
		return false
	}
	done[b] = true

	if terminate(b) {
		return true
	}

	for _, pre := range b.Preds {
		if !drop(pre) {
			if backtrace(pre, done, drop, terminate) {
				return true
			}
		}
	}

	return false
}

func (r *runner) checkVarInit() {
	ergoNew, ok := r.libFuncs["github.com/newmo-oss/ergo.New"]
	if !ok {
		return
	}

	ergoWrap, ok := r.libFuncs["github.com/newmo-oss/ergo.Wrap"]
	if !ok {
		return
	}

	for _, file := range r.pass.Files {
		for _, decl := range file.Decls {
			gendecl, ok := decl.(*ast.GenDecl)
			if !ok || gendecl.Tok != token.VAR {
				continue
			}

			for _, spec := range gendecl.Specs {
				valspec, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}

				for _, val := range valspec.Values {
					call, ok := val.(*ast.CallExpr)
					if !ok {
						continue
					}

					fun := r.getCallFun(call.Fun)
					if fun == nil {
						continue
					}

					switch fun {
					case ergoNew:
						r.pass.Reportf(call.Pos(), "%s must not be used in package variable initilization, it should be replaced by ergo.NewSentinel", fun.FullName())
					case ergoWrap:
						r.pass.Reportf(call.Pos(), "%s must not be used in package variable initilization, it should be replaced by errors.Join", fun.FullName())
					}
				}
			}
		}
	}
}

func (r *runner) getCallFun(fun ast.Expr) *types.Func {
	switch fun := fun.(type) {
	case *ast.Ident:
		obj, ok := r.pass.TypesInfo.ObjectOf(fun).(*types.Func)
		if ok {
			return obj
		}
	case *ast.SelectorExpr:
		obj, ok := r.pass.TypesInfo.ObjectOf(fun.Sel).(*types.Func)
		if ok {
			return obj
		}
	}
	return nil
}
