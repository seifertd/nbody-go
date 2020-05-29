package body

import (
	"fmt"
	"github.com/faiface/pixel"
	"github.com/seifertd/nbody-go/vector"
	"math"
)

const G = 6.674e-11

type Body struct {
	Id      string
	Name    string
	Pos     vector.Vector
	Vel     vector.Vector
	Acc     vector.Vector
	Radius  float64
	Mass    float64
	AccChan chan vector.Vector
	Sprite  *pixel.Sprite
}

func NewBody(name string, x float64, y float64, r float64, m float64,
	vx float64, vy float64, s *pixel.Sprite) *Body {
	return &Body{name, name, vector.New2DVector(x, y), vector.New2DVector(vx, vy),
		vector.New2DVector(0, 0), r, m, make(chan vector.Vector), s}
}
func NewBodyVector(name string, pos vector.Vector, vel vector.Vector,
	r float64, m float64, s *pixel.Sprite) *Body {
	return &Body{name, name, pos, vel, vector.New2DVector(0, 0),
		r, m, make(chan vector.Vector), s}
}

func (b Body) String() string {
	//return b.Name
	return fmt.Sprintf("BODY: %v: m:%v vel:%v,%v pos:%v,%v r:%v",
		b.Name, b.Mass, b.Vel.X, b.Vel.Y, b.Pos.X, b.Pos.Y, b.Radius)
}

func (b *Body) CalculateAcceleration(others []*Body) {
	deltaA := vector.Vector{0, 0, 0}
	for _, body2 := range others {
		if b == body2 {
			continue
		}
		d := math.Sqrt(math.Pow(b.Pos.X-body2.Pos.X, 2) + math.Pow(b.Pos.Y-body2.Pos.Y, 2))
		acc := vector.New2DVector((body2.Pos.X-b.Pos.X)/d, (body2.Pos.Y-b.Pos.Y)/d)
		acc.MultScalar(G * body2.Mass / (d * d))
		deltaA.Add(acc)
	}
	b.AccChan <- deltaA
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

func (b *Body) CollideWith(other *Body) {
	// Assume other is going away
	nr := math.Pow(math.Pow(b.Radius, 3)+math.Pow(other.Radius, 3), 1.0/3.0)
	vnx := (b.Mass*b.Vel.X + other.Mass*other.Vel.X) /
		(b.Mass + other.Mass)
	vny := (b.Mass*b.Vel.Y + other.Mass*other.Vel.Y) /
		(b.Mass + other.Mass)
	b.Radius = nr
	b.Mass += other.Mass
	b.Vel.X = vnx
	b.Vel.Y = vny
	b.Name = fmt.Sprintf("%v<-%v", b.Name, other.Name)
}
