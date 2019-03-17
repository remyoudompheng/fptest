package fptest

import (
	"fmt"
	"math/big"
	"math/bits"
	"testing"
)

func ExampleRat_Next() {
	// Generates the Farey sequence F_7.
	var nums, dens []uint64
	r := NewRat(1, 7, 3)
	for r.a*r.c != 1 {
		nums = append(nums, r.a)
		dens = append(dens, r.c)
		r.Next()
	}
	fmt.Println(nums)
	fmt.Println(dens)
	// Output:
	// [1 1 1 1 2 1 2 3 1 4 3 2 5 3 4 5 6]
	// [7 6 5 4 7 3 5 7 2 7 5 3 7 4 5 6 7]
}

func TestNewRat(t *testing.T) {
	r := NewRat(355, 113, 8)
	num, den := r.Fraction()
	if num != 355 || den != 113 {
		t.Errorf("got %d/%d, expect 355/113", num, den)
	}
	t.Logf("%d/%d = %v", num, den, r.cf)

	r = NewRat(355, 113, 4)
	num, den = r.Fraction()
	if num != 22 || den != 7 {
		t.Errorf("got %d/%d, expect 22/7", num, den)
	}
	t.Logf("%d/%d = %v", num, den, r.cf)

	r = NewRat(89, 55, 8)
	num, den = r.Fraction()
	if num != 89 || den != 55 {
		t.Errorf("got %d/%d, expect 89/55", num, den)
	}
	t.Logf("%d/%d = %v", num, den, r.cf)
}

func TestRatFromBig(t *testing.T) {
	// 3**50 / 10**24
	n, _ := new(big.Int).SetString("717897987691852588770249", 10)
	d, _ := new(big.Int).SetString("1000000000000000000000000", 10)
	r := NewRatFromBig(n, d, 64)
	t.Log(r.cf)
	num, den := r.Fraction()
	if num != 2159037562977366367 || den != 3007443397242258693 {
		t.Errorf("got %d/%d, expect 3^50/10^24 ~= 2159037562977366367/3007443397242258693",
			num, den)
	}
}

func TestRatNext(t *testing.T) {
	// Approximations of (10**24 ± 1) / 2**80 at 1.5e-29 precision
	r0 := NewRat(65352703432539, 79006570561214, 48)
	r1 := NewRat(34807131698651, 42079240217226, 48)

	r := r0
	count := 1
	for {
		n, d := r.Fraction()
		r.Next()
		// Check that r.a/r.c is the correct fraction
		num, den := r.slowFrac()
		if num != r.a || den != r.c {
			t.Errorf("expected %d/%d, got %d/%d", num, den, r.a, r.c)
		}
		// Check ordering
		x1, x0 := bits.Mul64(n, den)
		y1, y0 := bits.Mul64(d, num)
		if x1 > y1 || (x1 == y1 && x0 >= y0) {
			t.Errorf("r.Next <= r")
		}
		count++
		if r.Equals(r1) {
			break
		}
	}
	if count != 39930 {
		t.Errorf("expected 39930 elements, got %d", count)
	}
}

func BenchmarkNewRatFromBig(b *testing.B) {
	n, errn := new(big.Int).SetString("717897987691852588770249", 10)
	d, errd := new(big.Int).SetString("1000000000000000000000000", 10)
	if !errn || !errd {
		b.Fatal(errn, errd)
	}
	b.ResetTimer()

	var r *Rat
	for i := 0; i < b.N; i++ {
		r = NewRatFromBig(n, d, 64)
	}
	if b.N == 1 {
		b.Log(r.Fraction())
	}
}

func BenchmarkNewRat128(b *testing.B) {
	n := [2]uint64{717897987691852588770249 >> 64, 717897987691852588770249 & (1<<64 - 1)}
	d := [2]uint64{1000000000000000000000000 >> 64, 1000000000000000000000000 & (1<<64 - 1)}
	b.ResetTimer()

	var r *Rat
	for i := 0; i < b.N; i++ {
		r = NewRat128(n, d, 64)
	}
	if b.N == 1 {
		b.Log(r.Fraction())
	}
}

func BenchmarkRat_Next(b *testing.B) {
	// Enumerate rationals between (10**24 ± 1) / 2**80
	// with 60-bit denominators.
	// There are about 2**40 such numbers.

	// r0 and r1 are approximations of (10**24 ± 1) / 2**80
	// with a respective precision of 2.76e-36 and 3.89e-37.
	r0 := NewRat(132262670593960591, 159895757452223520, 60)
	r1 := NewRat(902438988994577458, 1090981794422466871, 60)

	r := r0
	for i := 0; i < b.N; i++ {
		r.Next()
		if r.Equals(r1) {
			b.Fatal("too fast!")
		}
	}
}

func BenchmarkRat_Interval(b *testing.B) {
	for i := 0; i < b.N; i++ {
		r0 := NewRat(65352703432539, 79006570561214, 48)
		r1 := NewRat(34807131698651, 42079240217226, 48)

		count := 0
		for r := r0; !r.Equals(r1); r.Next() {
			count++
		}
		if count != 39929 {
			b.Fatal("incorrect", count)
		}
	}
}
