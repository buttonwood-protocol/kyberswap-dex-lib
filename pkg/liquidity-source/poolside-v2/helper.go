package poolsidev2

import "github.com/holiman/uint256"

func kN(poolA, poolB, plBps *uint256.Int) *uint256.Int {
	defer func() {
		if r := recover(); r != nil {
			panic(r)
		}
	}()

	t1 := SafeMul(plBps, SafeAdd(poolA, poolB))
	t2 := SafeMul(uint256.NewInt(4), SafeMul(plBps, SafeMul(poolA, poolB)))
	t3 := SafeSub(uint256.NewInt(10000), plBps)

	t1Mul50 := SafeMul(t1, uint256.NewInt(50))
	sqrtVal := new(uint256.Int).Sqrt(SafeMul(SafeAdd(SafeMul(t1, t1), SafeMul(t2, t3)), uint256.NewInt(2500)))

	return SafeAdd(t1Mul50, sqrtVal)
}

func kDMult(plBps, scalar *uint256.Int) *uint256.Int {
	defer func() {
		if r := recover(); r != nil {
			panic(r)
		}
	}()

	return SafeMul(
		new(uint256.Int).Sqrt(
			SafeMul(plBps, SafeMul(SafeSub(uint256.NewInt(10000), plBps), SafeSub(uint256.NewInt(10000), plBps))),
		),
		scalar,
	)
}
