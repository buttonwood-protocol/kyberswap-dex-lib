package poolsidev2

import (
	"errors"

	"github.com/holiman/uint256"
)

var (
	ErrDSMathAddOverflow  = errors.New("ds-math-add-overflow")
	ErrDSMathSubUnderflow = errors.New("ds-math-sub-underflow")
	ErrDSMathMulOverflow  = errors.New("ds-math-mul-overflow")
	ErrDivisionByZero     = errors.New("division-by-zero")
)

func SafeAdd(x, y *uint256.Int) *uint256.Int {
	z := new(uint256.Int).Add(x, y)
	if z.Cmp(x) >= 0 {
		return z
	}

	panic(ErrDSMathAddOverflow)
}

func SafeSub(x, y *uint256.Int) *uint256.Int {
	z := new(uint256.Int).Sub(x, y)
	if z.Cmp(x) <= 0 {
		return z
	}

	panic(ErrDSMathSubUnderflow)
}

func SafeMul(x, y *uint256.Int) *uint256.Int {
	z := new(uint256.Int).Mul(x, y)
	if y.CmpUint64(0) == 0 || new(uint256.Int).Div(z, y).Cmp(x) == 0 {
		return z
	}

	panic(ErrDSMathMulOverflow)
}

func SafeDiv(x, y *uint256.Int) *uint256.Int {
	if y.IsZero() {
		panic("division by zero")
	}
	return new(uint256.Int).Div(x, y)
}
