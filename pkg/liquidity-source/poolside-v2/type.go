package poolsidev2

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type Gas struct {
	Swap int64
}

type RebaseTokenInfo struct {
	UnderlyingToken string   `json:"underlyingToken"`
	WrapRatio       *big.Int `json:"wrapRatio"`
	UnwrapRatio     *big.Int `json:"unwrapRatio"`
	Decimals        uint8    `json:"decimals"`
}

type PoolsListUpdaterMetadata struct {
	Offset int `json:"offset"`
}

type PairData struct {
	Token0 common.Address `json:"token0"`
	Token1 common.Address `json:"token1"`
	PlBps  uint16         `json:"plBps"`
	FeeBps uint16         `json:"feeBps"`
}

type Extra struct {
	PlBps              uint16                     `json:"plBps"`
	FeeBps             uint16                     `json:"feeBps"`
	RebaseTokenInfoMap map[string]RebaseTokenInfo `json:"rebaseTokenInfoMap"`
}
