package main

import (
	"fmt"
	"math"
	"math/big"

	"github.com/remyoudompheng/fptest"
)

const basePrec = 64

func main() {
	count := 0
	show := func(x float64) {
		mant, exp := math.Frexp(x)
		f := new(big.Float).SetMantExp(
			new(big.Float).SetInt64(int64(mant*(1<<54))+1),
			exp-54)
		count++
		fmt.Printf("count=%08d %dp%d %.18e midpoint=%.36e\n",
			count, int64(mant*(1<<53)), exp-53, x, f)
	}

	for digits := 16; digits > 0; digits-- {
		fmt.Println("===", digits, "digits ===")
		for exp := 60; exp < 1024-52; exp++ {
			fptest.AlmostDecimalPos(exp, digits, 53, uint(basePrec+2*digits), +1, show)
			fptest.AlmostDecimalPos(exp, digits, 53, uint(basePrec+2*digits), -1, show)
		}
		for exp := 60; exp < 1024+52; exp++ {
			if exp == 1023+52 {
				// denormals
				fptest.AlmostDecimalNeg(exp-1, digits, 52, uint(basePrec+2*digits), +1, true, show)
				fptest.AlmostDecimalNeg(exp-1, digits, 52, uint(basePrec+2*digits), -1, true, show)
			} else {
				fptest.AlmostDecimalNeg(exp, digits, 53, uint(basePrec+2*digits), +1, false, show)
				fptest.AlmostDecimalNeg(exp, digits, 53, uint(basePrec+2*digits), -1, false, show)
			}
		}
	}
}
