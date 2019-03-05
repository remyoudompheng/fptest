package fptest

import (
	"math/big"
	"math/bits"
)

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
	for den > 0 {
		quo, rem := num/den, num%den
		newc := quo*r.c + r.d
		switch {
		case bits.Len64(newc) > int(maxBits),
			r.c > 0 && bits.Len64(newc) == int(maxBits) && newc/r.c != quo:
			// stop here
			return r
		}
		r.cf = append(r.cf, quo)
		r.a, r.b = quo*r.a+r.b, r.a
		r.c, r.d = quo*r.c+r.d, r.c
		num, den = den, rem
	}
	return r
}

func NewRatFromBig(num, den *big.Int, maxBits uint) *Rat {
	r := &Rat{
		maxBits: maxBits,
		a:       1,
		d:       1,
	}
	for den.BitLen() > 0 {
		quoB, remB := new(big.Int), new(big.Int)
		quoB.DivMod(num, den, remB)
		if quoB.BitLen() >= 64 {
			// stop here
			return r
		}
		quo := quoB.Uint64()
		newc := quo*r.c + r.d
		switch {
		case bits.Len64(newc) > int(maxBits),
			r.c > 1 && newc/r.c != quo:
			// stop here
			return r
		}
		r.cf = append(r.cf, quo)
		r.a, r.b = quo*r.a+r.b, r.a
		r.c, r.d = quo*r.c+r.d, r.c
		num, den = den, remB
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
