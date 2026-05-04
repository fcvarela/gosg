package core

import (
	"math"
	"testing"

	"github.com/go-gl/mathgl/mgl64"
)

func TestAABB_NewAABB(t *testing.T) {
	a := NewAABB()
	if !math.IsInf(a.min[0], +1) || !math.IsInf(a.min[1], +1) || !math.IsInf(a.min[2], +1) {
		t.Error("NewAABB min should be +Inf")
	}
	if !math.IsInf(a.max[0], -1) || !math.IsInf(a.max[1], -1) || !math.IsInf(a.max[2], -1) {
		t.Error("NewAABB max should be -Inf")
	}
}

func TestAABB_ExtendWithPoint(t *testing.T) {
	a := NewAABB()
	a.ExtendWithPoint(mgl64.Vec3{1, 2, 3})
	a.ExtendWithPoint(mgl64.Vec3{4, 5, 6})

	if a.min != (mgl64.Vec3{1, 2, 3}) {
		t.Errorf("min = %v, want [1 2 3]", a.min)
	}
	if a.max != (mgl64.Vec3{4, 5, 6}) {
		t.Errorf("max = %v, want [4 5 6]", a.max)
	}
}

func TestAABB_Center(t *testing.T) {
	a := NewAABB()
	a.ExtendWithPoint(mgl64.Vec3{0, 0, 0})
	a.ExtendWithPoint(mgl64.Vec3{10, 10, 10})

	center := a.Center()
	expected := mgl64.Vec3{5, 5, 5}
	if center != expected {
		t.Errorf("center = %v, want %v", center, expected)
	}
}

func TestAABB_Size(t *testing.T) {
	a := NewAABB()
	a.ExtendWithPoint(mgl64.Vec3{1, 1, 1})
	a.ExtendWithPoint(mgl64.Vec3{5, 5, 5})

	size := a.Size()
	expected := mgl64.Vec3{4, 4, 4}
	if size != expected {
		t.Errorf("size = %v, want %v", size, expected)
	}
}

func TestAABB_ContainsPoint(t *testing.T) {
	a := NewAABB()
	a.ExtendWithPoint(mgl64.Vec3{0, 0, 0})
	a.ExtendWithPoint(mgl64.Vec3{10, 10, 10})

	tests := []struct {
		point mgl64.Vec3
		want  bool
	}{
		{mgl64.Vec3{5, 5, 5}, true},
		{mgl64.Vec3{0, 0, 0}, true},
		{mgl64.Vec3{10, 10, 10}, true},
		{mgl64.Vec3{-1, 5, 5}, false},
		{mgl64.Vec3{11, 5, 5}, false},
	}

	for _, tt := range tests {
		if got := a.ContainsPoint(tt.point); got != tt.want {
			t.Errorf("ContainsPoint(%v) = %v, want %v", tt.point, got, tt.want)
		}
	}
}

func TestAABB_ContainsBox(t *testing.T) {
	outer := NewAABB()
	outer.ExtendWithPoint(mgl64.Vec3{0, 0, 0})
	outer.ExtendWithPoint(mgl64.Vec3{10, 10, 10})

	inner := NewAABB()
	inner.ExtendWithPoint(mgl64.Vec3{2, 2, 2})
	inner.ExtendWithPoint(mgl64.Vec3{8, 8, 8})

	if !outer.ContainsBox(inner) {
		t.Error("outer should contain inner")
	}

	larger := NewAABB()
	larger.ExtendWithPoint(mgl64.Vec3{-1, -1, -1})
	larger.ExtendWithPoint(mgl64.Vec3{11, 11, 11})

	if outer.ContainsBox(larger) {
		t.Error("outer should not contain larger box")
	}
}

func TestAABB_Transformed(t *testing.T) {
	a := NewAABB()
	a.ExtendWithPoint(mgl64.Vec3{-1, -1, -1})
	a.ExtendWithPoint(mgl64.Vec3{1, 1, 1})

	identity := mgl64.Ident4()
	transformed := a.Transformed(identity)

	if transformed.min != a.min || transformed.max != a.max {
		t.Error("identity transform should preserve AABB")
	}

	translate := mgl64.Translate3D(10, 10, 10)
	transformed = a.Transformed(translate)

	expectedMin := mgl64.Vec3{9, 9, 9}
	expectedMax := mgl64.Vec3{11, 11, 11}
	if transformed.min != expectedMin || transformed.max != expectedMax {
		t.Errorf("translated AABB min=%v max=%v, want min=%v max=%v",
			transformed.min, transformed.max, expectedMin, expectedMax)
	}
}

