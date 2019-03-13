package fptest

import (
	"bytes"
	"math"
	"strconv"
	"testing"

	"github.com/cespare/ryu"
)

func TestTortureShortest64(t *testing.T) {
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

	basePrec := 64
	if testing.Short() {
		basePrec += 4
	}
	for digits := 16; digits > 0; digits-- {
		for exp := 60; exp < 1024-52; exp++ {
			AlmostDecimalMidpoint(exp, digits, 53, uint(basePrec+2*digits), +1, false, do)
			AlmostDecimalMidpoint(exp, digits, 53, uint(basePrec+2*digits), -1, false, do)
		}
		for exp := 60; exp < 1024+52; exp++ {
			if exp == 1023+52 {
				// denormals
				AlmostDecimalMidpoint(-(exp - 1), digits, 52, uint(basePrec+2*digits), +1, true, do)
				AlmostDecimalMidpoint(-(exp - 1), digits, 52, uint(basePrec+2*digits), -1, true, do)
			} else {
				AlmostDecimalMidpoint(-exp, digits, 53, uint(basePrec+2*digits), +1, false, do)
				AlmostDecimalMidpoint(-exp, digits, 53, uint(basePrec+2*digits), -1, false, do)
			}
		}
	}

	t.Logf("%d numbers tested", count)
}

func TestTortureShortest32(t *testing.T) {
	count := 0
	buf1 := make([]byte, 32)
	buf2 := make([]byte, 32)
	do := func(x float64) {
		s1 := ryu.AppendFloat32(buf1[:0], float32(x))
		s2 := strconv.AppendFloat(buf2[:0], x, 'e', -1, 32)
		if !bytes.Equal(s1, s2) {
			t.Errorf("x=%v => %q %q", x, s1, s2)
		}
		//t.Logf("x=%v => %q", x, s1)

		y := math.Nextafter32(float32(x), 2*float32(x))
		s1 = ryu.AppendFloat32(buf1[:0], float32(y))
		s2 = strconv.AppendFloat(buf2[:0], float64(y), 'e', -1, 32)
		if !bytes.Equal(s1, s2) {
			t.Errorf("x=%v => %q %q", y, s1, s2)
		}
		//t.Logf("y=%v => %q", y, s1)

		count += 2
	}

	basePrec := 24
	for digits := 7; digits > 0; digits-- {
		for exp := 10; exp < 127-23; exp++ {
			AlmostDecimalMidpoint(exp, digits, 24, uint(basePrec+2*digits), +1, false, do)
			AlmostDecimalMidpoint(exp, digits, 24, uint(basePrec+2*digits), -1, false, do)
		}
		for exp := 60; exp < 127+23; exp++ {
			if exp == 126+23 {
				// denormals
				AlmostDecimalMidpoint(-(exp - 1), digits, 23, uint(basePrec+2*digits), +1, true, do)
				AlmostDecimalMidpoint(-(exp - 1), digits, 23, uint(basePrec+2*digits), -1, true, do)
			} else {
				AlmostDecimalMidpoint(-exp, digits, 24, uint(basePrec+2*digits), +1, false, do)
				AlmostDecimalMidpoint(-exp, digits, 24, uint(basePrec+2*digits), -1, false, do)
			}
		}
	}

	t.Logf("%d numbers tested", count)
}

