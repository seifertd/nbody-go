package body

import (
	"dseifert.net/nbody/vector"
	"fmt"
	"math"
)

type Body struct {
	Name   string
	Pos    vector.Vector
	Vel    vector.Vector
	Acc    vector.Vector
	Radius float64
	Mass   float64
}

func NewBody(name string, x float64, y float64, r float64, m float64,
	vx float64, vy float64) *Body {
	return &Body{name, vector.New2DVector(x, y), vector.New2DVector(vx, vy),
		vector.New2DVector(0, 0), r, m}
}

func (b Body) String() string {
	return fmt.Sprintf("BODY: %v: m:%v vel:%v,%v pos:%v,%v r:%v",
		b.Name, b.Mass, b.Vel.X, b.Vel.Y, b.Pos.X, b.Pos.Y, b.Radius)
}

func (b Body) Collides(other *Body) bool {
	if &b == other {
		return false
	}
	dx := b.Pos.X - other.Pos.X
	dy := b.Pos.Y - other.Pos.Y
	r2 := b.Radius + other.Radius
	return dx*dx+dy*dy-r2*r2 <= 0
}

func (b *Body) CollideWith(other *Body) *Body {
	var biggest, smallest *Body
	if b.Mass > other.Mass {
		biggest = b
		smallest = other
	} else {
		biggest = other
		smallest = b
	}
	nr := math.Pow(math.Pow(biggest.Radius, 3)+math.Pow(smallest.Radius, 3), 1.0/3.0)
	vnx := (biggest.Mass*biggest.Vel.X + smallest.Mass*smallest.Vel.X) /
		(biggest.Mass + smallest.Mass)
	vny := (biggest.Mass*biggest.Vel.Y + smallest.Mass*smallest.Vel.Y) /
		(biggest.Mass + smallest.Mass)
	biggest.Radius = nr
	biggest.Mass += smallest.Mass
	biggest.Vel.X = vnx
	biggest.Vel.Y = vny
	biggest.Name = fmt.Sprintf("%v<-%v", biggest.Name, smallest.Name)
	return smallest
}
