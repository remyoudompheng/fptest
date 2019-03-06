package main

import (
	"fmt"
	"math"
	"math/big"

	"github.com/remyoudompheng/fptest"
)

func main() {
	count := 0
	show := func(x float64) {
		mant, exp := math.Frexp(x)
		f := new(big.Float).SetMantExp(
			new(big.Float).SetInt64(int64(mant*(1<<54))+1),
			exp-54)
		count++
		fmt.Printf("count=%08d %dp%d %.18e midpoint=%.36e\n",
			int64(mant*(1<<53)), exp-53, count, x, f)
	}

	for digits := 16; digits > 0; digits-- {
		fmt.Println("===", digits, "digits ===")
		for exp := 60; exp < 1024-52; exp++ {
			fptest.AlmostDecimalPos(exp, digits, 53, uint(64+2*digits), +1, show)
			fptest.AlmostDecimalPos(exp, digits, 53, uint(64+2*digits), -1, show)
		}
	}
}
