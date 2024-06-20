package poolsidev2

import (
	poolpkg "github.com/KyberNetwork/kyberswap-dex-lib/pkg/source/pool"
)

type PoolSimulator struct {
	poolpkg.Pool

	gas Gas

	rebaseTokenInfoMap map[string]RebaseTokenInfo
}
