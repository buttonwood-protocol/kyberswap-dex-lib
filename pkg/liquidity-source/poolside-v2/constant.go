package poolsidev2

const (
	DexType = "poolside-v2"
)

var (
	defaultGas      = Gas{Swap: 60000}
	defaultDecimals = uint8(18)
)

const (
	factoryMethodGetPair        = "allPairs"
	factoryMethodAllPairsLength = "allPairsLength"
	pairMethodToken0            = "token0"
	pairMethodToken1            = "token1"
	pairMethodFeeBps            = "feeBps"
	pairMethodPlBps             = "plBps"
	erc20TokenDecimals          = "decimals"
)
