package fptest

import (
	"math/bits"
)

// 128-bit arithmetic without math/big.

// Implementation of divmod for uint128.

func Divmod128(a, b [2]uint64) (quo, rem [2]uint64) {
	if b[0] == 0 {
		q, r := a[0]/b[1], a[0]%b[1]
		quo[0] = q
		a[0] = r
		q, r = bits.Div64(a[0], a[1], b[1])
		quo[1] = q
		rem[1] = r
	} else if a[0] < b[0] {
		return quo, a
	} else {
		// extract 64 top bits of b.
		btop := b[0]
		l := uint(bits.Len64(btop + 1)) // > 0
		btop <<= (64 - l)
		btop |= b[1] >> l
		// b = (btop << l) + ε
		// make sure a[0] is less than btop
		if a[0] > btop {
			// substract b << (64-l) == btop << 64 | b[1] << (64-l)
			a[0] -= btop
			if a[1] < b[1]<<(64-l) {
				a[0]--
			}
			a[1] -= b[1] << (64 - l)
			quo[1] += 1 << (64 - l)
		}
		// divide by btop+1
		btop += 1 // round up.
		q, _ := bits.Div64(a[0], a[1], btop)
		// a = q * btop + r
		//   = (q >> l) * (b + ε) + (qlow * btop) + r
		//   = (q >> l) * b + (q>>l)*ε + qlow * btop + r
		qhi := q >> l
		quo[1] += qhi // cannot overflow
		// substract qhi*b
		a[0] -= qhi * b[0]
		z0, z1 := bits.Mul64(qhi, b[1])
		a[0] -= z0
		if a[1] < z1 {
			a[0]--
		}
		a[1] -= z1

		for a[0] > b[0] || (a[0] == b[0] && a[1] >= b[1]) {
			quo[1]++
			a[0] -= b[0]
			if a[1] < b[1] {
				a[0]--
			}
			a[1] -= b[1]
		}
		return quo, a
	}
	return
}