func TestTortureFixed64(t *testing.T) {
	basePrec := 64
	if testing.Short() {
		basePrec += 4
	}

	count := 0
	tooshort := 0
	errors := 0
	buf1 := make([]byte, 64)
	buf2 := make([]byte, 64)
	for digits := 16; digits > 0; digits-- {
		roundUp := false
		do := func(x float64, n uint64, k int) {
			// x ~= (n + 1/2) × 10^k
			count++
			s1 := strconv.AppendFloat(buf1[:0], x, 'e', digits-1, 64)
			// Round up.
			if roundUp {
				n += 1
			}
			s2, ok := appendE(buf2[:0], n, digits, k)
			if !ok {
				//t.Logf("skip %.32e ~= %d.5e%d", x, n, k)
				tooshort++
				return
			}
			if !bytes.Equal(s1, s2) {
				t.Errorf("x=%.32e digits=%d => %q %q ERR", x, digits, s1, s2)
				errors++
			} else {
				//t.Logf("x=%.32e digits=%d => %q %q OK\n", x, digits, s1, s2)
			}
		}

		for exp := 40; exp < 1024-52; exp++ {
			prec := basePrec
			if exp < 64 {
				prec -= 4
			}
			roundUp = true
			AlmostHalfDecimal(exp, digits, 53, uint(prec+2*digits), +1, false, do)
			roundUp = false
			AlmostHalfDecimal(exp, digits, 53, uint(prec+2*digits), -1, false, do)
		}
		for exp := 50; exp < 1024+52; exp++ {
			prec := basePrec
			if exp < 64 {
				prec -= 4
			}
			if exp == 1023+52 {
				// denormals
				roundUp = true
				AlmostHalfDecimal(-(exp - 1), digits, 52, uint(prec+2*digits), +1, true, do)
				roundUp = false
				AlmostHalfDecimal(-(exp - 1), digits, 52, uint(prec+2*digits), -1, true, do)
			} else {
				roundUp = true
				AlmostHalfDecimal(-exp, digits, 53, uint(prec+2*digits), +1, false, do)
				roundUp = false
				AlmostHalfDecimal(-exp, digits, 53, uint(prec+2*digits), -1, false, do)
			}
		}
	}

	t.Logf("%d numbers tested, %d errors, %d skipped (too few digits)", count, errors, tooshort)
}

func TestTortureFixed32(t *testing.T) {
	prec := 24

	count := 0
	tooshort := 0
	errors := 0
	buf1 := make([]byte, 32)
	buf2 := make([]byte, 32)
	for digits := 7; digits > 0; digits-- {
		roundUp := false
		do := func(x float64, n uint64, k int) {
			// x ~= (n + 1/2) × 10^k
			count++
			s1 := strconv.AppendFloat(buf1[:0], x, 'e', digits-1, 32)
			// Round up.
			if roundUp {
				n += 1
			}
			// If n is a power of ten, the number of digits changes
			// and the test is invalid.
			switch n {
			case 1, 10, 100, 1000:
				tooshort++
				return
			}
			s2, ok := appendE(buf2[:0], n, digits, k)
			if !ok {
				//t.Logf("skip %.32e ~= %d.5e%d", x, n, k)
				tooshort++
				return
			}
			if !bytes.Equal(s1, s2) {
				t.Errorf("x=%.32e digits=%d => %q %q ERR", x, digits, s1, s2)
				errors++
			} else {
				//t.Logf("x=%.32e digits=%d => %q %q OK\n", x, digits, s1, s2)
			}
		}

		for exp := 10; exp < 127-23; exp++ {
			roundUp = true
			AlmostHalfDecimal(exp, digits, 24, uint(prec+2*digits), +1, false, do)
			roundUp = false
			AlmostHalfDecimal(exp, digits, 24, uint(prec+2*digits), -1, false, do)
		}
		for exp := 10; exp < 127+23; exp++ {
			if exp == 126+23 {
				// denormals
				roundUp = true
				AlmostHalfDecimal(-(exp - 1), digits, 23, uint(prec+2*digits), +1, true, do)
				roundUp = false
				AlmostHalfDecimal(-(exp - 1), digits, 23, uint(prec+2*digits), -1, true, do)
			} else {
				roundUp = true
				AlmostHalfDecimal(-exp, digits, 24, uint(prec+2*digits), +1, false, do)
				roundUp = false
				AlmostHalfDecimal(-exp, digits, 24, uint(prec+2*digits), -1, false, do)
			}
		}
	}

	t.Logf("%d numbers tested, %d errors, %d skipped (too few digits)", count, errors, tooshort)
}

func appendE(s []byte, mant uint64, digits int, exp int) ([]byte, bool) {
	exp += digits - 1
	s2 := strconv.AppendUint(s[0:1], mant, 10)
	if len(s2) != digits+1 {
		return nil, false
	}
	s2[0], s2[1] = s2[1], '.'
	if digits == 1 {
		s2 = append(s[:0], byte(mant+'0'))
	}
	s2 = append(s2, 'e')
	if exp >= 0 {
		s2 = append(s2, '+')
	} else {
		exp = -exp
		s2 = append(s2, '-')
	}
	if exp < 10 {
		s2 = append(s2, '0')
	}
	s2 = strconv.AppendInt(s2, int64(exp), 10)
	return s2, true
}
