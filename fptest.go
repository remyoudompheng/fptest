package fptest

import (
	"math"
	"math/big"
	"math/bits"
)

// AlmostDecimalMidpoint enumerates floating-point numbers
// mant × 2**e2 such that the midpoint (mant+1/2)×2**e2
// is very close to n × 10**k where n is an integer.
//
// direction = +1 will returns numbers slightly above n × 10**k
// direction = -1 will returns numbers slightly below n × 10**k
//
// Very close is interpreted as a relative difference less than
// 1 / 2^precision.
func AlmostDecimalMidpoint(e2 int, digits int, mantbits, precision uint, direction int, denormal bool, f func(float64)) {
	if e2 >= 0 {
		almostDecimalPos(e2, digits, mantbits, precision, direction, f)
	} else {
		almostDecimalNeg(-e2, digits, mantbits, precision, direction, denormal, f)
	}
}

const log2overlog10 = 0.30102999566398114

// almostDecimalPos is AlmostDecimalMidpoint for e2 >= 0.
func almostDecimalPos(e2 int, digits int, mantbits, precision uint, direction int, f func(float64)) {
	// Find all rationals n / (2*mant+1) close to 2**(e2-1) / 10**k
	//
	// (k + digits) * log(10) == (mantbits + e2) * log(2)
	e10 := int(math.Ceil(float64(e2+int(mantbits))*log2overlog10)) - digits

	num := big.NewInt(1)
	num.Lsh(num, uint(e2-1))
	den := big.NewInt(10)
	den.Exp(den, big.NewInt(int64(e10)), nil)
	var r1, r2 *Rat
	if direction == -1 {
		// n/(2*mant+1) is slight too large.
		// num2 = num * (1 << precision + 1)
		// den2 = den << precision
		_, r1 = NewRatFromBig(num, den, mantbits+1)
		r2 = slightlyOff(num, den, precision, +1, mantbits+1)
	} else {
		r1 = slightlyOff(num, den, precision, -1, mantbits+1)
		r2, _ = NewRatFromBig(num, den, mantbits+1)
		r2.Next()
	}
	for r := r1; r.Less(r2); r.Next() {
		_, b := r.Fraction()
		//fmt.Println(r.cf, r.a, r.c)
		if b%2 == 1 && bits.Len64(b) == int(mantbits+1) {
			f(math.Ldexp(float64(b/2), e2))
		}
	}
}

// almostDecimalNeg enumerates numbers mant/2**e2 such that
// the midpoint (mant+1/2)/2**e2 is very close to n/10**k for some integer n.
func almostDecimalNeg(e2 int, digits int, mantbits, precision uint,
	direction int, denormals bool, f func(float64)) {
	// Find all rationals n / (2*mant+1) close to 10**k/2**(e2+1)
	//
	// (digits - k) * log(10) == (mantbits - e2) * log(2)
	e10 := int(float64(e2-int(mantbits))*log2overlog10) + digits

	num := big.NewInt(10)
	num.Exp(num, big.NewInt(int64(e10)), nil)
	den := big.NewInt(1)
	den.Lsh(den, uint(e2+1))
	var r1, r2 *Rat
	if direction == -1 {
		// n/(2*mant+1) is slight too large.
		// num2 = num * (1 << precision + 1)
		// den2 = den << precision
		_, r1 = NewRatFromBig(num, den, mantbits+1)
		r2 = slightlyOff(num, den, precision, +1, mantbits+1)
	} else {
		r1 = slightlyOff(num, den, precision, -1, mantbits+1)
		r2, _ = NewRatFromBig(num, den, mantbits+1)
		r2.Next()
	}
	for r := r1; r.Less(r2); r.Next() {
		_, b := r.Fraction()
		if b%2 == 1 && (denormals || bits.Len64(b) == int(mantbits+1)) {
			f(math.Ldexp(float64(b/2), -e2))
		}
	}
}

// AlmostHalfDecimal enumerates floating-point numbers mant*2**e2
// are very close to half a decimal number (n+1/2)*10**k.
func AlmostHalfDecimal(e2 int, digits int, mantbits, precision uint,
	direction int, denormal bool, f func(x float64, n uint64, k int)) {
	if e2 >= 0 {
		almostHalfDecimalPos(e2, digits, mantbits, precision, direction, f)
	} else {
		almostHalfDecimalNeg(-e2, digits, mantbits, precision, direction, denormal, f)
	}
}

