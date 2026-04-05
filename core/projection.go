package core

import (
	"math"

	"github.com/go-gl/mathgl/mgl64"
)

// PerspectiveWebGPU creates a perspective projection matrix for WebGPU (Z range [0,1]).
func PerspectiveWebGPU(fovY, aspect, near, far float64) mgl64.Mat4 {
	f := 1.0 / math.Tan(fovY/2.0)
	rangeInv := 1.0 / (near - far)

	return mgl64.Mat4{
		f / aspect, 0, 0, 0,
		0, f, 0, 0,
		0, 0, far * rangeInv, -1,
		0, 0, near * far * rangeInv, 0,
	}
}

// OrthoWebGPU creates an orthographic projection matrix for WebGPU (Z range [0,1]).
func OrthoWebGPU(left, right, bottom, top, near, far float64) mgl64.Mat4 {
	rml := right - left
	tmb := top - bottom
	fmn := far - near

	return mgl64.Mat4{
		2.0 / rml, 0, 0, 0,
		0, 2.0 / tmb, 0, 0,
		0, 0, -1.0 / fmn, 0,
		-(right + left) / rml, -(top + bottom) / tmb, -near / fmn, 1,
	}
}
