package vector

import (
	"fmt"
	"testing"
)

func TestVector(t *testing.T) {
	v := Vector{1, 2, 3}
	if v.X != 1 {
		t.Errorf("Vector did not store X coordinate")
	}
	if v.Y != 2 {
		t.Errorf("Vector did not store Y coordinate")
	}
	if v.Z != 3 {
		t.Errorf("Vector did not store Z coordinate")
	}
}
func TestVectorMag(t *testing.T) {
	v := Vector{2, 3, 6}
	if v.Magnitude() != 7 {
		t.Errorf("Vector{2,3,6}.Magnitude() != 7")
	}
}

func ExampleVectorMagnitude() {
	v := New2DVector(3, 4)
	fmt.Printf("Vector x=%v,y=%v mag=%v\n", v.X, v.Y, v.Magnitude())
	// Output: Vector x=3,y=4 mag=5
}

func ExampleVectorMagnitude2() {
	v := New2DVector(3, 4)
	fmt.Printf("Vector %+v mag=%v\n", v, v.Magnitude())
	// Output: Vector {X:3 Y:4 Z:0} mag=5
}

func ExampleVectorUnit() {
	v := New2DVector(3, 4)
	u := v.Unit()
	fmt.Printf("UnitVector x=%v,y=%v mag=%v\n", u.X, u.Y, u.Magnitude())
	// Output: UnitVector x=0.6,y=0.8 mag=1
}

func ExampleVectorNormal() {
	v := New2DVector(3, 4)
	n2d := v.Normal2D()
	fmt.Printf("2DNormal x=%v,y=%v mag=%v\n", n2d.X, n2d.Y, n2d.Magnitude())
	// Output: 2DNormal x=-4,y=3 mag=5
}

func ExampleVector3d() {
	v2 := Vector{12, 13, 22}
	fmt.Printf("Vector2 x=%v,y=%v,z=%v mag=%v\n", v2.X, v2.Y, v2.Z, v2.Magnitude())
	v3 := Vector{6, -5, 12}
	fmt.Printf("Vector3 x=%v,y=%v,z=%v mag=%v\n", v3.X, v3.Y, v3.Z, v3.Magnitude())
	fmt.Printf("v2 . v3 = %v\n", v2.Dot(v3))

	sumv1v2 := Add(v2, v3)
	fmt.Printf("v2+v3 x=%v,y=%v,z=%v mag=%v\n", sumv1v2.X, sumv1v2.Y, sumv1v2.Z, sumv1v2.Magnitude())

	diffv1v2 := Sub(v2, v3)
	fmt.Printf("v2-v3 x=%v,y=%v,z=%v mag=%v\n", diffv1v2.X, diffv1v2.Y, diffv1v2.Z, diffv1v2.Magnitude())

	v2.MultScalar(0.5)
	fmt.Printf("Vector2*0.5 x=%v,y=%v,z=%v mag=%v\n", v2.X, v2.Y, v2.Z, v2.Magnitude())

	v3.DivScalar(1.1)
	fmt.Printf("Vector3/1.1 x=%v,y=%v,z=%v mag=%v\n", v3.X, v3.Y, v3.Z, v3.Magnitude())

	v2.Add(v3)
	fmt.Printf("Vector2+Vector3 x=%v,y=%v,z=%v mag=%v\n", v2.X, v2.Y, v2.Z, v2.Magnitude())

	v2.Sub(sumv1v2)
	fmt.Printf("Vector2-something x=%v,y=%v,z=%v mag=%v\n", v2.X, v2.Y, v2.Z, v2.Magnitude())
}
