package main

import (
	"github.com/wujingci-code/rule_demo/cel_demo"
	"github.com/wujingci-code/rule_demo/expr_demo"
	"github.com/wujingci-code/rule_demo/perf_test"
)

func main() {

	expr_demo.Expr()
	cel_demo.EVM()
	cel_demo.UTXO()
	cel_demo.SOL()
	perf_test.CelPerfRun()
	perf_test.ExprPerfRun()
}
