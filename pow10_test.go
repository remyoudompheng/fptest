package fptest

import (
	"fmt"
	"math/big"
	"testing"
)

func TestCarry64(t *testing.T) {
	// Use cases includes:
	// * ftoa32: multiplier is a 25-bit mantissa and we need
	//   a N-bit significand.
	// * atof32: multiplier is a N-bit mantissa and we need
	//   a 25-bit significand.

	// 36 bits yields 6 edge cases.
	// 35 bits yields 3 edge cases
	// - 29842624 * 1e58
	// - 22550054 * 1e61
	// - 29753718 * 1e-61
	// 34 bits yields no error.
	const FTOA_BITS = 34

	// 35 bits has exactly 1 edge case:
	// - 13860322284 * 1e-48
	// 34, 33, 32 bits has the same edge case
	// 31 bits yields no error
	const ATOF_BITS = 31

	for i := 28; i < 70; i++ {
		// Don't test exact powers of 10.
		var m1, m2 [2]uint64
		// 64-bit truncation of 10^67 and upper bound
		m1[1] = pow10wide[i][0]
		m2[1] = pow10wide[i][0] + 1

		title := fmt.Sprint("ftoa, exponent ", i)
		testNoCarry(t, title, m1, m2, 25, 64+24-FTOA_BITS)
		title = fmt.Sprint("atof, exponent ", i)
		testNoCarry(t, title, m1, m2, ATOF_BITS, 64+ATOF_BITS-1-25)
	}

	for i := 11; i < 70; i++ {
		// Don't test exact powers of 10.
		var m1, m2 [2]uint64
		// 64-bit truncation of 10^67 and upper bound
		m1[1] = invpow10wide[i][0]
		m2[1] = invpow10wide[i][0] + 1

		title := fmt.Sprint("ftoa, exponent ", -i)
		testNoCarry(t, title, m1, m2, 25, 64+24-FTOA_BITS)
		title = fmt.Sprint("atof, exponent ", -i)
		testNoCarry(t, title, m1, m2, ATOF_BITS, 64+ATOF_BITS-1-25)
	}

	t.Logf("tested multiplier of size %d, shift %d",
		25, 64+24-FTOA_BITS)
	t.Logf("tested multiplier of size %d, shift %d",
		ATOF_BITS, 64+ATOF_BITS-1-25)
}

func TestCarry128(t *testing.T) {
	const mantbitsFtoa = 55
	const mantbitsAtof = 54
	// Use cases includes:
	// * ftoa64: multiplier is a 55-bit mantissa and we need
	//   a N-bit significand.
	// * atof64: multiplier is a N-bit mantissa and we need
	//   a 54-bit significand.

	// For FTOA_BITS = 64, there is a single edge case:
	// 34742740578729299 * 1e167 for a 55 bit mantissa.
	const FTOA_BITS = 63
	const ATOF_BITS = 64

	for i := 56; i < len(pow10wide); i++ {
		// Don't test exact powers of 10.
		m1 := pow10wide[i]
		m2 := pow10wide[i]
		m2[1]++

		title := fmt.Sprint("ftoa, exponent ", i)
		testNoCarry(t, title, m1, m2, mantbitsFtoa,
			127+mantbitsFtoa-FTOA_BITS)
		title = fmt.Sprint("atof, exponent ", i)
		testNoCarry(t, title, m1, m2, ATOF_BITS,
			127+ATOF_BITS-mantbitsAtof)
	}

	for i := 28; i < len(invpow10wide); i++ {
		// Don't test exact powers of 10.
		m1 := invpow10wide[i]
		m1[1]--
		m2 := invpow10wide[i]

		title := fmt.Sprint("ftoa, exponent ", -i)
		testNoCarry(t, title, m1, m2, mantbitsFtoa,
			127+mantbitsFtoa-FTOA_BITS)
		title = fmt.Sprint("atof, exponent ", -i)
		testNoCarry(t, title, m1, m2, ATOF_BITS,
			127+ATOF_BITS-mantbitsAtof)
	}

	t.Logf("tested multiplier of size %d, shift %d",
		mantbitsFtoa, 127+mantbitsFtoa-FTOA_BITS)
	t.Logf("tested multiplier of size %d, shift %d",
		ATOF_BITS, 127+ATOF_BITS-mantbitsAtof)
}

// testNoCarry takes 128-bit values m1 and m2, and checks
// whether k * m1 >> shift == k * m2 >> shift
// for all k <= 1<<inbits
//     and shift = 128 + inbits - outbits
func testNoCarry(t *testing.T, title string, m1, m2 [2]uint64, inbits, shift int) {
	// The invariant will be broken if we find:
	//    k * m1 <= K << shift <= k * m2
	// i.e.
	//    m1 / 2^shift <= K / k <= m2 / 2^shift
	var r1, r2 *Rat
	if shift < 128 {
		pow2 := [2]uint64{0, 1 << uint(shift)}
		if shift >= 64 {
			pow2 = [2]uint64{1 << uint(shift-64), 0}
		}
		_, r1 = NewRat128(m1, pow2, uint(inbits))
		r2, _ = NewRat128(m2, pow2, uint(inbits))
		r2.Next()
	} else {
		n1, ok1 := new(big.Int).SetString(
			fmt.Sprintf("%016x%016x", m1[0], m1[1]), 16)
		n2, ok2 := new(big.Int).SetString(
			fmt.Sprintf("%016x%016x", m2[0], m2[1]), 16)
		if !ok1 || !ok2 {
			t.Fatal("impossible")
		}
		pow2 := big.NewInt(1)
		pow2.Lsh(pow2, uint(shift))
		_, r1 = NewRatFromBig(n1, pow2, uint(inbits))
		r2, _ = NewRatFromBig(n2, pow2, uint(inbits))
		r2.Next()
	}
	//t.Logf("(0x%016x%016x, 0x%016x%016x) >> %d",
	//	m1[0], m1[1], m2[0], m2[1], shift)
	//t.Log(r1.Fraction())
	//t.Log(r2.Fraction())

	for r := r1; r.Less(r2); r.Next() {
		num, den := r.Fraction()
		switch num {
		case 65536, 131072, 262144, 524288, 1048576,
			1 << 21, 1 << 22, 1 << 23, 1 << 24, 1 << 25:
			switch den {
			case 48828125, 244140625,
				1220703125, 6103515625, 30517578125:
				// don't error on exact 2*a / 5*b
				continue
			}
		}
		t.Errorf("%s: %d * (0x%016x%016x, 0x%016x%016x) >> %d contains %d...\n",
			title, den, m1[0], m1[1], m2[0], m2[1], shift, num)
	}
}
