# Discrete Fourier Transform

The discrete Fourier transform converts a sequence of samples in time space,
into its representation in the frequency domain.

$$
	X_k = \sum^{N-1}_{n=0} x_n \cdot e^{-2 i \pi \frac{k}{N} n}
$$

Explanation of the formula:
- https://www.youtube.com/watch?v=spUNpyF58BY
- https://www.youtube.com/watch?v=nmgFG7PUHfo

The simplest way to implement a DFT is to simply loop through all elements in
the sequence twice (i.e. $O(n^2)$) as per the definition of the formula above:

```go
import (
	"math"
	"math/cmplx"
)

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
```

For an optimized implementation, take a look at [Fast Fourier Fransforms](./fft.md).
