package main

import (
	"fmt"
	"math"
	"math/big"

	"github.com/remyoudompheng/fptest"
)

const basePrec = 64

func main() {
	hardMidpoints()
	hardFloats()
}

func hardMidpoints() {
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
			fptest.AlmostDecimalMidpoint(exp, digits, 53, uint(basePrec+2*digits), +1, false, show)
			fptest.AlmostDecimalMidpoint(exp, digits, 53, uint(basePrec+2*digits), -1, false, show)
		}
		for exp := 60; exp < 1024+52; exp++ {
			if exp == 1023+52 {
				// denormals
				fptest.AlmostDecimalMidpoint(exp-1, digits, 52, uint(basePrec+2*digits), +1, true, show)
				fptest.AlmostDecimalMidpoint(exp-1, digits, 52, uint(basePrec+2*digits), -1, true, show)
			} else {
				fptest.AlmostDecimalMidpoint(exp, digits, 53, uint(basePrec+2*digits), +1, false, show)
				fptest.AlmostDecimalMidpoint(exp, digits, 53, uint(basePrec+2*digits), -1, false, show)
			}
		}
	}
}

// These numbers are hard to round correctly (down or up?).
func hardFloats() {
	count := 0

	for digits := 19; digits > 0; digits-- {
		show := func(x float64) {
			mant, exp := math.Frexp(x)
			count++
			D := fmt.Sprint(digits - 1)
			fmt.Printf("count=%08d %dp%d %."+D+"e %.36e\n",
				count, int64(mant*(1<<53)), exp-53, x, x)
		}

		fmt.Println("===", digits, "digits ===")
		for exp := 45 + digits; exp < 1024-52; exp++ {
			fptest.AlmostHalfDecimal(exp, digits, 53, uint(basePrec+2*digits), +1, false, show)
			fptest.AlmostHalfDecimal(exp, digits, 53, uint(basePrec+2*digits), -1, false, show)
		}
		for exp := 60; exp < 1024+52; exp++ {
			if exp == 1023+52 {
				// denormals
				fptest.AlmostHalfDecimal(exp-1, digits, 52, uint(basePrec+2*digits), +1, true, show)
				fptest.AlmostHalfDecimal(exp-1, digits, 52, uint(basePrec+2*digits), -1, true, show)
			} else {
				fptest.AlmostHalfDecimal(exp, digits, 53, uint(basePrec+2*digits), +1, false, show)
				fptest.AlmostHalfDecimal(exp, digits, 53, uint(basePrec+2*digits), -1, false, show)
			}
		}
	}
}
