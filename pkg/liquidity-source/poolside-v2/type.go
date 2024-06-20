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

type GetReservesResult struct {
	Pool0              *big.Int
	Pool1              *big.Int
	Reservoir0         *big.Int
	Reservoir1         *big.Int
	Basin0             *big.Int
	Basin1             *big.Int
	BlockTimestampLast uint32
}

type ReserveData struct {
	Pool0 *big.Int
	Pool1 *big.Int
}

type PoolMeta struct {
	PlBps       uint16 `json:"plBps"`
	FeeBps      uint16 `json:"feeBps"`
	BlockNumber uint64 `json:"blockNumber"`
}