func TestAABB_DistanceToPoint(t *testing.T) {
	a := NewAABB()
	a.ExtendWithPoint(mgl64.Vec3{0, 0, 0})
	a.ExtendWithPoint(mgl64.Vec3{2, 2, 2})

	minDist, maxDist := a.DistanceToPoint(mgl64.Vec3{1, 1, 1})

	if minDist > 0.001 {
		t.Errorf("minDist = %v, want ~0", minDist)
	}
	if maxDist < 1.7 {
		t.Errorf("maxDist = %v, want ~1.73", maxDist)
	}
}

func TestAABB_Reset(t *testing.T) {
	a := NewAABB()
	a.ExtendWithPoint(mgl64.Vec3{0, 0, 0})
	a.ExtendWithPoint(mgl64.Vec3{10, 10, 10})

	a.Reset()

	if !math.IsInf(a.min[0], +1) {
		t.Error("Reset should restore min to +Inf")
	}
	if !math.IsInf(a.max[0], -1) {
		t.Error("Reset should restore max to -Inf")
	}
}

func TestPerspectiveWebGPU(t *testing.T) {
	fov := math.Pi / 4.0 // 45 degrees
	aspect := 16.0 / 9.0
	near := 0.1
	far := 100.0

	m := PerspectiveWebGPU(fov, aspect, near, far)

	// Test that a point at the center of the near plane maps to (0, 0, 0) in NDC
	nearCenter := mgl64.Vec4{0, 0, -near, 1}
	clipCenter := m.Mul4x1(nearCenter)
	if clipCenter[3] == 0 {
		t.Fatal("w component should not be zero")
	}
	ndcCenter := mgl64.Vec3{
		clipCenter[0] / clipCenter[3],
		clipCenter[1] / clipCenter[3],
		clipCenter[2] / clipCenter[3],
	}

	if math.Abs(ndcCenter[0]) > 0.001 || math.Abs(ndcCenter[1]) > 0.001 {
		t.Errorf("near center should map to (0, 0, z) in NDC, got %v", ndcCenter)
	}
	if ndcCenter[2] < -0.001 || ndcCenter[2] > 0.001 {
		t.Errorf("near center z should be ~0 in WebGPU NDC, got %v", ndcCenter[2])
	}
}

func TestOrthoWebGPU(t *testing.T) {
	m := OrthoWebGPU(-10, 10, -10, 10, 0.1, 100)

	// Test that center maps to origin in NDC
	center := mgl64.Vec4{0, 0, -50, 1}
	clipCenter := m.Mul4x1(center)
	ndcCenter := mgl64.Vec3{
		clipCenter[0] / clipCenter[3],
		clipCenter[1] / clipCenter[3],
		clipCenter[2] / clipCenter[3],
	}

	if math.Abs(ndcCenter[0]) > 0.001 || math.Abs(ndcCenter[1]) > 0.001 {
		t.Errorf("center should map to (0, 0, z) in NDC, got %v", ndcCenter)
	}
}

func TestMakeFrustum(t *testing.T) {
	proj := PerspectiveWebGPU(math.Pi/4, 1.0, 0.1, 100)
	view := mgl64.Ident4()

	frustum := MakeFrustum(proj, view)

	for i, plane := range frustum {
		len := plane.Vec3().Len()
		if math.Abs(len-1.0) > 0.001 {
			t.Errorf("frustum plane %d not normalized: len=%v", i, len)
		}
	}
}

func TestAABB_InFrustum(t *testing.T) {
	a := NewAABB()
	a.ExtendWithPoint(mgl64.Vec3{-1, -1, -1})
	a.ExtendWithPoint(mgl64.Vec3{1, 1, 1})

	proj := PerspectiveWebGPU(math.Pi/4, 1.0, 0.1, 100)
	view := mgl64.Ident4()
	frustum := MakeFrustum(proj, view)

	// AABB at origin should be in frustum
	if !a.InFrustum(frustum) {
		t.Error("AABB at origin should be in frustum")
	}

	// Move AABB far away
	far := mgl64.Translate3D(0, 0, -200)
	farAABB := a.Transformed(far)
	if farAABB.InFrustum(frustum) {
		t.Error("AABB far beyond far plane should not be in frustum")
	}
}
