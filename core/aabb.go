package core

import (
	"fmt"
	"math"

	"github.com/go-gl/mathgl/mgl64"
)

// AABB defines the smallest enclosing box for a given set of points
type AABB struct {
	min mgl64.Vec3
	max mgl64.Vec3
}

// NewAABB returns a ready to use AABB which contains no volume
func NewAABB() *AABB {
	b := AABB{}
	b.max = mgl64.Vec3{math.Inf(-1), math.Inf(-1), math.Inf(-1)}
	b.min = mgl64.Vec3{math.Inf(+1), math.Inf(+1), math.Inf(+1)}
	return &b
}

// Reset returns the AABB to its initial state
func (a *AABB) Reset() {
	a.max = mgl64.Vec3{math.Inf(-1), math.Inf(-1), math.Inf(-1)}
	a.min = mgl64.Vec3{math.Inf(+1), math.Inf(+1), math.Inf(+1)}
}

// Center returns the center point of the AABB
func (a *AABB) Center() mgl64.Vec3 {
	return a.min.Add(a.max).Mul(0.5)
}

// Min returns the min point
func (a *AABB) Min() mgl64.Vec3 {
	return a.min
}

// Max returns the min point
func (a *AABB) Max() mgl64.Vec3 {
	return a.max
}

// Size returns the size of the AABB
func (a *AABB) Size() mgl64.Vec3 {
	return a.max.Sub(a.min)
}

// ExtendWithPoint extends the bounding box to contain the given point
func (a *AABB) ExtendWithPoint(ip mgl64.Vec3) {
	for i := range ip {
		a.min[i] = math.Min(a.min[i], ip[i])
		a.max[i] = math.Max(a.max[i], ip[i])
	}
}

// DistanceToPoint returns the minimum distance from a point to any of the AABBs corners
func (a *AABB) DistanceToPoint(p mgl64.Vec3) (float64, float64) {
	var points [8]mgl64.Vec3
	points[0] = a.min
	points[1] = a.max
	points[2] = mgl64.Vec3{points[0][0], points[0][1], points[1][2]}
	points[3] = mgl64.Vec3{points[0][0], points[1][1], points[0][2]}
	points[4] = mgl64.Vec3{points[1][0], points[0][1], points[0][2]}
	points[5] = mgl64.Vec3{points[0][0], points[1][1], points[1][2]}
	points[6] = mgl64.Vec3{points[1][0], points[0][1], points[1][2]}
	points[7] = mgl64.Vec3{points[1][0], points[1][1], points[0][2]}

	var min, max = math.Inf(+1), math.Inf(-1)
	for i := range points {
		var d = points[i].Sub(p).Len()
		if d < min {
			min = d
		}
		if d > max {
			max = d
		}
	}

	return min, max
}

// ExtendWithBox extends the bounding box to contain the volume defined by
// the given bounding box
func (a *AABB) ExtendWithBox(ib *AABB) {
	a.ExtendWithPoint(ib.min)
	a.ExtendWithPoint(ib.max)
}

// String implements the stringer interface
func (a *AABB) String() string {
	return fmt.Sprintf("AABB min: %v max: %v", a.min, a.max)
}

// ContainsBox returns whether the bounding volume fully contains the volume
// defined by the given bounding box
func (a *AABB) ContainsBox(ib *AABB) bool {
	return a.ContainsPoint(ib.min) && a.ContainsPoint(ib.max)
}

// ContainsPoint returns whether the bounding volume contains the given point
func (a *AABB) ContainsPoint(ip mgl64.Vec3) bool {
	gt := ip[0] >= a.min[0] && ip[1] >= a.min[1] && ip[2] >= a.min[2]
	lt := ip[0] <= a.max[0] && ip[1] <= a.max[1] && ip[2] <= a.max[2]
	return gt && lt
}

// InFrustum returns whether the bounding volume is contained in the frustum
// defined by the given planes
func (a *AABB) InFrustum(planes [6]mgl64.Vec4) bool {
	insideOrIntersect := true
	var vmin, vmax mgl64.Vec3

	for i := 0; i < 6; i++ {
		// X axis
		if planes[i][0] > 0 {
			vmin[0] = a.min[0]
			vmax[0] = a.max[0]
		} else {
			vmin[0] = a.max[0]
			vmax[0] = a.min[0]
		}
		// Y axis
		if planes[i][1] > 0 {
			vmin[1] = a.min[1]
			vmax[1] = a.max[1]
		} else {
			vmin[1] = a.max[1]
			vmax[1] = a.min[1]
		}
		// Z axis
		if planes[i][2] > 0 {
			vmin[2] = a.min[2]
			vmax[2] = a.max[2]
		} else {
			vmin[2] = a.max[2]
			vmax[2] = a.min[2]
		}

		if planes[i].Vec3().Dot(vmax)+planes[i][3] < 0 {
			return false
		}
		if planes[i].Vec3().Dot(vmin)+planes[i][3] <= 0 {
			insideOrIntersect = true
		}
	}
	return insideOrIntersect
}

// Transformed returns a new ABB which contains the volume specified by
// the eight corners or the original AABB when multiplied by the passed transform.
// This can be used, for example, to transform an AABB in object space (OBB) to
// an AABB in world space.
func (a *AABB) Transformed(m mgl64.Mat4) *AABB {
	// create a new box
	newAABB := NewAABB()

	var points = [8]mgl64.Vec3{
		mgl64.Vec3{a.min[0], a.min[1], a.min[2]},
		mgl64.Vec3{a.max[0], a.min[1], a.min[2]},
		mgl64.Vec3{a.min[0], a.max[1], a.min[2]},
		mgl64.Vec3{a.max[0], a.max[1], a.min[2]},
		mgl64.Vec3{a.min[0], a.min[1], a.max[2]},
		mgl64.Vec3{a.max[0], a.min[1], a.max[2]},
		mgl64.Vec3{a.min[0], a.max[1], a.max[2]},
		mgl64.Vec3{a.max[0], a.max[1], a.max[2]},
	}

	for _, p := range points {
		min := mgl64.TransformCoordinate(p, m)
		max := mgl64.TransformCoordinate(p, m)
		newAABB.ExtendWithPoint(min)
		newAABB.ExtendWithPoint(max)
	}

	// return the new box
	return newAABB
}
