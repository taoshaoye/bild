package bild

import (
	"image"
	"math"
)

// ConvolutionMatrix interface for use as an image Kernel
type ConvolutionMatrix interface {
	At(x, y int) float64
	Sum() float64
	Normalized() ConvolutionMatrix
	Diameter() int
}

// NewKernel returns a kernel of the provided size
func NewKernel(diameter int) *Kernel {
	matrix := make([][]float64, diameter)
	for i := 0; i < diameter; i++ {
		matrix[i] = make([]float64, diameter)
	}
	return &Kernel{matrix}
}

// Kernel is used as a convolution matrix
type Kernel struct {
	Matrix [][]float64
}

// Sum returns the cumulative value of the matrix
func (k *Kernel) Sum() float64 {
	var sum float64
	diameter := k.Diameter()
	for x := 0; x < diameter; x++ {
		for y := 0; y < diameter; y++ {
			sum += k.Matrix[x][y]
		}
	}
	return sum
}

// Normalized returns a new Kernel with normalized values
func (k *Kernel) Normalized() ConvolutionMatrix {
	sum := k.Sum()
	diameter := k.Diameter()
	nk := NewKernel(diameter)

	for x := 0; x < diameter; x++ {
		for y := 0; y < diameter; y++ {
			nk.Matrix[x][y] = k.Matrix[x][y] / sum
		}
	}

	return nk
}

// Diameter returns the row/column length for the kernel
func (k *Kernel) Diameter() int {
	return len(k.Matrix)
}

// At returns the matrix value at position x, y
func (k *Kernel) At(x, y int) float64 {
	return k.Matrix[x][y]
}

// convolute applies a convolution matrix (kernel) to an image
func convolute(img image.Image, k ConvolutionMatrix) *image.RGBA {
	bounds := img.Bounds()
	src := CloneAsRGBA(img)
	dst := image.NewRGBA(bounds)

	w, h := bounds.Max.X, bounds.Max.Y
	diameter := k.Diameter()
	normKernel := k.Normalized()

	parallelize(h, func(start, end int) {
		for x := 0; x < w; x++ {
			for y := start; y < end; y++ {

				var r, g, b, a float64
				for kx := 0; kx < diameter; kx++ {
					for ky := 0; ky < diameter; ky++ {
						ix := x - diameter/2 + kx
						iy := y - diameter/2 + ky

						if ix < 0 || ix >= w || iy < 0 || iy >= h {
							continue
						}

						ipos := iy*dst.Stride + ix*4
						kvalue := normKernel.At(kx, ky)
						r += float64(src.Pix[ipos+0]) * kvalue
						g += float64(src.Pix[ipos+1]) * kvalue
						b += float64(src.Pix[ipos+2]) * kvalue
						a += float64(src.Pix[ipos+3]) * kvalue
					}
				}

				pos := y*dst.Stride + x*4

				dst.Pix[pos+0] = uint8(math.Max(math.Min(r, 255), 0))
				dst.Pix[pos+1] = uint8(math.Max(math.Min(g, 255), 0))
				dst.Pix[pos+2] = uint8(math.Max(math.Min(b, 255), 0))
				dst.Pix[pos+3] = uint8(math.Max(math.Min(a, 255), 0))
			}
		}
	})

	return dst
}
