package geneticalgorithm

import (
	"6CCS3PRJ/thomasholden/util"
	"bytes"
)

// PassManager initializes and contains pass data.
type PassManager struct {
	BidirectionalMap[string, int]
	MaxPass int // MaxPass is the highest index/size of the mapping in PassManager

	moduleRange   *util.Range[int]
	cgsccRange    *util.Range[int]
	functionRange *util.Range[int]
	loopRange     *util.Range[int]
}

var passManager *PassManager = nil

// GetPassManager returns a PassManager singleton.
func GetPassManager() *PassManager {
	if passManager == nil {
		passManager = newPassManager()
	}
	return passManager
}

// NewPassManager initializes and returns a PassManager pointer.
func newPassManager() *PassManager {
	manager := &PassManager{}
	manager.init()
	manager.loadPasses()

	return manager
}

// loadPasses adds all passes to the mapping
func (pm *PassManager) loadPasses() {
	passList := getPasses()
	passList = append(passList, getO3String()...)

	assignments := 0
	for _, pass := range passList {
		if _, ok := pm.direct[pass]; !ok {
			pm.Put(pass, assignments)
			assignments++
		}
	}
	pm.MaxPass = assignments

	pm.initRanges()
}

func (pm *PassManager) initRanges() {
	pm.moduleRange = util.NewRange(
		*pm.Get("always-inline"),
		*pm.Get("wholeprogramdevirt"),
	)
	pm.cgsccRange = util.NewRange(
		*pm.Get("argpromotion"),
		*pm.Get("coro-annotation-elide"),
	)
	pm.functionRange = util.NewRange(
		*pm.Get("adce"),
		*pm.Get("wasm-eh-prepare"),
	)
	pm.loopRange = util.NewRange(
		*pm.Get("loop-flatten"),
		*pm.Get("loop-versioning-licm"),
	)
}

