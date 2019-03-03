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
    import argparse
    p = argparse.ArgumentParser()
    p.add_argument("MODE", choices=("", "print64+", "print64-"))
    args = p.parse_args()
    arg = args.MODE

    if not arg or arg == "print64+":
        # 680k values
        for e2 in range(50, 1024-52):
            e10 = (e2 * 78913) >> 18
            find_hard_printf(e2, e10+1)

    if not arg or arg == "print64-":
        # 600k values
        for e2 in range(20, 1024+52):
            e10 = (e2 * 78913) >> 18
            find_hard_printf_negexp(e2, e10)

def find_hard_printf(e2, e10):
    """
    e.g. find floating-point numbers with exponent 385 hard to print.

    For example: 8640368759831959p+385

    The midpoint (8640368759831959 + 1/2) * 1<<385
    is 68089572682806429.999999999999999e115
    so it is hard at 16 digits.

    We are looking for:
        mantissa × 2**385 × 10**-116 = digits + ε
    where digits < 1e16
          mantissa < 2**54
          mantissa is odd

    that is:
        2**385 / 10**116 = digits / mantissa + ε'
    
    The typical threshold we are interested in is
       ε = 10**16 / 2**63 (rounding error in 64-bit arithmetic)
    or ε' = 10**16 / 2**(63+53)
    which gives about (1e16)*(2**54)*ε' = 2e13 candidates.

    Let's focus on rounding error at 96-bit precision,
    (± 1e16 / 2**(96+53)) which yields about 5000 candidates.
    """
    if e2 < 96:
        n1, d1 = approx(2**96 - 1, 10**e10 * 2**(96-e2), bound=2**54)
        n2, d2 = approx(2**96 + 1, 10**e10 * 2**(96-e2), bound=2**54)
    else:
        n1, d1 = approx(2**e2 - 2**(e2-96), 10**e10, bound=2**54)
        n2, d2 = approx(2**e2 + 2**(e2-96), 10**e10, bound=2**54)
    #print("bounds 2**{}/10**{}: {}/{} -> {}/{}".format(
    #    e2, e10, n1, d1, n2, d2))
    for x, y in walk(n1, d1, n2, d2, bound=2**54):
        digs, mant = x, y
        # try odd multiples
        if mant % 2 == 1:
            m = mant
            d = digs
            while m.bit_length() <= 54:
                if m.bit_length() == 54:
                    decimal = str(m << e2)
                    if e10 > 20:
                        decimal = decimal[:20-e10] + "..."
                    print('{:17} {:17} {:17}p+{} = {:>45}'.format(
                        m, d, m, e2, decimal))
                m += 2*mant
                d += 2*digs
        else:
            #print('{:17} {:17}'.format(mant, digs))
            pass

        # epsilon
        #epsilon = (2**e2 * mant - 10**e10 * digs) / (10**e10 * mant)
        #print("epsilon =", epsilon)

def find_hard_printf_negexp(e2, e10):
    """
    Like find_hard_printf but for negative exponents

    We look for:
        mantissa / 2**e2 = digits / 10**e10 + ε
        10**e10 / 2**e2 = digits / mantissa + ε'
    """
    if e10 < 96:
        # multiply by 2**(96-e10)
        n1, d1 = approx(10**e10 * 2**(96-e10) - 5**e10, 2**(e2+96-e10), bound=2**54)
        n2, d2 = approx(10**e10 * 2**(96-e10) + 5**e10, 2**(e2+96-e10), bound=2**54)
    else:
        n1, d1 = approx(10**e10 - (10**e10 >> 96), 2**e2, bound=2**54)
        n2, d2 = approx(10**e10 + (10**e10 >> 96), 2**e2, bound=2**54)

    for x, y in walk(n1, d1, n2, d2, bound=2**54):
        digs, mant = x, y
        # try odd multiples
        if mant % 2 == 1:
            m = mant
            d = digs
            while m.bit_length() <= 54:
                if m.bit_length() == 54:
                    decimal = str(m * 5**e2)
                    if e2 > 30:
                        trim = ((e2-30)*7) // 10
                        decimal = decimal[:-trim]
                    print('{:17} {:17} {:17}p-{} = {:>45}...'.format(
                        m, d, m, e2, decimal))
                m += 2*mant
                d += 2*digs