func almostHalfDecimalPos(e2 int, digits int, mantbits, precision uint, direction int,
	f func(float64, uint64, int)) {
	// Find all rationals (2n+1) / mant close to 2**(e2+1) / 10**k
	e10 := int(math.Ceil(float64(e2+int(mantbits))*log2overlog10)) - digits

	num := big.NewInt(1)
	num.Lsh(num, uint(e2+1))
	den := big.NewInt(10)
	den.Exp(den, big.NewInt(int64(e10)), nil)
	var r1, r2 *Rat
	if direction == -1 {
		// (2n+1)/mant is slightly too large.
		_, r1 = NewRatFromBig(num, den, mantbits)
		r2 = slightlyOff(num, den, precision, +1, mantbits)
	} else {
		r1 = slightlyOff(num, den, precision, -1, mantbits)
		r2, _ = NewRatFromBig(num, den, mantbits)
		r2.Next()
	}
	for r := r1; r.Less(r2); r.Next() {
		a, b := r.Fraction()
		if a%2 == 1 && bits.Len64(b) == int(mantbits) {
			f(math.Ldexp(float64(b), e2), a/2, e10)
		}
	}
}

// almostHalfDecimalNeg implements AlmostHalfDecimal for negative exponents.
func almostHalfDecimalNeg(e2 int, digits int, mantbits, precision uint, direction int, denormal bool,
	f func(float64, uint64, int)) {
	// Find all rationals (2n+1) / mant close to 10**k / 2**(e2-1)
	e10 := int(float64(e2-int(mantbits))*log2overlog10) + digits

	num := big.NewInt(10)
	num.Exp(num, big.NewInt(int64(e10)), nil)
	den := big.NewInt(1)
	den.Lsh(den, uint(e2-1))
	var r1, r2 *Rat
	if direction == -1 {
		// (2n+1)/mant is slightly too large.
		_, r1 = NewRatFromBig(num, den, mantbits)
		r2 = slightlyOff(num, den, precision, +1, mantbits)
	} else {
		r1 = slightlyOff(num, den, precision, -1, mantbits)
		r2, _ = NewRatFromBig(num, den, mantbits)
		r2.Next()
	}
	for r := r1; r.Less(r2); r.Next() {
		a, b := r.Fraction()
		if a%2 == 1 && (denormal || bits.Len64(b) == int(mantbits)) {
			f(math.Ldexp(float64(b), -e2), a/2, -e10)
		}
	}
}

func slightlyOff(num, den *big.Int, precision uint, direction int, maxBits uint) *Rat {
	// num2 = num * (1 << precision + 1)
	// den2 = den << precision
	num2 := new(big.Int).Lsh(num, precision)
	den2 := new(big.Int).Lsh(den, precision)
	if direction == +1 {
		num2 = num2.Add(num2, num)
		_, r := NewRatFromBig(num2, den2, maxBits)
		return r
	} else {
		num2 = num2.Sub(num2, num)
		r, _ := NewRatFromBig(num2, den2, maxBits)
		return r
	}
}

// A Rat is a positive rational number, internally
// represented as a continued fraction.
// The numerator and denominator must fit in 64 bits.
type Rat struct {
	maxBits uint

	// cf is a continued fraction expansion.
	cf []uint64

	// (a b) is the product of matrices (cf[i] 1)
	// (c d)                            (  1   0)
	//
	// In particular, a/c is the irreductible
	// fraction representing the rational number.
	a, b uint64 // a > b
	c, d uint64 // c > d
}

// It will be required to find lower and upper rational
// approximations r- <= num/den <= r+
// Assuming that num/den has a continued fraction expansion:
//   [a0, a1, ... an ...]
// Then approximations are:
// on one side:
//   [a0, ..., ak]
//   [a0, ..., ak, a_(k+1)]

