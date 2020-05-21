package body

import (
	"testing"
)

func TestBodyCollisions(t *testing.T) {
	b1 := NewBody("b1", 0, 0, 10, 10, 0, 0)
	b2 := NewBody("b2", 0, 0, 10, 5, 5, 5)
	toDelete := b1.CollideWith(b2)
	if toDelete != b2 {
		t.Errorf("b2 should be absorbed by b1 in collision")
	}
	if b1.Mass != 15 {
		t.Errorf("b2 should absorb b1's mass: %v", b1.Mass)
	}
	if b1.Vel.X != (5.0 / 3.0) {
		t.Errorf("b2 should conserve momentum in x dir: %v != %v", b1.Vel.X, 5.0/3.0)
	}
	if b1.Vel.Y != (5.0 / 3.0) {
		t.Errorf("b2 should conserve momentum in x dir: %v != %v", b1.Vel.Y, 5.0/3.0)
	}
	if b1.Name != "b1<-b2" {
		t.Errorf("b2's name should incorporate b1: %v", b1.Name)
	}
}

func TestBodyCollisionTesting(t *testing.T) {
	b1 := NewBody("b1", 0, 0, 10, 20, 0, 0)
	b2 := NewBody("b2", 100, 100, 10, 20, 0, 0)
	b3 := NewBody("b3", 0, 15, 7, 20, 0, 0)
	b4 := NewBody("b4", 0, 15, 5, 20, 0, 0)

	if b1.Collides(b2) {
		t.Errorf("b1 and b2 should not be colliding")
	}
	if !b1.Collides(b4) {
		dx := b1.Pos.X - b4.Pos.X
		dy := b1.Pos.Y - b4.Pos.Y
		sr := b1.Radius + b4.Radius
		t.Errorf("b1 and b4 should be colliding: dx = %v dy = %v r1+r2 = %v, %v = %v (%v)",
			dx, dy, sr, dx*dx+dy*dy, sr*sr, dx*dx+dy*dy-sr*sr <= 0)
	}
	if !b1.Collides(b3) {
		t.Errorf("b1 and b3 should be colliding")
	}
}
