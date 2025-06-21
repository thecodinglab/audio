# Fast Fourier Transform

A Fast Fourier Transform is an algorithm that computes the [Discrete Fourier
Transform](./dft.md) of a sequence using a more efficient algorithm.

## Cooley–Tukey FFT algorithm

The Cooley–Tukey FFT rearranges the DFT algorithm into two parts: a sum over
the even-numbered indices and a sum over the odd-numbered indices.

```math
X_k 
= \sum^{N-1}_{n=0} x_n \cdot e^{-2 i \pi \frac{k}{N} n}
= \sum^{\frac{N}{2}-1}_{m=0} x_{2m} \cdot e^{-2 i \pi \frac{k}{N} (2m)}
+ \sum^{\frac{N}{2}-1}_{m=0} x_{2m+1} \cdot e^{-2 i \pi \frac{k}{N} (2m+1)}
```

It is possible to factor a common multiplier $e^{-2 i \pi \frac{k}{N}}$ out of
the odd-numbered indices part:

```math
\begin{aligned}
\sum^{\frac{N}{2}-1}_{m=0} x_{2m+1} \cdot e^{-2 i \pi \frac{k}{N} (2m+1)}
&= \sum^{\frac{N}{2}-1}_{m=0} x_{2m+1} \cdot e^{-2 i \pi \frac{k}{N} (2m) - 2 i \pi \frac{k}{N}} \\
&= \sum^{\frac{N}{2}-1}_{m=0} x_{2m+1} \cdot e^{-2 i \pi \frac{k}{N} (2m)} \cdot e^{-2 i \pi \frac{k}{N}} \\
&= e^{-2 i \pi \frac{k}{N}} \cdot \sum^{\frac{N}{2}-1}_{m=0} x_{2m+1} \cdot e^{-2 i \pi \frac{k}{N} (2m)}
\end{aligned}
```

To simplify the expression, we define $E_k$ to be the sum over the
even-numbered indices and $O_k$ the sum over the odd-numbered indices.

```math
X_k 
= \underbrace{\sum^{\frac{N}{2}-1}_{m=0} x_{2m} \cdot e^{-2 i \pi \frac{k}{N} (2m)}}_{\text{DFT of even-indexed part } E_k}
+ e^{-2 i \pi \frac{k}{N}} \cdot \underbrace{\sum^{\frac{N}{2}-1}_{m=0} x_{2m+1} \cdot e^{-2 i \pi \frac{k}{N} (2m)}}_{\text{DFT of odd-indexed part } O_k}
= E_k + e^{-2 i \pi \frac{k}{N}} O_k
```

Furthermore, the operations are almost symmetric:

```math
\begin{aligned}
X_{k+N/2}

&= \sum^{\frac{N}{2}-1}_{m=0} x_{2m} \cdot e^{-2 i \pi \frac{k + N/2}{N} (2m)}
+ e^{-2 i \pi \frac{k + N/2}{N}} \cdot \sum^{\frac{N}{2}-1}_{m=0} x_{2m+1} \cdot e^{-2 i \pi \frac{k + N/2}{N} (2m)} \\

&= \sum^{\frac{N}{2}-1}_{m=0} x_{2m} \cdot e^{-2 i \pi \frac{k}{N} (2m)} \cdot \underbrace{e^{-2 i \pi m}}_{= 1 \text{, as } m \in \mathbb{N}}
+ e^{-2 i \pi \frac{k}{N}} \cdot \underbrace{e^{-i \pi}}_{= -1}
\cdot \sum^{\frac{N}{2}-1}_{m=0} x_{2m+1} \cdot e^{-2 i \pi \frac{k}{N} (2m)} \cdot \underbrace{e^{-2 i \pi m}}_{= 1 \text{, as } m \in \mathbb{N}} \\

&= \sum^{\frac{N}{2}-1}_{m=0} x_{2m} \cdot e^{-2 i \pi \frac{k}{N} (2m)}
- e^{-2 i \pi \frac{k}{N}} \cdot \sum^{\frac{N}{2}-1}_{m=0} x_{2m+1} \cdot e^{-2 i \pi \frac{k}{N} (2m)} \\

&= E_k - e^{-2 i \pi \frac{k}{N}} O_k
\end{aligned}
```

Thus, it is possible to rewrite:

```math
\begin{aligned}
X_{k} &= E_k + e^{-2 i \pi \frac{k}{N}} O_k \\
X_{k+\frac{N}{2}} &= E_k - e^{-2 i \pi \frac{k}{N}} O_k
\end{aligned}
```

This, results in a recursive divide and conquer algorithm with complexity $O(n
\cdot log_2(n))$:

```go
import (
	"math"
	"math/cmplx"
)

func FFT(in, out []complex128) {
	radix2(in, out, n, 1)
}

func radix2(in, out []complex128, n, stride int) {
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
```
