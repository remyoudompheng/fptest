package fptest

import (
	"bytes"
	"math"
	"strconv"
	"testing"
)

// These tests enumerate all possible corner cases for atof/ftoa
// algorithm, except for small positive/negative exponents
// where the corner cases are not longer hard to find by fuzz testing.
//
// A corner case is a number which is close to a binary or decimal
// midpoint with a relative difference less than 1/2^difficulty.

func TestTortureShortest64(t *testing.T) {
	// The corner cases for ftoa are such that:
	// - they are close to a short decimal number D
	// - the short decimal number itself is very close to x±ulp/2
	//   so that it is difficult to decide whether D is a valid
	//   representation of x (ParseFloat(D) == x or x±ulp)
	//
	// Cases where D == x±ulp/2 are not emitted by the generator.
	// They may happen for small values of the exponent, e.g.
	// - M p+N where M is divisible by 5^N (only for N <= 23)
	// - M p-N == M*5^N / 10^N can have <= 19 digits for N <= 4
	count := 0
	buf0 := make([]byte, 64)
	buf1 := make([]byte, 64)
	buf2 := make([]byte, 64)
	roundUp := false
	do := func(x float64, n uint64, k int) {
		// prepare the shortest representation of n*10^k
		// if n is short enough:
		// if roundUp is true, this is the result for nextfloat(x)
		// if roundUp is false, this is the result for x
		// otherwise, we will test shortest property below.
		s0, _ := appendE(buf0[:0], n, -1, k)
		y := math.Nextafter(x, 2*x)
		s1 := strconv.AppendFloat(buf1[:0], x, 'e', -1, 64)
		s2 := strconv.AppendFloat(buf2[:0], y, 'e', -1, 64)
		if n < 1e15 && !roundUp && !bytes.Equal(s0, s1) {
			t.Errorf("x=%.20e => %q want %q", x, s1, s0)
		}
		if n < 1e15 && roundUp && !bytes.Equal(s0, s2) {
			t.Errorf("y=%.20e => %q want %q", y, s2, s0)
		}
		if count%200000 == 0 || (testing.Short() && count%20000 == 0) {
			t.Logf("x=%.20e < %s < y=%.20e", x, s0, y)
			t.Logf("x => %s", s1)
			t.Logf("y => %s", s2)
		}
		// Also the representation must be correct and the shortest.
		// Test round trip.
		if z, _ := strconv.ParseFloat(string(s1), 64); z != x {
			t.Errorf("x=%v => %q => %v incorrect round trip", x, s1, z)
		}
		if z, _ := strconv.ParseFloat(string(s2), 64); z != y {
			t.Errorf("y=%v => %q => %v incorrect round trip", y, s2, z)
		}
		// Test shortest
		prec := countDigits(s1)
		s0 = strconv.AppendFloat(buf0[:0], x, 'e', prec-1, 64)
		if !bytes.Equal(s0, s1) {
			t.Errorf("x=%v => %q != %q", x, s0, s1)
		}
		if prec > 1 {
			s0 := strconv.AppendFloat(buf0[:0], x, 'e', prec-2, 64)
			if z, _ := strconv.ParseFloat(string(s0), 64); z == x {
				t.Errorf("x=%v => %q not shortest", x, s1)
			}

		}
		prec = countDigits(s2)
		s0 = strconv.AppendFloat(buf0[:0], y, 'e', prec-1, 64)
		if !bytes.Equal(s0, s2) {
			t.Errorf("y=%v => %q != %q", y, s0, s2)
		}
		if prec > 1 {
			s0 := strconv.AppendFloat(buf0[:0], y, 'e', prec-2, 64)
			if z, _ := strconv.ParseFloat(string(s0), 64); z == y {
				t.Errorf("x=%v => %q not shortest", y, s2)
			}

		}
		count += 2
	}

	for digits := 18; digits > 0; digits-- {
		difficulty := 48 + 3*digits
		if testing.Short() {
			difficulty += 4
		}
		if difficulty < 64 {
			difficulty = 64
		}
		count = 0
		for exp := 0; exp < 1024-52; exp++ {
			roundUp = false
			AlmostDecimalMidpoint(exp, digits, 53, uint(difficulty), +1, false, do)
			roundUp = true
			AlmostDecimalMidpoint(exp, digits, 53, uint(difficulty), -1, false, do)
		}
		for exp := 1; exp < 1024+52; exp++ {
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
	buf0 := make([]byte, 32)
	buf1 := make([]byte, 32)
	buf2 := make([]byte, 32)
	roundUp := false
	do := func(x float64, n uint64, k int) {
		// prepare the shortest representation of n*10^k
		// if n is short enough:
		// if roundUp is true, this is the result for nextfloat32(x)
		// if roundUp is false, this is the result for x
		// otherwise, we will test shortest property below.
		s0, _ := appendE(buf0[:0], n, -1, k)
		y := float64(math.Nextafter32(float32(x), 2*float32(x)))
		s1 := strconv.AppendFloat(buf1[:0], x, 'e', -1, 32)
		s2 := strconv.AppendFloat(buf2[:0], y, 'e', -1, 32)
		if n < 1e6 && !roundUp && !bytes.Equal(s0, s1) {
			t.Errorf("x=%.10e => %q want %q", float32(x), s1, s0)
		}
		if n < 1e6 && roundUp && !bytes.Equal(s0, s2) {
			t.Errorf("y=%.10e => %q want %q", float32(y), s2, s0)
		}
		if count%20000 == 0 || (testing.Short() && count%10000 == 0) {
			t.Logf("x=%.10e < %s < y=%.10e", float32(x), s0, float32(y))
			t.Logf("x => %s", s1)
			t.Logf("y => %s", s2)
		}
		// Also the representation must be correct and the shortest.
		// Test round trip.
		if z, _ := strconv.ParseFloat(string(s1), 32); z != x {
			t.Errorf("x=%v => %q => %v incorrect round trip", float32(x), s1, z)
		}
		if z, _ := strconv.ParseFloat(string(s2), 32); z != y {
			t.Errorf("y=%v => %q => %v incorrect round trip", float32(y), s2, z)
		}
		// Test shortest
		prec := countDigits(s1)
		s0 = strconv.AppendFloat(buf0[:0], x, 'e', prec-1, 32)
		if !bytes.Equal(s0, s1) {
			t.Errorf("x=%v => %q != %q", float32(x), s0, s1)
		}
		if prec > 1 {
			s0 := strconv.AppendFloat(buf0[:0], x, 'e', prec-2, 32)
			if z, _ := strconv.ParseFloat(string(s0), 32); z == x {
				t.Errorf("x=%v => %q not shortest", float32(x), s1)
			}

		}
		prec = countDigits(s2)
		s0 = strconv.AppendFloat(buf0[:0], y, 'e', prec-1, 32)
		if !bytes.Equal(s0, s2) {
			t.Errorf("y=%v => %q != %q", float32(y), s0, s2)
		}
		if prec > 1 {
			s0 := strconv.AppendFloat(buf0[:0], y, 'e', prec-2, 32)
			if z, _ := strconv.ParseFloat(string(s0), 32); z == y {
				t.Errorf("x=%v => %q not shortest", float32(y), s2)
			}

		}
		count += 2
	}

	basePrec := 24
	for digits := 10; digits > 0; digits-- {
		count = 0
		for exp := 0; exp <= 127-23; exp++ {
			roundUp = false
			AlmostDecimalMidpoint(exp, digits, 24, uint(basePrec+2*digits), +1, false, do)
			roundUp = true
			AlmostDecimalMidpoint(exp, digits, 24, uint(basePrec+2*digits), -1, false, do)
		}
		for exp := 1; exp <= 127+23; exp++ {
			if exp == 127+23 {
				// denormals
				roundUp = false
				AlmostDecimalMidpoint(-(exp - 1), digits, 23, uint(basePrec+2*digits), +1, true, do)
				roundUp = true
				AlmostDecimalMidpoint(-(exp - 1), digits, 23, uint(basePrec+2*digits), -1, true, do)
			} else {
				roundUp = false
				AlmostDecimalMidpoint(-exp, digits, 24, uint(basePrec+2*digits), +1, false, do)
				roundUp = true
				AlmostDecimalMidpoint(-exp, digits, 24, uint(basePrec+2*digits), -1, false, do)
			}
		}
		t.Logf("%d digits: %d numbers tested", digits, count)
	}
}

func TestTortureAtof64(t *testing.T) {
	count := 0
	buf := make([]byte, 64)
	roundUp := false
	do := func(x float64, n uint64, k int) {
		y := math.Nextafter(x, 2*x)
		// (x+y)/2 is very close to n * 10**k

		s := strconv.AppendUint(buf[:0], n, 10)
		s = append(s, 'e')
		s = strconv.AppendInt(s, int64(k), 10)

		z, err := strconv.ParseFloat(string(s), 64)
		if err != nil {
			t.Errorf("could not parse %q: %s", s, err)
			return
		}
		expect := x
		if roundUp {
			expect = y
		}
		if z != expect {
			t.Errorf("expected to parse %q as %b, got %b", s, expect, z)
		}
		//fmt.Printf("parse %q => %b (lo=%.42e, up=%.42e)\n", s, expect, x, y)
		count++
	}

	for digits := 18; digits > 0; digits-- {
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

func TestTortureAtof32(t *testing.T) {
	count := 0
	roundUp := false
	buf := make([]byte, 32)
	do := func(xx float64, n uint64, k int) {
		x := float32(xx)
		y := math.Nextafter32(x, 2*x)

		s := strconv.AppendUint(buf[:0], n, 10)
		s = append(s, 'e')
		s = strconv.AppendInt(s, int64(k), 10)

		zz, err := strconv.ParseFloat(string(s), 32)
		if err != nil {
			t.Errorf("could not parse %q: %s", s, err)
			return
		}
		z := float32(zz)
		expect := x
		if roundUp {
			expect = y
		}
		//fmt.Printf("x=%b y=%b midpoint=%.30e\n", x, y, (float64(x)+float64(y))/2)
		if z != expect {
			t.Errorf("expected to parse %q as %b, got %b", s, expect, z)
		}
		//fmt.Printf("parse %q => %b got %b (lo=%.18e, up=%.18e)\n", s, expect, z, x, y)
		count++
	}

	basePrec := 24
	for digits := 10; digits > 0; digits-- {
		count = 0
		for exp := 10; exp <= 127-23; exp++ {
			roundUp = false
			AlmostDecimalMidpoint(exp, digits, 24, uint(basePrec+2*digits), +1, false, do)
			roundUp = true
			AlmostDecimalMidpoint(exp, digits, 24, uint(basePrec+2*digits), -1, false, do)
		}
		for exp := 10; exp <= 127+23; exp++ {
			if exp == 127+23 {
				// denormals
				roundUp = false
				AlmostDecimalMidpoint(-(exp - 1), digits, 23, uint(basePrec+2*digits), +1, true, do)
				roundUp = true
				AlmostDecimalMidpoint(-(exp - 1), digits, 23, uint(basePrec+2*digits), -1, true, do)
			} else {
				roundUp = false
				AlmostDecimalMidpoint(-exp, digits, 24, uint(basePrec+2*digits), +1, false, do)
				roundUp = true
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
				t.Errorf("x=%.32e digits=%d => %q want %q ERR", x, digits, s1, s2)
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

		for exp := 0; exp < 1024-52; exp++ {
			roundUp = true
			AlmostHalfDecimal(exp, digits, 53, uint(difficulty), +1, false, do)
			roundUp = false
			AlmostHalfDecimal(exp, digits, 53, uint(difficulty), -1, false, do)
		}
		for exp := 1; exp < 1024+52; exp++ {
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
				t.Errorf("x=%.32e digits=%d => %q want %q ERR", x, digits, s1, s2)
				errors++
			} else {
				//t.Logf("x=%.32e digits=%d => %q %q OK\n", x, digits, s1, s2)
			}
		}

		for exp := 0; exp <= 127-23; exp++ {
			roundUp = true
			AlmostHalfDecimal(exp, digits, 24, uint(prec+2*digits), +1, false, do)
			roundUp = false
			AlmostHalfDecimal(exp, digits, 24, uint(prec+2*digits), -1, false, do)
		}
		for exp := 1; exp <= 127+23; exp++ {
			if exp == 127+23 {
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

// appendE prints mant*10^exp in scientific notation.
func appendE(s []byte, mant uint64, digits int, exp int) ([]byte, bool) {
	if digits == -1 {
		// Shortest representation
		for mant%10 == 0 {
			mant /= 10
			exp += 1
		}
		digits = 0
		m := mant
		for m > 0 {
			digits++
			m /= 10
		}
	}
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

func countDigits(s []byte) (digits int) {
	for _, c := range s {
		if '0' <= c && c <= '9' {
			digits++
		}
		if c == 'e' {
			return
		}
	}
	return
}
