package fptest

import (
	"math"
	"math/big"
	"math/bits"
)

// AlmostDecimalPos enumerates floating-point numbers
// mant × 2**e2 such that the midpoint (mant+1/2)×2**e2
// is very close to n × 10**k where n is an integer.
//
// direction = +1 will returns numbers slightly above n × 10**k
// direction = -1 will returns numbers slightly below n × 10**k
//
// Very close is interpreted as a relative difference less than
// 1 / 2^precision.
func AlmostDecimalPos(e2 int, digits int, mantbits, precision uint, direction int, f func(float64)) {
	// Find all rationals n / (2*mant+1) close to 2**(e2-1) / 10**k
	//
	// (k + digits) * log(10) == (mantbits + e2) * log(2)
	e10 := int(math.Ceil(float64(e2+int(mantbits))*log2overlog10)) - digits

	num := big.NewInt(1)
	num.Lsh(num, uint(e2-1))
	den := big.NewInt(10)
	den.Exp(den, big.NewInt(int64(e10)), nil)
	r1 := NewRatFromBig(num, den, mantbits+1)
	var r2 *Rat
	if direction == -1 {
		// n/(2*mant+1) is slight too large.
		// num2 = num * (1 << precision + 1)
		// den2 = den << precision
		num2 := new(big.Int).Lsh(num, precision)
		den2 := new(big.Int).Lsh(den, precision)
		num2 = num2.Add(num2, num)
		r2 = NewRatFromBig(num2, den2, mantbits+1)
	} else {
		num2 := new(big.Int).Lsh(num, precision)
		den2 := new(big.Int).Lsh(den, precision)
		num2 = num2.Sub(num2, num)
		r2 = r1
		r1 = NewRatFromBig(num2, den2, mantbits+1)
	}
	for r := r1; r.Less(r2); r.Next() {
		_, b := r.Fraction()
		//fmt.Println(r.cf, r.a, r.c)
		if b%2 == 1 && bits.Len64(b) == int(mantbits+1) {
			f(math.Ldexp(float64(b/2), e2))
		}
	}
}

const log2overlog10 = 0.30102999566398114

// AlmostDecimalNeg enumerates numbers mant/2**e2 such that
// the midpoint (mant+1/2)/2**e2 is very close to n/10**k for some integer n.
func AlmostDecimalNeg(e2 int, digits int, mantbits, precision uint,
	direction int, denormals bool, f func(float64)) {
	// Find all rationals n / (2*mant+1) close to 10**k/2**(e2+1)
	//
	// (digits - k) * log(10) == (mantbits - e2) * log(2)
	e10 := int(float64(e2-int(mantbits))*log2overlog10) + digits

	num := big.NewInt(10)
	num.Exp(num, big.NewInt(int64(e10)), nil)
	den := big.NewInt(1)
	den.Lsh(den, uint(e2+1))
	r1 := NewRatFromBig(num, den, mantbits+1)
	var r2 *Rat
	if direction == -1 {
		// n/(2*mant+1) is slight too large.
		// num2 = num * (1 << precision + 1)
		// den2 = den << precision
		num2 := new(big.Int).Lsh(num, precision)
		den2 := new(big.Int).Lsh(den, precision)
		num2 = num2.Add(num2, num)
		r2 = NewRatFromBig(num2, den2, mantbits+1)
	} else {
		num2 := new(big.Int).Lsh(num, precision)
		den2 := new(big.Int).Lsh(den, precision)
		num2 = num2.Sub(num2, num)
		r2 = r1
		r1 = NewRatFromBig(num2, den2, mantbits+1)
	}
	for r := r1; r.Less(r2); r.Next() {
		_, b := r.Fraction()
		if b%2 == 1 && (denormals || bits.Len64(b) == int(mantbits+1)) {
			f(math.Ldexp(float64(b/2), -e2))
		}
	}
}

