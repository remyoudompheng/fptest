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

	// 36 bits yields 5 edge cases.
	// 33 bits yields 1 edge case:
	// 25287277 * 1e-30 => 25287277p+108 = 8206190558.000000024e30
	// 32 bits has 1 edge case:
	// 23215553 * 1e-11 => 23215553p+44 = 408412327500000002048
	const FTOA_BITS = 31

	// 32 bits has 1 error:
	// 4192293909e-14 = 46094759.00000000016p-40
	// 30 bits has 1 error:
	// 530184534e-13 = 29147203.0000000008192p-39
	const ATOF_BITS = 29

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
	r1 := NewRat128(m1, pow2, uint(inbits))
	r2 := NewRat128(m2, pow2, uint(inbits))
	//t.Logf("(0x%016x%016x, 0x%016x%016x) >> %d",
	//	m1[0], m1[1], m2[0], m2[1], shift)
	//t.Log(r1.Fraction())
	//t.Log(r2.Fraction())

	for r := r1; r.Less(r2); r.Next() {
		num, den := r.Fraction()
		t.Errorf("%d * (0x%016x%016x, 0x%016x%016x) >> %d contains %d...\n",
			den, m1[0], m1[1], m2[0], m2[1], shift, num)
	}
}
