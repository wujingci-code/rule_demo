package cel_demo

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"reflect"

	bin "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/ext"
)

type CELInstruction struct {
	ProgramAddress string
	Data           string
}

type CELMessage struct {
	AccountKeys  []string
	Instructions []CELInstruction
}

func decodeSolTx(rawTx string) CELMessage {
	decoded, err := base64.StdEncoding.DecodeString(rawTx)
	if err != nil {
		log.Fatalf("Raw TX 解码失败1: %v", err)
	}
	decodedTx, err := solana.TransactionFromDecoder(bin.NewBinDecoder(decoded))
	if err != nil {
		log.Fatalf("Raw TX 解码失败: %v", err)
	}
	msg := decodedTx.Message

	celMsg := CELMessage{
		AccountKeys:  make([]string, len(msg.AccountKeys)),
		Instructions: make([]CELInstruction, len(msg.Instructions)),
	}
	for i, pk := range msg.AccountKeys {
		celMsg.AccountKeys[i] = pk.String()
	}

	for i, ins := range msg.Instructions {
		celMsg.Instructions[i] = CELInstruction{
			ProgramAddress: msg.AccountKeys[ins.ProgramIDIndex].String(),
			Data:           hex.EncodeToString(ins.Data),
		}
	}
	return celMsg
}

func solRuleCheck(rawTx string) {
	policyMap := map[string][]string{
		"11111111111111111111111111111111":            {"02", "03"},
		"ComputeBudget111111111111111111111111111111": {"02", "03"},
		"TokenzQdBNbLqP5VEhdkAS6EPFLC1PHnBqCXEpPxuEb": {"01", "03", "09", "0c", "11"},
		"pytS9TjG1qyAZypk7n8rw8gfW9sUaqqYyMhJQ4E7JCQ": {"f8"},
	}
	celMsg := decodeSolTx(rawTx)
	env, err := cel.NewEnv(
		ext.Strings(),
		ext.Lists(),
		ext.NativeTypes(reflect.TypeOf(CELMessage{}), reflect.TypeOf(CELInstruction{})),
		cel.Variable("message", cel.ObjectType("cel_demo.CELMessage")),
		cel.Variable(
			"policyMap",
			cel.MapType(cel.StringType, cel.ListType(cel.StringType)),
		),
	)
	if err != nil {
		log.Fatalf("CEL 环境创建失败: %v", err)
	}

	rule := `
message.Instructions.all(inst,
  inst.ProgramAddress in policyMap &&
  policyMap[inst.ProgramAddress]
    .exists(pref, inst.Data.startsWith(pref))
)
`
	ast, issues := env.Compile(rule)
	if issues.Err() != nil {
		log.Fatalf("规则编译错误: %v", issues.Err())
	}
	prg, err := env.Program(ast)
	if err != nil {
		log.Fatalf("构建 CEL 程序失败: %v", err)
	}

	out, _, err := prg.Eval(map[string]any{
		"message":   celMsg,
		"policyMap": policyMap,
	})
	if err != nil {
		log.Fatalf("规则执行失败: %v", err)
	}

	passed, _ := out.Value().(bool)
	if passed {
		fmt.Println("✅ 所有 instruction 通过白名单校验 (使用原生结构体)")
	} else {
		fmt.Println("❌ 有 instruction 未通过校验 (使用原生结构体)")
	}
}

func SOL() {
	rawTx1 := "AQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACAAQAABiSDnX4XfCnWNCoMq5sXcPo4ngZ74Loa071mRtEKjZMgxdhgPrY5Um8wpfGx3Sj4RfRKPdfaNdy/wRcookEMph0Gm4hX/quBhPtof2NGGMA12sQ53BrrO1WYoPAAAAAAAQxKnsArVmgdpJsEupskz4n9gPks8ZHjfb0zb0bntxPZg1hok37PttsZZAWf8hrJzjjWLlYJ+ex3Wy9YP4uT3R0DBkZv5SEXMv/srbpyw5vnvIzlu8X3EmssQ5s6QAAAAK744dYUZu/YDAMF/f+6w1k2eTWT4JIsZc4lsXwAw2CLAgIGAAECBAcGC/jGnpHhdYcUcHGRBQAJA1DDAAAAAAAAAWHqO2dSshWyiFeVPciyEU6x4Jhr82nMxnFTEu5jtiIPAQkBBQ=="
	solRuleCheck(rawTx1)
	rawTx2 := "AQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACAAQAABiSDnX4XfCnWNCoMq5sXcPo4ngZ74Loa071mRtEKjZMgxdhgPrY5Um8wpfGx3Sj4RfRKPdfaNdy/wRcookEMph0Gm4hX/quBhPtof2NGGMA12sQ53BrrO1WYoPAAAAAAAQxKnsArVmgdpJsEupskz4n9gPks8ZHjfb0zb0bntxPZg1hok37PttsZZAWf8hrJzjjWLlYJ+ex3Wy9YP4uT3R0DBkZv5SEXMv/srbpyw5vnvIzlu8X3EmssQ5s6QAAAAK744dYUZu/YDAMF/f+6w1k2eTWT4JIsZc4lsXwAw2CLAgMGAAECBAcGC/jGnpHhdYcUcHGRBQAJA1DDAAAAAAAAAWHqO2dSshWyiFeVPciyEU6x4Jhr82nMxnFTEu5jtiIPAQkBBQ=="
	solRuleCheck(rawTx2)
}
