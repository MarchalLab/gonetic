package powerlaws

import (
	"math"
)

// ePSILON is a sufficiently small value
const ePSILON = 1e-12

// m contains expansion coefficients for Euler-Maclaurin summation formula (2k)! / B2k where B2k are Bernoulli numbers
var m = []float64{
	12.0,
	-720.0,
	30240.0,
	-1209600.0,
	47900160.0,
	-1.8924375803183791606e9, // 1.307674368e12/691
	7.47242496e10,
	-2.950130727918164224e12,  // 1.067062284288e16/3617
	1.1646782814350067249e14,  // 5.109094217170944e18/43867
	-4.5979787224074726105e15, // 8.028576626982912e20/174611
	1.8152105401943546773e17,  // 1.5511210043330985984e23/854513
	-7.1661652561756670113e18, // 1.6938241367317436694528e27/236364091
}

// HurwitzZeta implements the generalized zeta function for x >= 1.0
// implementation ported from the scipy/special/cephes library (scipy/special/cephes/zeta.c)
func HurwitzZeta(x, q float64) float64 {
	// check arguments
	if x == 1.0 {
		return math.MaxFloat64
	}
	if x < 1.0 {
		// not implemented
		return math.NaN()
	}
	if q <= 0.0 {
		if q == math.Floor(q) {
			// hurwitz zeta does not converge for negative integer values of q
			return math.NaN()
		}
		if x != math.Floor(x) {
			// q ^ -x is not real
			return math.NaN()
		}
	}
	if q > 1e8 {
		// asymptotic expansion, https://dlmf.nist.gov/25.11#E43
		return (1/(x-1) + 1/(2*q)) * math.Pow(q, 1-x)
	}

	// permit negative q but continue sum until n+q > 9
	// zeta(x, q) = zeta(x, q + 1) + q ^ {-x}
	s := math.Pow(q, -x)
	a := q
	i := 0
	b := 0.0
	for (i < 9) || (a <= 9.0) {
		i += 1
		a += 1.0
		b = math.Pow(a, -x)
		s += b
		if math.Abs(b/s) < ePSILON {
			return s
		}
	}

	// Euler-Maclaurin summation formula
	w := a
	s += b * w / (x - 1.0)
	s -= 0.5 * b
	a = 1.0
	k := 0.0
	for i := 0; i < 12; i++ {
		a *= x + k
		b /= w
		t := a * b / m[i]
		s = s + t
		t = math.Abs(t / s)
		if t < ePSILON {
			return s
		}
		k += 1.0
		a *= x + k
		b /= w
		k += 1.0
	}
	return s
}
