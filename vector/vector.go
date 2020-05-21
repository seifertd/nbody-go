package vector

import (
	"math"
)

type Vector struct {
	X float64
	Y float64
	Z float64
}

func New2DVector(x, y float64) Vector {
	return Vector{x, y, 0.0}
}
func Add(v1, v2 Vector) Vector {
	return Vector{v1.X + v2.X, v1.Y + v2.Y, v1.Z + v2.Z}
}
func Sub(v1, v2 Vector) Vector {
	return Vector{v1.X - v2.X, v1.Y - v2.Y, v1.Z - v2.Z}
}
func MultScalar(v Vector, scalar float64) Vector {
	return Vector{v.X * scalar, v.Y * scalar, v.Z * scalar}
}
func DivScalar(v Vector, scalar float64) Vector {
	return Vector{v.X / scalar, v.Y / scalar, v.Z / scalar}
}

func (v Vector) Magnitude() float64 {
	return math.Sqrt(math.Pow(v.X, 2) + math.Pow(v.Y, 2) + math.Pow(v.Z, 2))
}
func (v Vector) Unit() Vector {
	mag := v.Magnitude()
	return Vector{v.X / mag, v.Y / mag, v.Z / mag}
}
func (v Vector) Normal2D() Vector {
	return Vector{-v.Y, v.X, 0.0}
}
func (v Vector) Dot(other Vector) float64 {
	return v.X*other.X + v.Y*other.Y + v.Z*other.Z
}
func (v *Vector) Add(other Vector) {
	v.X += other.X
	v.Y += other.Y
	v.Z += other.Z
}
func (v *Vector) Sub(other Vector) {
	v.X -= other.X
	v.Y -= other.Y
	v.Z -= other.Z
}
func (v *Vector) DivScalar(scalar float64) {
	v.X /= scalar
	v.Y /= scalar
	v.Z /= scalar
}
func (v *Vector) MultScalar(scalar float64) {
	v.X *= scalar
	v.Y *= scalar
	v.Z *= scalar
}