// Passes extracted from
// https://github.com/llvm/llvm-project/blob/d162c91c01a66e4af0af190044961e60db0eeb3d/llvm/lib/Passes/PassRegistry.def
//
// All passes including keywords analysis, print, view, dot, verify, callback, eval, no-op, prof,
// pgo, trigger, and invalidate are removed.
// Other miscellaneous and fragile passes removed e.g. helloworld
func getPasses() []string {
	return []string{
		"always-inline", // start module passes - <pass>
		"assign-guid",
		"attributor",
		"attributor-light",
		"called-value-propagation",
		"canonicalize-aliases",
		// "cg-profile", // requires profile (probably)
		"constmerge",
		"coro-cleanup",
		"coro-early",
		"cross-dso-cfi",
		// "ctx-instr-gen", // not beneficial for performance + fragile
		// "ctx-prof-flatten",
		// "ctx-prof-flatten-prethinlink", // both require profile
		"noinline-nonprevailing",
		"deadargelim",
		// "declare-runtime-libcalls", // unknown module pass
		// "dfsan", // requires libraries
		"dxil-upgrade",
		"elim-avail-extern",
		"extract-blocks",
		"expand-variadics",
		"forceattrs",
		//"function-import", requires file
		"global-merge-func",
		"globalopt",
		"globalsplit",
		/*"hipstdpar-interpose-alloc",
		"hipstdpar-math-fixup",
		"hipstdpar-select-accelerator-code",*/ // All platform specific
		"hotcoldsplit",
		"inferattrs",
		//"inliner-ml-advisor-release", // too fragile
		"inliner-wrapper",
		"inliner-wrapper-no-mandatory-first",
		// "insert-gcov-profiling", // requires profile (probably)
		// "instrprof", // requires profile (probably)
		// "ctx-instr-lower", // not beneficial for performance + fragile
		"iroutliner",
		"jmc-instrumenter",
		"lower-emutls",
		"lower-global-dtors",
		"lower-ifunc",
		"simplify-type-tests",
		"lowertypetests",
		"fatlto-cleanup",
		// "pgo-force-function-attrs", // requires profile
		// "memprof-context-disambiguation", // requires profile
		// "memprof-remove-attributes", // unknown module pass
		//"memprof-module", // requires memprof data
		"mergefunc",
		"module-inline",
		// "norecurse-lto-inference", // unknown module pass
		// "nsan", // requires libraries
		"openmp-opt",
		"openmp-opt-postlink",
		"partial-inliner",
		// "pgo-icall-prom", // requires profile
		// "pgo-instr-gen", // requires profile
		//"pgo-instr-use", // requires profile
		"pre-isel-intrinsic-lowering",
		"pseudo-probe",
		"pseudo-probe-update",
		"rel-lookup-table-converter",
		"rewrite-statepoints-for-gc",
		"rewrite-symbols",
		"rpo-function-attrs",
		// "rtsan", // requires libraries
		//"sample-profile", // requires profile
		"sancov-module",
		"sanmd-module",
		"scc-oz-module-inliner",
		"shadow-stack-gc-lowering",
		// "tsan-module", // requires libraries
		// "tysan", // requires libraries
		"wholeprogramdevirt", // end module passes
		"argpromotion",       // start cgscc passes - cgscc(<pass>)
		"attributor-cgscc",
		"attributor-light-cgscc",
		"openmp-opt-cgscc",
		// start O3 cgscc passes
		"inline",
		"function-attrs<skip-non-recursive-function-attrs>",
		"function-attrs",
		// end O3 cgscc passes
		"coro-annotation-elide", // end cgscc passes
		"adce",                  // start function passes function(<pass>)
		"add-discriminators",
		"aggressive-instcombine",
		"alignment-from-assumptions",
		"assume-builder",
		"assume-simplify",
		"atomic-expand",
		"bdce",
		"break-crit-edges",
		"callbr-prepare",
		"callsite-splitting",
		"chr",
		"codegenprepare",
		"complex-deinterleaving",
		"consthoist",
		"constraint-elimination",
		"coro-elide",
		"correlated-propagation",
		"dce",
		"declare-to-assign", // Seems to be a module pass but is declared as a function pass
		// Transformation is the same whether wrapped or not
		"dfa-jump-threading",
		"div-rem-pairs",
		"dse",
		"dwarf-eh-prepare",
		"expand-large-div-rem",
		"expand-memcmp",
		"expand-reductions",
		"extra-vector-passes",
		"fix-irreducible",
		"flatten-cfg",
		"float2int",
		"gc-lowering",
		"guard-widening",
		"gvn-hoist",
		"gvn-sink",
		"indirectbr-expand",
		"infer-address-spaces",
		"infer-alignment",
		"inject-tli-mappings",
		"instsimplify",
		"interleaved-access",
		"interleaved-load-combine",
		"irce",
		"jump-threading",
		"jump-table-to-switch",
		"kcfi",
		"lcssa",
		"libcalls-shrinkwrap",
		"load-store-vectorizer",
		"loop-data-prefetch",
		"loop-distribute",
		"loop-fusion",
		"loop-load-elim",
		"loop-simplify",
		"loop-sink",
		"loop-versioning",
		"lower-atomic",
		"lower-constant-intrinsics",
		"lower-expect",
		"lower-guard-intrinsic",
		"lower-invoke",
		"lower-switch",
		"lower-widenable-condition",
		"make-guards-explicit",
		"mem2reg",
		"memcpyopt",
		// "memprof", // requires profile
		"mergeicmps",
		"mergereturn",
		"move-auto-init",
		"nary-reassociate",
		"newgvn",
		"objc-arc",
		"objc-arc-contract",
		"objc-arc-expand",
		"partially-inline-libcalls",
		// "pgo-memop-opt", // requires profile
		"place-safepoints",
		"reassociate",
		"redundant-dbg-inst-elim",
		"replace-with-veclib",
		"reg2mem",
		"sandbox-vectorizer",
		"scalarize-masked-mem-intrin",
		"sccp",
		"select-optimize",
		"separate-const-offset-from-gep",
		"sink",
		"simplifycfg",
		"sjlj-eh-prepare",
		"slp-vectorizer",
		"slsr",
		"tailcallelim",
		// "tsan", // requires libraries
		"typepromotion",
		"unify-loop-exits",
		"unreachableblockelim",
		"vector-combine",
		// start O3 specific functional passes
		"instcombine<max-iterations=1;no-verify-fixpoint>",
		"ee-instrument<>",
		"simplifycfg<bonus-inst-threshold=1;" +
			"no-forward-switch-cond;no-switch-range-to-icmp;no-switch-to-lookup;keep-loops;" +
			"no-hoist-common-insts;no-hoist-loads-stores-with-cond-faulting;no-sink-common-insts;" +
			"speculate-blocks;simplify-cond-branch;no-speculate-unpredictables>,sroa<modify-cfg>",
		"early-cse<>",
		"simplifycfg<bonus-inst-threshold=1;" +
			"no-forward-switch-cond;switch-range-to-icmp;no-switch-to-lookup;keep-loops;" +
			"no-hoist-common-insts;no-hoist-loads-stores-with-cond-faulting;no-sink-common-insts;" +
			"speculate-blocks;simplify-cond-branch;no-speculate-unpredictables>",
		"loop-vectorize<no-interleave-forced-only;no-vectorize-forced-only;>",
		"simplifycfg<bonus-inst-threshold=1;" +
			"forward-switch-cond;switch-range-to-icmp;switch-to-lookup;no-keep-loops;" +
			"hoist-common-insts;no-hoist-loads-stores-with-cond-faulting;sink-common-insts;" +
			"speculate-blocks;simplify-cond-branch;no-speculate-unpredictables>",
		"vector-combine",
		"loop-unroll<O3>",
		"transform-warning",
		"sroa<preserve-cfg>",
		"loop-mssa(licm<allowspeculation>)",
		"simplifycfg<bonus-inst-threshold=1;" +
			"no-forward-switch-cond;switch-range-to-icmp;no-switch-to-lookup;keep-loops;" +
			"no-hoist-common-insts;hoist-loads-stores-with-cond-faulting;no-sink-common-insts;" +
			"speculate-blocks;simplify-cond-branch;speculate-unpredictables>",
		// end O3 specific functional passes
		"wasm-eh-prepare", // end function passes
		"loop-flatten",    // start loopnest passes - function(loop(<pass>))
		"loop-interchange",
		"loop-unroll-and-jam", // end loopnest passes
		"canon-freeze",        // start loop passes - function(loop(<pass>))
		"guard-widening",
		"extra-simple-loop-unswitch-passes",
		"indvars",
		"loop-bound-split",
		"loop-deletion",
		"loop-idiom",
		"loop-idiom-vectorize",
		"loop-instsimplify",
		// "loop-mssa(licm<allowspeculation>)", // Is a pass manager
		"loop-predication",
		"loop-reduce",
		"loop-term-fold",
		"loop-simplifycfg",
		"loop-unroll-full",
		// start O3 loop passes
		"loop-rotate<header-duplication;no-prepare-for-lto>",
		// end O3 loop passes
		"loop-versioning-licm", // end loop passes
	}
}

