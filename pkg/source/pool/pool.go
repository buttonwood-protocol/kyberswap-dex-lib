package pool

import (
	"math/big"
)

type Pool struct {
	Info PoolInfo
}

func (t *Pool) GetInfo() PoolInfo {
	return t.Info
}

func (t *Pool) GetTokens() []string {
	return t.Info.Tokens
}

// CanSwapTo is the base method to get all swappable tokens from a pool by a given token address
// Pools with custom logic should override this method
func (t *Pool) CanSwapTo(address string) []string {
	result := make([]string, 0, len(t.Info.Tokens))
	var tokenIndex = t.GetTokenIndex(address)
	if tokenIndex < 0 {
		return result
	}

	for i := 0; i < len(t.Info.Tokens); i += 1 {
		if i != tokenIndex {
			result = append(result, t.Info.Tokens[i])
		}
	}

	return result
}

func (t *Pool) GetAddress() string {
	return t.Info.Address
}

func (t *Pool) GetExchange() string {
	return t.Info.Exchange
}

func (t *Pool) Equals(other IPoolSimulator) bool {
	return t.GetAddress() == other.GetAddress()
}

func (t *Pool) GetTokenIndex(address string) int {
	return t.Info.GetTokenIndex(address)
}

func (t *Pool) GetType() string {
	return t.Info.Type
}

type CalcAmountOutResult struct {
	TokenAmountOut *TokenAmount
	Fee            *TokenAmount
	Gas            int64
	SwapInfo       interface{}
}

func (r *CalcAmountOutResult) IsValid() bool {
	return r.TokenAmountOut != nil && r.TokenAmountOut.Amount != nil && r.TokenAmountOut.Amount.Cmp(ZeroBI) > 0
}

type UpdateBalanceParams struct {
	TokenAmountIn  TokenAmount
	TokenAmountOut TokenAmount
	Fee            TokenAmount
	SwapInfo       interface{}
}

type PoolToken struct {
	Token               string
	Balance             *big.Int
	Weight              uint
	PrecisionMultiplier *big.Int
	VReserve            *big.Int
}

type PoolInfo struct {
	Address    string
	ReserveUsd float64
	SwapFee    *big.Int
	Exchange   string
	Type       string
	Tokens     []string
	Reserves   []*big.Int
	Checked    bool
}

func (t *PoolInfo) GetTokenIndex(address string) int {
	for i, poolToken := range t.Tokens {
		if poolToken == address {
			return i
		}
	}
	return -1
}
