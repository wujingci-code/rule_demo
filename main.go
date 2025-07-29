package main

import (
	"fmt"
	"log"
	"reflect"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/ext"
	"github.com/wujingci-code/rule_demo/cel_demo"
	"github.com/wujingci-code/rule_demo/expr_demo"
	"github.com/wujingci-code/rule_demo/perf_test"
)

type TestStruct struct {
	Field string
}

func main() {
	fmt.Println("Type name:", reflect.TypeOf(TestStruct{}).String())

	_, err := cel.NewEnv(
		ext.Strings(),
		ext.Lists(),
		cel.Types(&TestStruct{}),
		cel.Variable("test", cel.ObjectType(reflect.TypeOf(TestStruct{}).String())),
	)
	if err != nil {
		log.Fatalf("CEL 环境创建失败: %v", err)
	}
	expr_demo.Expr()
	cel_demo.EVM()
	cel_demo.UTXO()
	cel_demo.SOL()
	perf_test.CelPerfRun()
	perf_test.ExprPerfRun()
}
