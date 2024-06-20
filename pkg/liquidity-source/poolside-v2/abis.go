package poolsidev2

import (
	"bytes"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

var (
	poolsideV2PairABI    abi.ABI
	poolsideV2FactoryABI abi.ABI
)

func init() {
	builder := []struct {
		ABI  *abi.ABI
		data []byte
	}{
		{
			&poolsideV2PairABI, pairV2ABIJson,
		},
		{
			&poolsideV2FactoryABI, factoryV2ABIJson,
		},
	}

	for _, b := range builder {
		var err error
		*b.ABI, err = abi.JSON(bytes.NewReader(b.data))
		if err != nil {
			panic(err)
		}
	}
}