// NewRats returns two Rats, r1, r2 such that:
// r1 <= num/den <= r2, and the denominator of r1 and r2
// have at most maxBits bits.
func NewRat(num, den uint64, maxBits uint) (lower, upper *Rat) {
	return NewRat128([2]uint64{0, num}, [2]uint64{0, den}, maxBits)
}

func NewRatFromBig(num, den *big.Int, maxBits uint) (lower, upper *Rat) {
	r := &Rat{
		maxBits: maxBits,
		a:       1,
		d:       1,
	}
	var midCF []uint64
euclid:
	for den.BitLen() > 0 {
		quoB, remB := new(big.Int), new(big.Int)
		quoB.DivMod(num, den, remB)
		if quoB.BitLen() >= 64 {
			// stop here
			midCF = append(midCF, r.cf...)
			midCF = append(midCF, ^uint64(0))
			break
		}
		quo := quoB.Uint64()
		newc := quo*r.c + r.d
		switch {
		case bits.Len64(newc) > int(maxBits),
			r.c > 1 && newc/r.c != quo:
			// stop here
			midCF = append(midCF, r.cf...)
			midCF = append(midCF, quo)
			break euclid
		}
		r.appendContinued(quo)
		if len(r.cf)%2 == 0 {
			upper = r.clone()
		} else {
			lower = r.clone()
		}
		num, den = den, remB
	}
	if den.BitLen() == 0 {
		lower, upper = r, r
	} else {
		// Find closest approximations
		l := lower.clone()
		for l.leqCF(midCF) {
			lower = l.clone()
			l.Next()
		}
		upper = l.clone()
	}
	lower.normalize()
	upper.normalize()
	return
}

func NewRat128(num, den [2]uint64, maxBits uint) (lower, upper *Rat) {
	r := &Rat{
		maxBits: maxBits,
		a:       1,
		d:       1,
	}
	var midCF []uint64
euclid:
	for den != [2]uint64{} {
		quo, rem := Divmod128(num, den)
		if quo[0] > 0 {
			// stop here, unsupported
			midCF = append(midCF, r.cf...)
			midCF = append(midCF, ^uint64(0))
			break
		}
		q := quo[1]
		newc := q*r.c + r.d
		switch {
		case bits.Len64(newc) > int(maxBits),
			r.c > 1 && newc/r.c != q:
			// stop here
			midCF = append(midCF, r.cf...)
			midCF = append(midCF, q)
			break euclid
		}
		r.appendContinued(q)
		if len(r.cf)%2 == 0 {
			upper = r.clone()
		} else {
			lower = r.clone()
		}
		num, den = den, rem
	}
	if den == [2]uint64{} {
		lower, upper = r, r
	} else {
		// Find closest approximations
		l := lower.clone()
		for l.leqCF(midCF) {
			lower = l.clone()
			l.Next()
		}
		upper = l.clone()
	}
	lower.normalize()
	upper.normalize()
	return
}

func (r *Rat) appendContinued(q uint64) {
	r.cf = append(r.cf, q)
	r.a, r.b = q*r.a+r.b, r.a
	r.c, r.d = q*r.c+r.d, r.c
}

func (r *Rat) normalize() {
	if k := r.cf[len(r.cf)-1]; k == 1 {
		r.cf = r.cf[:len(r.cf)-1]
		r.cf[len(r.cf)-1]++
		r.b = r.a - r.b
		r.d = r.c - r.d
	}
}

// leq returns whether r is less than or equal to
// the fraction represented in continuous form (cf)
func (r *Rat) leqCF(cf []uint64) bool {
	for i := 0; i < len(r.cf) && i < len(cf); i++ {
		if r.cf[i] != cf[i] {
			if i%2 == 0 {
				return r.cf[i] <= cf[i]
			} else {
				return r.cf[i] >= cf[i]
			}
		}
	}
	// otherwise r.cf is a prefix of cf (or conversely)
	switch {
	case len(r.cf) == len(cf),
		// [3] <= [3, 7]
		len(r.cf) < len(cf) && len(r.cf)%2 == 1,
		// [1, 3, 7] <= [1, 3]
		len(r.cf) > len(cf) && len(cf)%2 == 0:
		return true
	}
	return false
}

func (r *Rat) clone() *Rat {
	rr := *r
	rr.cf = nil
	rr.cf = append(rr.cf, r.cf...)
	return &rr
}

