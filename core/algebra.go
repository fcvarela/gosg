package core

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
)

// Mat4DoubleToFloat converts an mgl64.Mat4 into an mgl32.Mat4
func Mat4DoubleToFloat(d mgl64.Mat4) mgl32.Mat4 {
	var out mgl32.Mat4

	for i := 0; i < 16; i++ {
		out[i] = float32(d[i])
	}

	return out
}

// Vec4DoubleToFloat converts an mgl64.Vec4 into an mgl32.Vec4
func Vec4DoubleToFloat(d mgl64.Vec4) mgl32.Vec4 {
	var out mgl32.Vec4

	for i := 0; i < 4; i++ {
		out[i] = float32(d[i])
	}

	return out
}

// Vec3DoubleToFloat converts an mgl64.Vec3 into an mgl32.Vec3
func Vec3DoubleToFloat(d mgl64.Vec3) mgl32.Vec3 {
	var out mgl32.Vec3

	for i := 0; i < 3; i++ {
		out[i] = float32(d[i])
	}

	return out
}

// Vec2DoubleToFloat converts an mgl64.Vec2 into an mgl32.Vec2
func Vec2DoubleToFloat(d mgl64.Vec2) mgl32.Vec2 {
	var out mgl32.Vec2

	for i := 0; i < 2; i++ {
		out[i] = float32(d[i])
	}

	return out
}

// Clamp constrains the passed value to lie between two other values
func Clamp(val, c0, c1 float64) float64 {
	switch {
	case val < c0:
		return c0
	case val > c1:
		return c1
	default:
		return val
	}
}

// SmoothStep performs hermite interpolation between two values
func SmoothStep(from, to, t float64) float64 {
	t = Clamp((t-from)/(to-from), 0.0, 1.0)
	return (t * t) * (3.0 - 2.0*t)
}
