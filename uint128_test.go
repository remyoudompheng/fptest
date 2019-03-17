package fptest

import (
	"math/big"
	"testing"
)

func BenchmarkBigDiv128(b *testing.B) {
	n, errn := new(big.Int).SetString("1000000000000000000000000", 10)
	d, errd := new(big.Int).SetString("717897987691852588770249", 10)
	if !errn || !errd {
		b.Fatal(errn, errd)
	}
	b.ResetTimer()

	quoB, remB := new(big.Int), new(big.Int)
	for i := 0; i < b.N; i++ {
		quoB.DivMod(n, d, remB)
		if b.N == 1 {
			b.Logf("%d / %d = %d, rem %d", n, d, quoB, remB)
		}
	}
}

func BenchmarkBigDiv128Simple(b *testing.B) {
	n, errn := new(big.Int).SetString("123456789123456789123456789023456789", 10)
	d, errd := new(big.Int).SetString("123456789", 10)
	if !errn || !errd {
		b.Fatal(errn, errd)
	}
	b.ResetTimer()

	quo, rem := new(big.Int), new(big.Int)
	for i := 0; i < b.N; i++ {
		quo.DivMod(n, d, rem)
	}
	if b.N == 1 {
		b.Logf("%d / %d = %d, rem %d", n, d, quo, rem)
	}
}

func BenchmarkBigDiv128Large(b *testing.B) {
	n, errn := new(big.Int).SetString("123456789123456789123456789123456789", 10)
	d, errd := new(big.Int).SetString("123456789123456789123456789", 10)
	if !errn || !errd {
		b.Fatal(errn, errd)
	}
	b.ResetTimer()

	quo, rem := new(big.Int), new(big.Int)
	for i := 0; i < b.N; i++ {
		quo.DivMod(n, d, rem)
	}
	if b.N == 1 {
		b.Logf("%d / %d = %d, rem %d", n, d, quo, rem)
	}
}

func BenchmarkDiv128(b *testing.B) {
	n := [2]uint64{
		1000000000000000000000000 >> 64,
		1000000000000000000000000 & (1<<64 - 1),
	}
	d := [2]uint64{
		717897987691852588770249 >> 64,
		717897987691852588770249 & (1<<64 - 1),
	}
	quo, rem := Divmod128(n, d)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		quo, rem = Divmod128(n, d)
	}
	if b.N == 1 {
		b.Logf("%x / %x = %x, rem %d", n, d, quo, rem)
	}
}

func BenchmarkDiv128Simple(b *testing.B) {
	n := [2]uint64{
		123456789123456789123456789023456789 >> 64,
		123456789123456789123456789023456789 & (1<<64 - 1),
	}
	d := [2]uint64{0, 123456789}
	quo, rem := Divmod128(n, d)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		quo, rem = Divmod128(n, d)
	}
	if b.N == 1 {
		b.Logf("%x / %x = %x, rem %d", n, d, quo, rem)
	}
}

func BenchmarkDiv128Large(b *testing.B) {
	n := [2]uint64{
		123456789123456789123456789123456789 >> 64,
		123456789123456789123456789123456789 & (1<<64 - 1),
	}
	d := [2]uint64{
		123456789123456789123456789 >> 64,
		123456789123456789123456789 & (1<<64 - 1),
	}
	quo, rem := Divmod128(n, d)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		quo, rem = Divmod128(n, d)
	}
	if b.N == 1 {
		b.Logf("%x / %x = %d, rem %d", n, d, quo, rem)
	}
}

func TestDiv128(t *testing.T) {
	n := [2]uint64{
		123456789123456789123456789123456789 >> 64,
		123456789123456789123456789123456789 & (1<<64 - 1),
	}
	d := [2]uint64{
		123456789123456789123456789 >> 64,
		123456789123456789123456789 & (1<<64 - 1),
	}
	quo, rem := Divmod128(n, d)
	t.Log("123456789123456789123456789123456789 /",
		"123456789123456789123456789 =",
		quo, rem)
	if quo[1] != 1000000000 || rem[1] != 123456789 {
		t.Error("error")
	}
}
