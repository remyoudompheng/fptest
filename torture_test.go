package fptest

import (
	"bytes"
	"math"
	"strconv"
	"testing"

	"github.com/cespare/ryu"
)

func TestRyu(t *testing.T) {
	count := 0
	buf1 := make([]byte, 64)
	buf2 := make([]byte, 64)
	do := func(x float64) {
		s1 := ryu.AppendFloat64(buf1[:0], x)
		s2 := strconv.AppendFloat(buf2[:0], x, 'e', -1, 64)
		if !bytes.Equal(s1, s2) {
			t.Errorf("x=%v => %q %q", x, s1, s2)
		}

		y := math.Nextafter(x, 2*x)
		s1 = ryu.AppendFloat64(buf1[:0], y)
		s2 = strconv.AppendFloat(buf2[:0], y, 'e', -1, 64)
		if !bytes.Equal(s1, s2) {
			t.Errorf("x=%v => %q %q", y, s1, s2)
		}

		count += 2
	}

	const basePrec = 64
	for digits := 16; digits > 0; digits-- {
		for exp := 60; exp < 1024-52; exp++ {
			AlmostDecimalPos(exp, digits, 53, uint(basePrec+2*digits), +1, do)
			AlmostDecimalPos(exp, digits, 53, uint(basePrec+2*digits), -1, do)
		}
		for exp := 60; exp < 1024+52; exp++ {
			if exp == 1023+52 {
				// denormals
				AlmostDecimalNeg(exp-1, digits, 52, uint(basePrec+2*digits), +1, true, do)
				AlmostDecimalNeg(exp-1, digits, 52, uint(basePrec+2*digits), -1, true, do)
			} else {
				AlmostDecimalNeg(exp, digits, 53, uint(basePrec+2*digits), +1, false, do)
				AlmostDecimalNeg(exp, digits, 53, uint(basePrec+2*digits), -1, false, do)
			}
		}
	}

	t.Logf("%d numbers tested", count)
}
