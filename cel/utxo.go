package cel

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"
	"reflect"

	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil/psbt"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

type Packet struct {
	UnsignedTx *wire.MsgTx
	Inputs     []PInput
	Outputs    []POutput
	Unknowns   []Unknown
}
type PInput struct{ Value int64 }
type POutput struct{ Value int64 }
type Unknown struct{}

func DecodePSBTHex(hexStr string) (*Packet, error) {
	raw, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, err
	}
	p, err := psbt.NewFromRawBytes(bytes.NewReader(raw), false)
	if err != nil {
		return nil, err
	}
	pkt := &Packet{
		UnsignedTx: p.UnsignedTx,
		Inputs:     make([]PInput, len(p.Inputs)),
		Outputs:    make([]POutput, len(p.UnsignedTx.TxOut)),
	}
	for i, in := range p.Inputs {
		var val int64
		if in.WitnessUtxo != nil {
			val = in.WitnessUtxo.Value
		} else if in.NonWitnessUtxo != nil {
			idx := p.UnsignedTx.TxIn[i].PreviousOutPoint.Index
			val = in.NonWitnessUtxo.TxOut[idx].Value
		} else {
			return nil, fmt.Errorf("input %d: missing UTXO info", i)
		}
		pkt.Inputs[i] = PInput{Value: val}
	}
	for i, txOut := range p.UnsignedTx.TxOut {
		pkt.Outputs[i] = POutput{Value: txOut.Value}
	}
	return pkt, nil
}

func sumFn(args ...ref.Val) ref.Val {
	if len(args) != 1 {
		return types.NewErr("sum: 需要1个参数，列表或数组")
	}
	raw := args[0].Value()
	rv := reflect.ValueOf(raw)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return types.NewErr("sum: 参数必须是列表或数组")
	}
	var sum float64
	for i := 0; i < rv.Len(); i++ {
		elem := rv.Index(i).Interface()
		switch v := elem.(type) {
		case int:
			sum += float64(v)
		case int8:
			sum += float64(v)
		case int16:
			sum += float64(v)
		case int32:
			sum += float64(v)
		case int64:
			sum += float64(v)
		case uint:
			sum += float64(v)
		case uint8:
			sum += float64(v)
		case uint16:
			sum += float64(v)
		case uint32:
			sum += float64(v)
		case uint64:
			sum += float64(v)
		case float32:
			sum += float64(v)
		case float64:
			sum += v
		default:
			ev := reflect.ValueOf(elem)
			if ev.Kind() == reflect.Struct {
				fld := ev.FieldByName("Value")
				if fld.IsValid() && fld.CanInt() {
					sum += float64(fld.Int())
					continue
				}
			}
			return types.NewErr("sum: 列表元素类型不支持")
		}
	}
	return types.Double(sum)
}

func UTXO() {
	hexStr := "70736274ff0100a00200000002ab0949a08c5af7c49b8212f417e2f15ab3f5c33dcf153821a8139f877a5b7be40000000000feffffffab0949a08c5af7c49b8212f417e2f15ab3f5c33dcf153821a8139f877a5b7be40100000000feffffff02603bea0b000000001976a914768a40bbd740cbe81d988e71de2a4d5c71396b1d88ac8e240000000000001976a9146f4620b553fa095e721b9ee0efe9fa039cca459788ac00000000000100df0200000001268171371edff285e937adeea4b37b78000c0566cbb3ad64641713ca42171bf6000000006a473044022070b2245123e6bf474d60c5b50c043d4c691a5d2435f09a34a7662a9dc251790a022001329ca9dacf280bdf30740ec0390422422c81cb45839457aeb76fc12edd95b3012102657d118d3357b8e0f4c2cd46db7b39f6d9c38d9a70abcb9b2de5dc8dbfe4ce31feffffff02d3dff505000000001976a914d0c59903c5bac2868760e90fd521a4665aa7652088ac00e1f5050000000017a9143545e6e33b832c47050f24d3eeb93c9c03948bc787b32e13000001012000e1f5050000000017a9143545e6e33b832c47050f24d3eeb93c9c03948bc787010416001485d13537f2e265405a34dbafa9e3dda01fb8230800220202ead596687ca806043edc3de116cdf29d5e9257c196cd055cf698c8d02bf24e9910b4a6ba670000008000000080020000800022020394f62be9df19952c5587768aeb7698061ad2c4a25c894f47d8c162b4d7213d0510b4a6ba6700000080010000800200008000" // 填你的 PSBT hex
	pkt, err := DecodePSBTHex(hexStr)
	if err != nil {
		log.Fatalf("PSBT 解码失败: %v", err)
	}
	pktMap := structToMap(pkt)

	env, err := cel.NewEnv(
		cel.Variable("packet", cel.MapType(cel.StringType, cel.DynType)),
		cel.Function("sum",
			cel.Overload("sum_list_dyn",
				[]*cel.Type{cel.ListType(cel.DynType)},
				cel.DoubleType,
				cel.FunctionBinding(sumFn),
			),
		),
	)
	if err != nil {
		log.Fatalf("CEL 环境创建失败: %v", err)
	}

	// 3. 写规则、编译、执行
	cond := `
      sum(packet.Inputs)/100000000.0 <= 10.0 &&
      (sum(packet.Inputs)/100000000.0 - sum(packet.Outputs)/100000000.0) <= 0.1
    `
	ast, iss := env.Compile(cond)
	if iss.Err() != nil {
		log.Fatalf("规则编译错误: %v", iss.Err())
	}
	prg, err := env.Program(ast)
	if err != nil {
		log.Fatalf("CEL 程序构建失败: %v", err)
	}
	out, _, err := prg.Eval(map[string]interface{}{
		"packet": pktMap,
	})
	if err != nil {
		log.Fatalf("规则执行失败: %v", err)
	}
	pass, ok := out.Value().(bool)
	if !ok {
		log.Fatalf("期望 bool 返回值，实际是 %T", out.Value())
	}

	if pass {
		log.Println("ALLOW")
	} else {
		log.Println("FORBID")
	}
}
