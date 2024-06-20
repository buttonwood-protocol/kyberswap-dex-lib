package poolsidev2

import _ "embed"

//go:embed abis/ButtonswapV2Pair.json
var pairV2ABIJson []byte

//go:embed abis/ButtonswapV2Factory.json
var factoryV2ABIJson []byte

//go:embed abis/ERC20.json
var erc20ABIJson []byte

//go:embed abis/ButtonToken.json
var buttonTokenABIJson []byte
