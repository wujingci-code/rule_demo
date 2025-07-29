package main

import (
	"rule_demo/cel"
	"rule_demo/expr"
	"rule_demo/perf_test"
)

func main() {
	expr.Expr()
	cel.EVM()
	cel.UTXO()
	perf_test.CelPerfRun()
	perf_test.ExprPerfRun()
}
