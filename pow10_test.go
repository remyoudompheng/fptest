package fptest

import (
	"testing"
)

func TestCarry64(t *testing.T) {
	// Use cases includes:
	// * ftoa32: multiplier is a 25-bit mantissa and we need
	//   a 32-bit significand.
	// * atof32: multiplier is a n-bit mantissa and we need
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

		t.Log("ftoa, exponent", i)
		testNoCarry(t, m1, m2, 25, 64+24-FTOA_BITS)
		t.Log("atof, exponent", i)
		testNoCarry(t, m1, m2, ATOF_BITS, 64+ATOF_BITS-1-25)
	}

	for i := 11; i < 70; i++ {
		// Don't test exact powers of 10.
		var m1, m2 [2]uint64
		// 64-bit truncation of 10^67 and upper bound
		m1[1] = invpow10wide[i][0]
		m2[1] = invpow10wide[i][0] + 1

		t.Log("ftoa, exponent", -i)
		testNoCarry(t, m1, m2, 25, 64+24-FTOA_BITS)
		t.Log("atof, exponent", -i)
		testNoCarry(t, m1, m2, ATOF_BITS, 64+ATOF_BITS-1-25)
	}
}

// testNoCarry takes 128-bit values m1 and m2, and checks
// whether k * m1 >> shift == k * m2 >> shift
// for all k <= 1<<inbits
//     and shift = 128 + inbits - outbits
func testNoCarry(t *testing.T, m1, m2 [2]uint64, inbits, shift int) {
	// The invariant will be broken if we find:
	//    k * m1 <= K << shift <= k * m2
	// i.e.
	//    m1 / 2^shift <= K / k <= m2 / 2^shift
	pow2 := [2]uint64{0, 1 << uint(shift)}
	if shift >= 64 {
		pow2 = [2]uint64{1 << uint(shift-64), 0}
	}
	_, r1 := NewRat128(m1, pow2, uint(inbits))
	r2, _ := NewRat128(m2, pow2, uint(inbits))
	r2.Next()
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
		t.Errorf("%d * (0x%016x%016x, 0x%016x%016x) >> %d contains %d...\n",
			den, m1[0], m1[1], m2[0], m2[1], shift, num)
	}
}
