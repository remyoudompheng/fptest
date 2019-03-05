"""
Generate hard test cases for floating point conversion.

Requires Python 3.
"""

def main():
    """
    Sample expected output:
    for e2=827, e10=249, digs=14236688121214300 / mant=15907522898771511
    ε' = 2**827/10**249 - digs/mant = 5.765155354479547e-32
    """
    MODES = [
        "parse64+", "parse64-",
        "parse32+", "parse32-",
        "print64+", "print64-",
        "print32+", "print32-",
    ]
    import argparse
    p = argparse.ArgumentParser()
    p.add_argument("MODE", choices=MODES, nargs='?')
    args = p.parse_args()
    arg = args.MODE

    if not arg or arg == "parse64+":
        # 680k values
        for e2 in range(50, 1024-52):
            e10 = (e2 * 78913) >> 18
            find_hard_parse(e2, e10+1, mantbits=54, prec=96)

    if not arg or arg == "parse64-":
        # 600k values
        for e2 in range(20, 1024+52):
            e10 = (e2 * 78913) >> 18
            if e2 == 1075:
                # denormals have exponent p-1074 so midpoint have p-1075
                find_hard_parse_negexp(e2, e10, mantbits=53, prec=96, denormal=True)
            else:
                find_hard_parse_negexp(e2, e10, mantbits=54, prec=96)

    if not arg or arg == "print64+":
        # 275k values
        for e2 in range(30, 1024-52):
            e10 = (e2 * 78913) >> 18
            find_hard_print(e2, e10+1, mantbits=53, prec=96)

    if not arg or arg == "print64-":
        # 500k values
        for e2 in range(53, 1024+52):
            e10 = (e2 * 78913) >> 18
            if e2 == 1075:
                # denormals
                e2 = 1074
                find_hard_print_negexp(e2, e10, mantbits=52, prec=96, denormal=True)
            else:
                find_hard_print_negexp(e2, e10, mantbits=53, prec=96)

    # For float32, check values where 52 bit precision is not enough.
    if not arg or arg == "parse32+":
        # 138 values
        for e2 in range(24, 128-23):
            e10 = (e2 * 78913) >> 18
            find_hard_parse(e2, e10+1, mantbits=25, prec=52)

    if not arg or arg == "parse32-":
        # 145 values
        for e2 in range(16, 128+23):
            e10 = (e2 * 78913) >> 18
            if e2 == 150:
                # denormals have exponent p-149 (so midpoint is XXp-150)
                find_hard_parse_negexp(e2, e10+1, mantbits=24, prec=52, denormal=True)
            else:
                find_hard_parse_negexp(e2, e10+1, mantbits=25, prec=52)

    if not arg or arg == "print32+":
        for e2 in range(24, 128-23):
            e10 = (e2 * 78913) >> 18
            find_hard_print(e2, e10+1, mantbits=24, prec=48)

    if not arg or arg == "print32-":
        # 500k values
        for e2 in range(24, 128+23):
            e10 = (e2 * 78913) >> 18
            if e2 == 150:
                # denormals
                e2 = 149
                find_hard_print_negexp(e2, e10-1, mantbits=23, prec=48, denormal=True)
            else:
                find_hard_print_negexp(e2, e10, mantbits=24, prec=48)


