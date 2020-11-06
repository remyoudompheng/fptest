# fptest

An iterator over floating-point numbers which are hard
to convert from/to decimal form.

## Math

Converting a number `x` from/to another base (usually 2 or 10), as done in
`atof` or `ftoa` functions, involves multiplying

    x × bⁿ = digits + ε

where `|ε| < 1/2` and `digits` is the rounded result, used for display.

A conversion is **hard** is ε is very close to ±1/2, meaning that any
lack of precision might make rounding go the wrong side.

Usually `x` itself is expressed in some base B, meaning that the
computation is:

    mantissa × Bᵖ × bⁿ = digits + ε

where `mantissa` and `digits` have fixed predefined precision,
and the exponents `n` and `p` can be computed from each other.

Thus the problem can be expressed as enumerating all fractions
`mantissa / digits`, within the range of fixed precision integers,
which are very close to `Bᵖ × bⁿ` up to some precision.

An easy way to perform this enueration is to walk the Stern-Brocot
tree, which arranges rational numbers in an (unbounded depth) binary
search tree, such that children of a node always have larger denominator
than their parent node.

Manipulating the Stern-Brocot tree is easiest where elements are written
in continued fraction form.

## Implementation

The algorithm is implemented in Python 3 and Go.

For each pair of interesting exponents, a interval of fractions
around `2ᵖ × 10ⁿ` is computed, such that lower and upper bounds
have numerators and denominators, bounded by a max integer `M`.

This can easily by precomputed as a small table for languages
where arbitrary-precision arithmetic ("big integers") is hard
or inconvenient.

Then the algorithm traverses entirely the tree between these 2 bounds,
stopping at the depth where integers would overflow the bound `M`.
This step can be done using finite precision arithmetic only.

## Performance

The Python script takes about 1 minute to enumerate double-precision
floats (float64) such that formatting in decimal form is hard
even using 96-bit arithmetic (about 1 million values).

It takes less than a second to enumerate the single-precision
(float32) values that are hard to format when using 52-bit arithmetic
(about 280 values).

The Go implementation can enumerate an interval of 1 million rational numbers
in a few milliseconds.

## Available tests

The following torture tests are implemented in Go:

- TestCarry64, TestCarry128: check wide-precision powers of ten for
  exceptional carries. This is used in the proof of Ryū-like algorithms.

- TestTortureFixed32/64: check edge cases for fixed-precission decimal
  formatting. The iterators provide the expected answer so it is checked
  exactly without depending on strconv correctness.

- TestTortureShortest32/64: check edge cases for shortest floating-point
  formatting. Currently exact verification is not done, instead the test
  verifies that selected numbers are formatted shorter than their
  neighbours. The edge cases make it hard to find the correct rounding
  direction.

- TestTortureAtof32/64: check edge cases for parsing floats.
  The inputs are decimal representations whch are very close
  to midpoints between consecutive float32/float64s. The iterator
  gives the expected answer without using strconv functions.

Exact midpoints (commonly found when using small exponents) are not tested.
Small exponents are not tested (|exp| < 55 for float64, |exp| < 10 for
float32)

## References

- [Wikipedia (Stern-Brocot Tree)](https://en.wikipedia.org/wiki/Stern%E2%80%93Brocot_tree)
