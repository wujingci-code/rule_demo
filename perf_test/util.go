package perf_test

import (
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

type EvmTx struct {
	ChainId string `json:"chainId"`
	Data    string `json:"data"`
	From    string `json:"from"`
	To      string `json:"to"`
}

func txDecode(rawTx string) EvmTx {
	b, err := hexutil.Decode(rawTx)
	if err != nil {
		log.Fatalf("invalid tx hex: %v", err)
	}
	var tx etypes.Transaction
	if err := rlp.DecodeBytes(b, &tx); err != nil {
		log.Fatalf("RLP decode error: %v", err)
	}
	cid := big.NewInt(0)
	if tx.ChainId() != nil {
		cid = tx.ChainId()
	}
	signer := etypes.LatestSignerForChainID(cid)
	from, err := etypes.Sender(signer, &tx)
	if err != nil {
		log.Fatalf("recover sender error: %v", err)
	}
	return EvmTx{
		ChainId: fmt.Sprintf("0x%s", cid.Text(16)),
		To:      tx.To().Hex(),
		From:    from.Hex(),
		Data:    hexutil.Encode(tx.Data()),
	}
}