// Returns the list of strings making up O3
// Has had minor tweaks to remove passes intended for only debugging
func getO3String() []string {
	return []string{
		"annotation2metadata",
		"forceattrs",
		"declare-to-assign",
		"inferattrs",
		"coro-early",
		"ee-instrument<>",
		"lower-expect",
		"simplifycfg<bonus-inst-threshold=1;" +
			"no-forward-switch-cond;no-switch-range-to-icmp;no-switch-to-lookup;keep-loops;" +
			"no-hoist-common-insts;no-hoist-loads-stores-with-cond-faulting;no-sink-common-insts;" +
			"speculate-blocks;simplify-cond-branch;no-speculate-unpredictables>,sroa<modify-cfg>",
		"early-cse<>",
		"callsite-splitting",
		"openmp-opt",
		"ipsccp",
		"called-value-propagation",
		"globalopt",
		"mem2reg",
		"instcombine<max-iterations=1;no-verify-fixpoint>",
		"simplifycfg<bonus-inst-threshold=1;no-forward-switch-cond;switch-range-to-icmp;" +
			"no-switch-to-lookup;keep-loops;no-hoist-common-insts;" +
			"no-hoist-loads-stores-with-cond-faulting;no-sink-common-insts;speculate-blocks;" +
			"simplify-cond-branch;no-speculate-unpredictables>",
		"always-inline",
		"require<globals-aa>",
		"function(invalidate<aa>)",
		"require<profile-summary>",
		"cgscc(devirt<4>(inline,function-attrs<skip-non-recursive-function-attrs>,argpromotion," +
			"openmp-opt-cgscc,function<eager-inv;no-rerun>(sroa<modify-cfg>,early-cse<memssa>," +
			"speculative-execution<only-if-divergent-target>,jump-threading,correlated-propagation," +
			"simplifycfg<bonus-inst-threshold=1;no-forward-switch-cond;switch-range-to-icmp;" +
			"no-switch-to-lookup;keep-loops;no-hoist-common-insts;" +
			"no-hoist-loads-stores-with-cond-faulting;no-sink-common-insts;speculate-blocks;" +
			"simplify-cond-branch;no-speculate-unpredictables>,instcombine<max-iterations=1;" +
			"no-verify-fixpoint>,aggressive-instcombine,libcalls-shrinkwrap,tailcallelim," +
			"simplifycfg<bonus-inst-threshold=1;no-forward-switch-cond;switch-range-to-icmp;" +
			"no-switch-to-lookup;keep-loops;no-hoist-common-insts;" +
			"no-hoist-loads-stores-with-cond-faulting;no-sink-common-insts;speculate-blocks;" +
			"simplify-cond-branch;no-speculate-unpredictables>,reassociate,constraint-elimination," +
			"loop-mssa(loop-instsimplify,loop-simplifycfg,licm<no-allowspeculation>," +
			"loop-rotate<header-duplication;no-prepare-for-lto>,licm<allowspeculation>," +
			"simple-loop-unswitch<nontrivial;trivial>),simplifycfg<bonus-inst-threshold=1;" +
			"no-forward-switch-cond;switch-range-to-icmp;no-switch-to-lookup;keep-loops;" +
			"no-hoist-common-insts;no-hoist-loads-stores-with-cond-faulting;no-sink-common-insts;" +
			"speculate-blocks;simplify-cond-branch;no-speculate-unpredictables>," +
			"instcombine<max-iterations=1;no-verify-fixpoint>,loop(loop-idiom,indvars," +
			"extra-simple-loop-unswitch-passes,loop-deletion,loop-unroll-full),sroa<modify-cfg>," +
			"vector-combine,mldst-motion<no-split-footer-bb>,gvn<>,sccp,bdce," +
			"instcombine<max-iterations=1;no-verify-fixpoint>,jump-threading,correlated-propagation," +
			"adce,memcpyopt,dse,move-auto-init,loop-mssa(licm<allowspeculation>),coro-elide," +
			"simplifycfg<bonus-inst-threshold=1;no-forward-switch-cond;switch-range-to-icmp;" +
			"no-switch-to-lookup;keep-loops;hoist-common-insts;" +
			"no-hoist-loads-stores-with-cond-faulting;sink-common-insts;speculate-blocks;" +
			"simplify-cond-branch;no-speculate-unpredictables>,instcombine<max-iterations=1;" +
			"no-verify-fixpoint>),function-attrs,coro-split,coro-annotation-elide))",
		"deadargelim",
		"coro-cleanup",
		"globalopt",
		"globaldce",
		"elim-avail-extern",
		"rpo-function-attrs",
		"recompute-globalsaa",
		"float2int",
		"lower-constant-intrinsics",
		"chr",
		"loop-rotate<header-duplication;no-prepare-for-lto>",
		"loop-deletion",
		"loop-distribute",
		"inject-tli-mappings",
		"loop-vectorize<no-interleave-forced-only;no-vectorize-forced-only;>",
		"infer-alignment",
		"loop-load-elim",
		"instcombine<max-iterations=1;no-verify-fixpoint>",
		"simplifycfg<bonus-inst-threshold=1;forward-switch-cond;switch-range-to-icmp;" +
			"switch-to-lookup;no-keep-loops;hoist-common-insts;no-hoist-loads-stores-with-cond-faulting;" +
			"sink-common-insts;speculate-blocks;simplify-cond-branch;no-speculate-unpredictables>",
		"slp-vectorizer",
		"vector-combine",
		"instcombine<max-iterations=1;no-verify-fixpoint>",
		"loop-unroll<O3>",
		"transform-warning",
		"sroa<preserve-cfg>",
		"infer-alignment",
		"instcombine<max-iterations=1;no-verify-fixpoint>",
		"loop-mssa(licm<allowspeculation>)",
		"alignment-from-assumptions",
		"loop-sink",
		"instsimplify",
		"div-rem-pairs",
		"tailcallelim",
		"simplifycfg<bonus-inst-threshold=1;no-forward-switch-cond;switch-range-to-icmp;" +
			"no-switch-to-lookup;keep-loops;no-hoist-common-insts;" +
			"hoist-loads-stores-with-cond-faulting;no-sink-common-insts;speculate-blocks;" +
			"simplify-cond-branch;speculate-unpredictables>",
		"globaldce",
		"constmerge",
		"cg-profile",
		"rel-lookup-table-converter",
		"function(annotation-remarks)",
	}
}