func (r *Rat) Fraction() (num, den uint64) {
	return r.a, r.c
}

func (r *Rat) slowFrac() (num, den uint64) {
	num, den = 1, 0
	for i := len(r.cf) - 1; i >= 0; i-- {
		q := r.cf[i]
		num, den = q*num+den, num
	}
	return
}

func (r *Rat) Equals(s *Rat) bool {
	return r.a == s.a && r.c == s.c
}

func (r *Rat) Less(s *Rat) bool {
	x1, x0 := bits.Mul64(r.a, s.c)
	y1, y0 := bits.Mul64(s.a, r.c)
	return x1 < y1 || (x1 == y1 && x0 < y0)
}

// child mutates r to its left(idx=0) orright(idx=1)
// child in the Stern-Brocot tree.
func (r *Rat) child(idx int) {
	if len(r.cf)%2 == idx {
		r.cf[len(r.cf)-1]++
		// Multiply by (1 0)
		//             (1 1)
		r.a += r.b
		r.c += r.d
	} else {
		// (..., k) -> (..., k-1, 2)
		r.cf[len(r.cf)-1]--
		r.cf = append(r.cf, 2)
		// Multiply by ( 1 0) (2 1) = ( 2  1)
		//             (-1 1) (1 0)   (-1 -1)
		r.a, r.b = 2*r.a-r.b, r.a-r.b
		r.c, r.d = 2*r.c-r.d, r.c-r.d
	}
}

// peek returns the fraction for the specified Stern-Brocot child node,
// without mutating r.
func (r *Rat) peekChild(idx int) (num, den uint64) {
	if len(r.cf)%2 == idx {
		return r.a + r.b, r.c + r.d
	} else {
		return 2*r.a - r.b, 2*r.c - r.d
	}
}

// Next mutates r to the next rational number in the Farey sequence
// F_(1<<maxBits-1).
func (r *Rat) Next() {
	// The next element in the tree is either:
	// - the left-most leaf from the right child
	// - the first right-ancestor, i.e. N such that
	//   r is the right-most leaf of N.left_child
	_, den := r.peekChild(1)
	if bits.Len64(den) <= int(r.maxBits) && den >= r.c {
		// Right child is within bounds, go left-most.
		r.child(1)
		for {
			_, den = r.peekChild(0)
			if bits.Len64(den) > int(r.maxBits) || den < r.c {
				break
			}
			r.child(0)
		}
	} else {
		// Right child is out of bounds. Go up-left and right.
		if len(r.cf) <= 1 {
			println(r.a, r.c)
			panic("impossible")
		} else if len(r.cf)%2 == 0 {
			//                     (..k+1)
			//          ..(..k, 2)´
			// (..k, n)´
			//
			// Decrement the last coefficient.
			last := r.cf[len(r.cf)-1]
			if last > 2 {
				r.cf[len(r.cf)-1] = last - 1
				r.a -= r.b
				r.c -= r.d
			} else {
				// (.., k, 2) -> (.., k+1)
				// multiply by ( 1  1)
				//             (-1 -2)
				r.cf = r.cf[:len(r.cf)-1]
				r.cf[len(r.cf)-1]++
				r.a, r.b = r.a-r.b, r.a-2*r.b
				r.c, r.d = r.c-r.d, r.c-2*r.d
			}
		} else {
			//         _______________(..k)
			// (..k+1)´
			//    `(..k, 2)
			//            `... (..k, n)
			n := r.cf[len(r.cf)-1]
			r.cf = r.cf[:len(r.cf)-1]
			r.a, r.b = r.b, r.a-n*r.b
			r.c, r.d = r.d, r.c-n*r.d
			// Normalize k = 1
			// (..., l, k=1) => (..., l+1)
			k := r.cf[len(r.cf)-1]
			if k == 1 {
				r.cf = r.cf[:len(r.cf)-1]
				r.cf[len(r.cf)-1]++
				// Multiply by (0  1) (1 0) = (1  1)
				//             (1 -1) (1 1)   (0 -1)
				r.b = r.a - r.b
				r.d = r.c - r.d
			}
		}
	}
}