def find_hard_parse(e2, e10, mantbits=54, prec=96):
    """
    Find floating point numbers which are hard to parse from decimal
    representation. The same numbers will be hard to format
    to their "shortest representation" because doing so requires
    knowing whether a representation parses back to the original number.

    e.g. find floating-point numbers with exponent 385 hard to parse.

    For example: 8640368759831959p+385

    The midpoint (8640368759831959 + 1/2) * 1<<385
    is 68089572682806429.999999999999999e115
    so it is hard to determine whethere 68089572682806430e115
    should parse to 8640368759831959p385 or 8640368759831960p385.

    We are looking for:
        mantissa × 2**385 × 10**-116 = digits + ε
    where digits < 1e16
          mantissa < 2**54
          mantissa is odd (mantissa of the midpoint)

    that is:
        2**385 / 10**116 = digits / mantissa + ε'

    The typical threshold we are interested in is
       ε = 10**16 / 2**63 (rounding error in 64-bit arithmetic)
    or ε' = 10**16 / 2**(63+53)
    which gives about (1e16)*(2**54)*ε' = 2e13 candidates.

    If we focus on rounding error at 96-bit precision,
    (± 1e16 / 2**(96+53)) which yields about 5000 candidates.
    """
    if e2 < prec:
        r1 = Rat(2**prec - 1, 10**e10 * 2**(prec-e2), bound=2**mantbits)
        r2 = Rat(2**prec + 1, 10**e10 * 2**(prec-e2), bound=2**mantbits)
    else:
        r1 = Rat(2**e2 - 2**(e2-prec), 10**e10, bound=2**mantbits)
        r2 = Rat(2**e2 + 2**(e2-prec), 10**e10, bound=2**mantbits)
    #print("bounds 2**{}/10**{}: {}/{} -> {}/{}".format(
    #    e2, e10, n1, d1, n2, d2))
    for x, y in walk(r1, r2, bound=2**mantbits):
        digs, mant = x, y
        # try odd multiples
        if mant % 2 == 1:
            m = mant
            d = digs
            while m.bit_length() <= mantbits:
                if m.bit_length() == mantbits:
                    decimal = str(m << e2)
                    if e10 > 20:
                        decimal = decimal[:20-e10] + "..."
                    print('{:17} {:17}e+{:03} {:17}p+{} = {:>45}'.format(
                        m, d, e10, m, e2, decimal))
                m += 2*mant
                d += 2*digs
        else:
            #print('{:17} {:17}'.format(mant, digs))
            pass

        # epsilon
        #epsilon = (2**e2 * mant - 10**e10 * digs) / (10**e10 * mant)
        #print("epsilon =", epsilon)

def find_hard_parse_negexp(e2, e10, mantbits=54, prec=96, denormal=False):
    """
    Like find_hard_parse but for negative exponents

    We look for:
        mantissa / 2**e2 = digits / 10**e10 + ε
        10**e10 / 2**e2 = digits / mantissa + ε'
    """
    if e10 < prec:
        # multiply by 2**(prec-e10)
        r1 = Rat(10**e10 * 2**(prec-e10) - 5**e10, 2**(e2+prec-e10), bound=2**mantbits)
        r2 = Rat(10**e10 * 2**(prec-e10) + 5**e10, 2**(e2+prec-e10), bound=2**mantbits)
    else:
        r1 = Rat(10**e10 - (10**e10 >> prec), 2**e2, bound=2**mantbits)
        r2 = Rat(10**e10 + (10**e10 >> prec), 2**e2, bound=2**mantbits)

    for x, y in walk(r1, r2, bound=2**mantbits):
        digs, mant = x, y
        # try odd multiples
        if mant % 2 == 1:
            m = mant
            d = digs
            while m.bit_length() <= mantbits:
                if denormal or m.bit_length() == mantbits:
                    decimal = str(m * 5**e2)
                    if e2 > 30:
                        trim = ((e2-30)*7) // 10
                        decimal = decimal[:-trim]
                    print('{:17} {:17}e-{:03d} {:17}p-{} = {:>45}...'.format(
                        m, d, e10, m, e2, decimal))
                m += 2*mant
                d += 2*digs