// GetO3Int returns an int slice where each O3 pass is transformed to its integer representation.
func (pm *PassManager) GetO3Int() []int {
	return util.ForEach(getO3String(), pm.PassToInt)
}

// PassToInt returns the int mapped to by the string in the PassManager.
func (pm *PassManager) PassToInt(s string) int {
	return pm.direct[s]
}

// PassToInt returns the string mapped to by the int in the PassManager.
func (pm *PassManager) IntToPass(n int) string {
	return pm.inverse[n]
}

// makeValid transforms a list of ints, representing passes, into a valid opt pipeline.
func (pm *PassManager) makeValid(seq []int) string {
	const (
		INMODULE  byte   = iota
		INFUNC    byte   = iota
		INCGSCC   byte   = iota
		INLOOP    byte   = iota
		FUNCTEXT  string = "function<eager-inv>("
		CGSCCTEXT string = "cgscc(devirt<4>("
		LOOPTEXT  string = "function<eager-inv>(loop("
	)

	getClosing := func(scope byte) string {
		if scope == INFUNC {
			return ")"
		}
		return "))"
	}

	var buffer bytes.Buffer
	currentScope := INMODULE

	firstVal := seq[0]
	if pm.cgsccRange.Within(firstVal) {
		buffer.WriteString(CGSCCTEXT)
		currentScope = INCGSCC
	} else if pm.functionRange.Within(firstVal) {
		buffer.WriteString(FUNCTEXT)
		currentScope = INFUNC
	} else if pm.loopRange.Within(firstVal) {
		buffer.WriteString(LOOPTEXT)
		currentScope = INLOOP
	}
	buffer.WriteString(pm.IntToPass(firstVal))

	var next string
	for _, val := range seq[1:] {
		if pm.cgsccRange.Within(val) {
			next, currentScope = pm.nextAppend(val, currentScope, INMODULE, INCGSCC, CGSCCTEXT, getClosing)
		} else if pm.functionRange.Within(val) {
			next, currentScope = pm.nextAppend(val, currentScope, INMODULE, INFUNC, FUNCTEXT, getClosing)
		} else if pm.loopRange.Within(val) {
			next, currentScope = pm.nextAppend(val, currentScope, INMODULE, INLOOP, LOOPTEXT, getClosing)
		} else {
			next, currentScope = pm.nextAppend(val, currentScope, INMODULE, INMODULE, "", getClosing)
		}
		buffer.WriteString(next)
	}

	if currentScope != INMODULE {
		buffer.WriteString(getClosing(currentScope))
	}

	return buffer.String()
}

// nextAppend manages the logic for appending a pass based upon the previous pass.
func (pm *PassManager) nextAppend(pass int, currentScope, defaultScope, desiredScope byte,
	prefix string, getClosing func(byte) string) (string, byte) {

	retString := ""

	switch currentScope {
	// If we are in the same scope as the current pass
	case desiredScope:
		retString += ","
	// If we are in the default scope (module) and need to move into the current pass' scope
	case defaultScope:
		retString += "," + prefix
	// If we are in a nested scope, but have to move to a different nested scope
	default:
		retString += getClosing(currentScope) + "," + prefix
	}
	retString += pm.IntToPass(pass)
	currentScope = desiredScope

	return retString, currentScope
}