// AlmostHalfDecimalPos enumerates floating-point numbers mant*2**e2
// are very close to half a decimal number (n+1/2)*10**k.
func AlmostHalfDecimalPos(e2 int, digits int, mantbits, precision uint, direction int, f func(float64)) {
	// Find all rationals (2n+1) / mant close to 2**(e2+1) / 10**k
	e10 := int(math.Ceil(float64(e2+int(mantbits))*log2overlog10)) - digits

	num := big.NewInt(1)
	num.Lsh(num, uint(e2+1))
	den := big.NewInt(10)
	den.Exp(den, big.NewInt(int64(e10)), nil)
	r1 := NewRatFromBig(num, den, mantbits)
	var r2 *Rat
	if direction == -1 {
		// (2n+1)/mant is slightly too large.
		num2 := new(big.Int).Lsh(num, precision)
		den2 := new(big.Int).Lsh(den, precision)
		num2 = num2.Add(num2, num)
		r2 = NewRatFromBig(num2, den2, mantbits)
	} else {
		num2 := new(big.Int).Lsh(num, precision)
		den2 := new(big.Int).Lsh(den, precision)
		num2 = num2.Sub(num2, num)
		r2 = r1
		r1 = NewRatFromBig(num2, den2, mantbits)
	}
	for r := r1; r.Less(r2); r.Next() {
		a, b := r.Fraction()
		if a%2 == 1 && bits.Len64(b) == int(mantbits) {
			f(math.Ldexp(float64(b), e2))
		}
	}
}

// AlmostHalfDecimalNeg enumerates floating-point numbers mant/2**e2
// are very close to half a decimal number (n+1/2)/10**k.
func AlmostHalfDecimalNeg(e2 int, digits int, mantbits, precision uint, direction int, denormal bool, f func(float64)) {
	// Find all rationals (2n+1) / mant close to 10**k / 2**(e2-1)
	e10 := int(float64(e2-int(mantbits))*log2overlog10) + digits

	num := big.NewInt(10)
	num.Exp(num, big.NewInt(int64(e10)), nil)
	den := big.NewInt(1)
	den.Lsh(den, uint(e2-1))
	r1 := NewRatFromBig(num, den, mantbits)
	var r2 *Rat
	if direction == -1 {
		// (2n+1)/mant is slightly too large.
		r2 = slightlyOff(num, den, precision, +1, mantbits)
	} else {
		r2 = r1
		if len(r2.cf)%2 == 1 {
			r2.Next() // r2 is included
		}
		r1 = slightlyOff(num, den, precision, -1, mantbits)
	}
	for r := r1; r.Less(r2); r.Next() {
		a, b := r.Fraction()
		if a%2 == 1 && (denormal || bits.Len64(b) == int(mantbits)) {
			f(math.Ldexp(float64(b), -e2))
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
	} else {
		num2 = num2.Sub(num2, num)
	}
	return NewRatFromBig(num2, den2, maxBits)
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

// NewRat returns a new Rat equal to num/den, except
// if Bitlen(den) > maxBits, in which case a convergent
// of the continued fraction expansion will be returned
// instead.
func NewRat(num, den uint64, maxBits uint) *Rat {
	r := &Rat{
		maxBits: maxBits,
		a:       1,
		d:       1,
	}
euclid:
	for den > 0 {
		quo, rem := num/den, num%den
		newc := quo*r.c + r.d
		switch {
		case bits.Len64(newc) > int(maxBits),
			r.c > 0 && bits.Len64(newc) == int(maxBits) && newc/r.c != quo:
			// stop here
			break euclid
		}
		r.cf = append(r.cf, quo)
		r.a, r.b = quo*r.a+r.b, r.a
		r.c, r.d = quo*r.c+r.d, r.c
		num, den = den, rem
	}
	if k := r.cf[len(r.cf)-1]; k == 1 {
		// normalize
		r.cf = r.cf[:len(r.cf)-1]
		r.cf[len(r.cf)-1]++
		r.b = r.a - r.b
		r.d = r.c - r.d
	}
	return r
}

func NewRatFromBig(num, den *big.Int, maxBits uint) *Rat {
	r := &Rat{
		maxBits: maxBits,
		a:       1,
		d:       1,
	}
euclid:
	for den.BitLen() > 0 {
		quoB, remB := new(big.Int), new(big.Int)
		quoB.DivMod(num, den, remB)
		if quoB.BitLen() >= 64 {
			// stop here
			break
		}
		quo := quoB.Uint64()
		newc := quo*r.c + r.d
		switch {
		case bits.Len64(newc) > int(maxBits),
			r.c > 1 && newc/r.c != quo:
			// stop here
			break euclid
		}
		r.cf = append(r.cf, quo)
		r.a, r.b = quo*r.a+r.b, r.a
		r.c, r.d = quo*r.c+r.d, r.c
		num, den = den, remB
	}
	if k := r.cf[len(r.cf)-1]; k == 1 {
		// normalize
		r.cf = r.cf[:len(r.cf)-1]
		r.cf[len(r.cf)-1]++
		r.b = r.a - r.b
		r.d = r.c - r.d
	}
	return r
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
