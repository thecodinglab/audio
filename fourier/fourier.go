package fourier

import (
	"math"
	"math/cmplx"
)

// DFT computes the discrete fourier transform of the given input. It's
// computation time is O(n^2) and thus should only be used for small inputs.
//
// Note: the input & output arrays must be of the same size.
//
// Formula: Xₖ = ∑ⁿ⁻¹ⱼ₌₀ inᵢ·e^(-2iπ·jk/n)
func DFT(in, out []complex128) {
	n := len(in)
	ang := -2 * math.Pi / float64(n)
	for k := range n {
		out[k] = 0
		for j := range n {
			out[k] += in[j] * cmplx.Rect(1, ang*float64(k)*float64(j))
		}
	}
}

func FFT(in, out []complex128) {
	radix2(in, out, len(in), 1)
}

// radix2 computes the fast fourier transform of the given input using the
// Cooley–Tukey FFT algorithm.
func radix2(in, out []complex128, n, stride int) {
	if !isPowerOfTwo(n) {
		panic("radix2 fft must be called with n=2^s")
	}

	if n == 1 {
		out[0] = in[0]
		return
	}

	radix2(in, out, n/2, 2*stride)
	radix2(in[stride:], out[n/2:], n/2, 2*stride)

	ang := -2 * math.Pi / float64(n)
	for k := range n / 2 {
		tf := cmplx.Rect(1, ang*float64(k)) * out[k+n/2]
		out[k], out[k+n/2] = out[k]+tf, out[k]-tf
	}
}

func isPowerOfTwo(x int) bool {
	return x != 0 && x&(x-1) == 0
}