def walk(x1, y1, x2, y2, bound):
    """
    Walk enumerates fractions between x1/y2 and x2/y2
    for a given denominator bound (the Farey sequence).

    >>> list(walk(1, 4, 1, 2, bound=8))
    [(1, 4), (2, 7), (1, 3), (3, 8), (2, 5), (3, 7), (1, 2)]

    In this example, the tree is:
                ___  1/3 (0,3) ____
    1/4 (0, 4) ´                   ` 2/5 (0,2,2)
       ` 2/7 (0,3,2)  3/8 (0,2,1,2)´             ` 3/7 (0,2,3)
    """
    start = expand_cont_frac(x1, y1)
    end = expand_cont_frac(x2, y2)
    l = start
    yield x1, y1
    while l != end:
        lr = right(l.copy())
        num, den = to_fraction(lr)
        if den > bound:
            l = next_up(l)
            yield to_fraction(l)
        else:
            # go down the left branch
            l = lr
            while True:
                l2 = left(l.copy())
                n2, d2 = to_fraction(l2)
                if d2 <= bound:
                    l, num, den = l2, n2, d2
                else:
                    break
            yield num, den

def expand_cont_frac(a, b):
    """
    Computes a continued fraction expansion of a / b.
    The result is a list of integers.

    >>> expand_cont_frac(89, 55)
    [1, 1, 1, 1, 1, 1, 1, 1, 2]
    >>> expand_cont_frac(355, 113)
    [3, 7, 16]
    >>> expand_cont_frac(30, 20)
    [1, 2]
    """
    exp = []
    while b > 1:
        q = a // b
        exp.append(q)
        r = a % b
        a, b = b, r

    if b == 1:
        exp.append(a)

    return exp

def approx(a, b, bound):
    """
    Computes a good approximation to a / b
    within the provided denominator bound.

    This is not necessarily the best approximation for that bound.

    >>> approx(89, 55, 20)
    (21, 13)
    >>> approx(355, 113, 200)
    (355, 113)
    >>> approx(355, 113, 100)
    (22, 7)
    """
    # Invariant: (p*a + q*b, r*a + s*b)
    p, q = 1, 0
    r, s = 0, 1

    while b > 1:
        quo = a // b
        rem = a % b
        a, b = b, rem
        # (p*quo + q)*b + p*rem = p*a + q*b
        p, q = p*quo + q, p
        r, s = r*quo + s, r
        if r > bound:
            return q, s

    if b == 1:
        p, q = p*a + q, p
        r, s = r*a + s, r
        if r > bound:
            return q, s

    return p, r

def to_fraction(expansion):
    """
    >>> to_fraction([3, 7, 16])
    (355, 113)
    >>> to_fraction([1, 1, 1, 1, 1, 1, 1, 1, 1, 1])
    (89, 55)
    """
    a, b = 1, 0
    for q in reversed(expansion):
        a, b = a*q + b, a
    return a, b

def left(expansion):
    if len(expansion) % 2 == 0:
        expansion[-1] += 1
    else:
        expansion[-1] -= 1
        expansion.append(2)
    return expansion

def right(expansion):
    """
    next returns the right child in the Stern-Brocot tree.
    The argument may be mutated.

    The children of (a0, ..., an) in the tree are:
        (a0, ..., an+1)
        (a0, ..., an-1, 2)

    >>> right([0, 3])
    [0, 2, 2]
    """
    if len(expansion) % 2 == 1:
        expansion[-1] += 1
    else:
        expansion[-1] -= 1
        expansion.append(2)
    return expansion

def next_up(expansion):
    """
    next_up returns the next element in the Stern-Brocot tree
    which is not a child.

    That is, next_up(N) is the nearest ancestor A of N
    such that N is a child of A->left

    >>> next_up([1, 1, 1, 2])
    [1, 1, 2]
    >>> next_up([1, 1, 1, 1, 2])
    [1, 1, 2]
    >>> next_up([1, 1, 1, 1, 256])
    [1, 1, 2]
    >>> next_up([1, 1, 1, 1, 256123456789])
    [1, 1, 2]
    >>> next_up([2])
    [3]
    >>> next_up([0, 3, 2])
    [0, 3]
    """

    exp = expansion
    if len(exp) == 1:
        return [exp[0]+1]
    elif len(exp) % 2 == 0:
        #                     (..k+1)
        #          ..(..k, 2)´
        # (..k, n)´
        last = exp[-1]
        if last > 2:
            exp[-1] -= 1
            return exp
        else:
            exp = exp[:-1]
            exp[-1] += 1
            return exp
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
        return exp

if __name__ == "__main__":
    main()