def find_hard_print(e2, e10, mantbits=53, prec=96):
    """
    Like find_hard_parse but now we are looking for:

        mantissa × 2**e2 × 10**-e10 = digits + 1/2 + ε

    where ε is very small.

    The fractions we are looking for are:

        (2*digits+1) / (2*mantissa)
    """
    BOUND = 2**(1+mantbits)

    if e2 < prec:
        r1 = Rat(2**prec - 1, 10**e10 * 2**(prec-e2), bound=BOUND)
        r2 = Rat(2**prec + 1, 10**e10 * 2**(prec-e2), bound=BOUND)
    else:
        r1 = Rat(2**e2 - 2**(e2-prec), 10**e10, bound=BOUND)
        r2 = Rat(2**e2 + 2**(e2-prec), 10**e10, bound=BOUND)
    #print("bounds 2**{}/10**{}: {}/{} -> {}/{}".format(
    #    e2, e10, n1, d1, n2, d2))
    for x, y in walk(r1, r2, bound=BOUND):
        digs, mant = x, y
        # try odd multiples
        if mant & 1 == 0 and digs & 1 == 1:
            m = mant
            d = digs
            while m.bit_length() <= mantbits+1:
                if m.bit_length() == mantbits+1:
                    decimal = str(m << (e2-1))
                    if e10 > 20:
                        decimal = decimal[:20-e10] + "..."
                    print('{:17} {:17}e+{:03} {:17}p+{} = {:>45}'.format(
                        m // 2, d // 2, e10, m // 2, e2, decimal))
                m += 2*mant
                d += 2*digs

        # epsilon
        #epsilon = (2**e2 * mant - 10**e10 * digs) / (10**e10 * mant)
        #print("epsilon =", epsilon)

def find_hard_print_negexp(e2, e10, mantbits=53, prec=96, denormal=False):
    BOUND = 2**(1+mantbits)

    if e10 < prec:
        # multiply by 2**(prec-e10)
        r1 = Rat(10**e10 * 2**(prec-e10) - 5**e10, 2**(e2+prec-e10), bound=BOUND)
        r2 = Rat(10**e10 * 2**(prec-e10) + 5**e10, 2**(e2+prec-e10), bound=BOUND)
    else:
        r1 = Rat(10**e10 - (10**e10 >> prec), 2**e2, bound=BOUND)
        r2 = Rat(10**e10 + (10**e10 >> prec), 2**e2, bound=BOUND)

    for x, y in walk(r1, r2, bound=BOUND):
        digs, mant = x, y
        # try odd multiples
        if mant & 1 == 0 and digs & 1 == 1:
            m = mant
            d = digs
            while m.bit_length() <= mantbits+1:
                if denormal or m.bit_length() == mantbits+1:
                    decimal = str((m//2) * 5**e2)
                    if e2 > 36:
                        trim = ((e2-30)*7) // 10
                        decimal = decimal[:-trim] + "..."
                    print('{:17} {:17}e-{:03} {:17}p-{} = {:>45}'.format(
                        m // 2, d // 2, e10, m // 2, e2, decimal))
                m += 2*mant
                d += 2*digs


def walk(r1, r2, bound):
    """
    Walk enumerates fractions between r1 and r2
    for a given denominator bound (the Farey sequence).

    >>> list(walk(Rat(1, 4), Rat(1, 2), bound=8))
    [(1, 4), (2, 7), (1, 3), (3, 8), (2, 5), (3, 7), (1, 2)]

    In this example, the tree is:
                ___  1/3 (0,3) ____
    1/4 (0, 4) ´                   ` 2/5 (0,2,2)
       ` 2/7 (0,3,2)  3/8 (0,2,1,2)´             ` 3/7 (0,2,3)

    >>> list(walk(Rat(1, 4), Rat(1, 3), bound=12))
    [(1, 4), (3, 11), (2, 7), (3, 10), (1, 3)]
    >>> list(walk(Rat(1, 4), Rat(2, 7), bound=32))
    [(1, 4), (8, 31), (7, 27), (6, 23), (5, 19), (4, 15), (7, 26), (3, 11), (8, 29), (5, 18), (7, 25), (9, 32), (2, 7)]
    >>> len(list(walk(Rat(65352703432539, 79006570561214),
    ...          Rat(34807131698651, 42079240217226), bound=2**48)))
    39930
    """

    # FIXME: a couple of duplicates seem to appear
    l = r1.clone()
    yield l.fraction()
    while l.fraction() != r2.fraction():
        l = l.next(bound)
        yield l.fraction()

class Rat:
    def __init__(self, num, den, bound=None):
        """
        Computes a continued fraction expansion of a / b.
        If bound is not None, the expansion stops at the
        largest denominator <= bound.

        >>> Rat(89, 55).cont
        [1, 1, 1, 1, 1, 1, 1, 1, 2]
        >>> Rat(355, 113).cont
        [3, 7, 16]
        >>> Rat(30, 20).cont
        [1, 2]

        >>> Rat(89, 55, bound=20).fraction()
        (21, 13)
        >>> Rat(355, 113, bound=200).fraction()
        (355, 113)
        >>> Rat(355, 113, bound=100).fraction()
        (22, 7)
        """
        self.cont = []
        # Invariant: (p*a + q*b, r*a + s*b)
        p, q = 1, 0
        r, s = 0, 1
        a, b = num, den
        while b > 1:
            quo = a // b
            rem = a % b
            a, b = b, rem
            # (p*quo + q)*b + p*rem = p*a + q*b
            p, q = p*quo + q, p
            r, s = r*quo + s, r
            if bound is not None and r > bound:
                return
            self.cont.append(quo)

        if b == 1:
            p, q = p*a + q, p
            r, s = r*a + s, r
            if bound is not None and r > bound:
                return
            self.cont.append(a)

    def __eq__(self, other):
        if not isinstance(other, Rat):
            return NotImplemented
        return self.cont == other.cont

    def __neq__(self, other):
        if not isinstance(other, Rat):
            return NotImplemented
        return self.cont != other.cont

    @staticmethod
    def from_expansion(l):
        r = Rat(0, 1)
        r.cont = l
        return r

    def clone(self):
        return Rat.from_expansion(self.cont.copy())

    def fraction(self):
        """
        >>> Rat.from_expansion([3, 7, 16]).fraction()
        (355, 113)
        >>> Rat.from_expansion([1, 1, 1, 1, 1, 1, 1, 1, 1, 1]).fraction()
        (89, 55)
        """
        a, b = 1, 0
        for q in reversed(self.cont):
            a, b = a*q + b, a
        return a, b

    def child(self, i):
        """
        Sets self to its left (i=0) or right (i=1) child
        in the Stern-Brocot tree.

        The children of (a0, ..., an) in the tree are:
            (a0, ..., an+1)
            (a0, ..., an-1, 2)

        >>> r = Rat.from_expansion([0, 3])
        >>> r.child(1).cont
        [0, 2, 2]
        """
        expansion = self.cont
        if len(expansion) % 2 == i:
            expansion[-1] += 1
        else:
            expansion[-1] -= 1
            expansion.append(2)
        self.cont = expansion
        return self

    def child_fraction(self, i):
        """
        self.child_fraction(i) == self.child(i).fraction()

        >>> r = Rat(123456, 456789)
        >>> r.child_fraction(0) == r.clone().child(0).fraction()
        True
        >>> r.child_fraction(1) == r.clone().child(1).fraction()
        True
        """
        a, b = 1, 0
        c, d = 0, 1
        for idx in range(len(self.cont)-1):
            a, b = a * self.cont[idx] + b, a
            c, d = c * self.cont[idx] + d, c
        if len(self.cont) % 2 == i:
            # last + 1
            q = self.cont[-1] + 1
            a, b = a * q + b, a
            c, d = c * q + d, c
        else:
            # (last - 1, 2)
            q = self.cont[-1] - 1
            a, b = a * q + b, a
            c, d = c * q + d, c
            a, b = a * 2 + b, a
            c, d = c * 2 + d, c
        return a, c

    def next_up(self):
        """
        next_up returns the next element in the Stern-Brocot tree
        which is not a child.

        That is, next_up(N) is the nearest ancestor A of N
        such that N is a child of A->left

        >>> Rat.from_expansion([1, 1, 1, 2]).next_up().cont
        [1, 1, 2]
        >>> Rat.from_expansion([1, 1, 1, 1, 2]).next_up().cont
        [1, 1, 2]
        >>> Rat.from_expansion([1, 1, 1, 1, 256]).next_up().cont
        [1, 1, 2]
        >>> Rat.from_expansion([1, 1, 1, 1, 256123456789]).next_up().cont
        [1, 1, 2]
        >>> Rat.from_expansion([2]).next_up().cont
        [3]
        >>> Rat.from_expansion([0, 3, 2]).next_up().cont
        [0, 3]
        """

        exp = self.cont
        if len(exp) == 1:
            # exception for integers
            exp = [exp[0]+1]
        elif len(exp) % 2 == 0:
            #                     (..k+1)
            #          ..(..k, 2)´
            # (..k, n)´
            last = exp[-1]
            if last > 2:
                exp[-1] -= 1
            else:
                exp = exp[:-1]
                exp[-1] += 1
        else:
            # parent obtained by decrementing n = exp[-1]
            # is always smaller. Let's assume n=2
            #         _______________(..k)
            # (..k+1)´
            #    `(..k, 2)
            #            `... (..k, n)
            exp = exp[:-1]
            if exp[-1] == 1:
                # normalize
                exp = exp[:-1]
                exp[-1] += 1

        self.cont = exp
        return self

    def next(self, bound):
        """
        The next item in the Farey sequence of bound bound.
        """
        # Peek right child
        num, den = self.child_fraction(1)
        if den > bound:
            return self.next_up()
        else:
            # go down the left branch
            self.child(1)
            while True:
                n2, d2 = self.child_fraction(0)
                if d2 <= bound:
                    self.child(0)
                else:
                    return self

if __name__ == "__main__":
    main()
