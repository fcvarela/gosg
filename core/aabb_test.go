package core

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl64"
)

func ExampleAABB_Transformed() {
	var a AABB

	a.ExtendWithPoint(mgl64.Vec3{-10.0, -10.0, -10.0})
	a.ExtendWithPoint(mgl64.Vec3{10.0, 10.0, 10.0})

	transform := mgl64.QuatRotate(1.5, mgl64.Vec3{0.0, 1.0, 0.0}).Mat4()
	fmt.Println(transform)
}
