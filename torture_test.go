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
	roundUp := false
	checkShortest := false
	do := func(x float64) {
		s1 := ryu.AppendFloat64(buf1[:0], x)
		s2 := strconv.AppendFloat(buf2[:0], x, 'e', -1, 64)
		if !bytes.Equal(s1, s2) {
			t.Errorf("x=%v => %q %q", x, s1, s2)
		}

		y := math.Nextafter(x, 2*x)
		t1 := ryu.AppendFloat64(buf1[:0], y)
		t2 := strconv.AppendFloat(buf2[:0], y, 'e', -1, 64)
		if !bytes.Equal(t1, t2) {
			t.Errorf("x=%v => %q %q", y, t1, t2)
		}

		if checkShortest {
			u := strconv.AppendFloat(buf1[:0], x, 'e', -1, 64)
			v := strconv.AppendFloat(buf2[:0], y, 'e', -1, 64)
			if roundUp && len(u) < len(v) {
				t.Errorf("expected %.30e longer than %.30e, got %s %s",
					x, y, u, v)
			}
			if !roundUp && len(u) > len(v) {
				t.Errorf("expected %.30e shorter than %.30e, got %s %s",
					x, y, u, v)
			}
		}

		count += 2
	}

	for digits := 18; digits > 0; digits-- {
		checkShortest = (digits <= 15)
		difficulty := 48 + 3*digits
		if testing.Short() {
			difficulty += 4
		}
		if difficulty < 64 {
			difficulty = 64
		}
		count = 0
		for exp := 55; exp < 1024-52; exp++ {
			roundUp = false
			AlmostDecimalMidpoint(exp, digits, 53, uint(difficulty), +1, false, do)
			roundUp = true
			AlmostDecimalMidpoint(exp, digits, 53, uint(difficulty), -1, false, do)
		}
		for exp := 55; exp < 1024+52; exp++ {
			if exp == 1023+52 {
				// denormals
				roundUp = false
				AlmostDecimalMidpoint(-(exp - 1), digits, 52, uint(difficulty), +1, true, do)
				roundUp = true
				AlmostDecimalMidpoint(-(exp - 1), digits, 52, uint(difficulty), -1, true, do)
			} else {
				roundUp = false
				AlmostDecimalMidpoint(-exp, digits, 53, uint(difficulty), +1, false, do)
				roundUp = true
				AlmostDecimalMidpoint(-exp, digits, 53, uint(difficulty), -1, false, do)
			}
		}
		t.Logf("%d numbers tested (%d decimal digits)", count, digits)
	}
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
	for digits := 10; digits > 0; digits-- {
		count = 0
		for exp := 10; exp < 127-23; exp++ {
			AlmostDecimalMidpoint(exp, digits, 24, uint(basePrec+2*digits), +1, false, do)
			AlmostDecimalMidpoint(exp, digits, 24, uint(basePrec+2*digits), -1, false, do)
		}
		for exp := 10; exp < 127+23; exp++ {
			if exp == 126+23 {
				// denormals
				AlmostDecimalMidpoint(-(exp - 1), digits, 23, uint(basePrec+2*digits), +1, true, do)
				AlmostDecimalMidpoint(-(exp - 1), digits, 23, uint(basePrec+2*digits), -1, true, do)
			} else {
				AlmostDecimalMidpoint(-exp, digits, 24, uint(basePrec+2*digits), +1, false, do)
				AlmostDecimalMidpoint(-exp, digits, 24, uint(basePrec+2*digits), -1, false, do)
			}
		}
		t.Logf("%d digits: %d numbers tested", digits, count)
	}
}

func TestTortureFixed64(t *testing.T) {
	buf1 := make([]byte, 64)
	buf2 := make([]byte, 64)
	for digits := 18; digits > 0; digits-- {
		count := 0
		tooshort := 0
		errors := 0

		roundUp := false
		do := func(x float64, n uint64, k int) {
			// x ~= (n + 1/2) × 10^k
			count++
			s1 := strconv.AppendFloat(buf1[:0], x, 'e', digits-1, 64)
			// Round up.
			if roundUp {
				n += 1
			}
			// If n is a power of ten, the number of digits changes
			// and the test is invalid.
			switch n {
			case 1, 10, 100, 1000, 10000:
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
			} else if false {
				t.Logf("x=%.32e digits=%d => %q %q OK", x, digits, s1, s2)
			}
		}

		difficulty := 48 + 3*digits
		if testing.Short() {
			difficulty += 4
		}
		if difficulty < 64 {
			difficulty = 64
		}

		for exp := 40; exp < 1024-52; exp++ {
			roundUp = true
			AlmostHalfDecimal(exp, digits, 53, uint(difficulty), +1, false, do)
			roundUp = false
			AlmostHalfDecimal(exp, digits, 53, uint(difficulty), -1, false, do)
		}
		for exp := 50; exp < 1024+52; exp++ {
			if exp == 1023+52 {
				// denormals
				roundUp = true
				AlmostHalfDecimal(-(exp - 1), digits, 52, uint(difficulty), +1, true, do)
				roundUp = false
				AlmostHalfDecimal(-(exp - 1), digits, 52, uint(difficulty), -1, true, do)
			} else {
				roundUp = true
				AlmostHalfDecimal(-exp, digits, 53, uint(difficulty), +1, false, do)
				roundUp = false
				AlmostHalfDecimal(-exp, digits, 53, uint(difficulty), -1, false, do)
			}
		}

		t.Logf("%d digits: %d numbers tested, %d errors, %d skipped (too few digits)",
			digits, count, errors, tooshort)
	}
}

func TestTortureFixed32(t *testing.T) {
	prec := 24

	buf1 := make([]byte, 32)
	buf2 := make([]byte, 32)
	for digits := 10; digits > 0; digits-- {
		count := 0
		tooshort := 0
		errors := 0

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

		t.Logf("%d digits: %d numbers tested, %d errors, %d skipped (too few digits)",
			digits, count, errors, tooshort)
	}
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
