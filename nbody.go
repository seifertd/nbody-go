package main

import (
	crypto_rand "crypto/rand"
	"dseifert.net/nbody/body"
	"dseifert.net/nbody/vector"
	"encoding/binary"
	"fmt"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
	"image"
	_ "image/png"
	"math"
	math_rand "math/rand"
	"os"
	"time"
)

const (
	G          = 6.674e-11
	MinRadius  = 1.0
)

func initRand() {
	var b [8]byte
	_, err := crypto_rand.Read(b[:])
	if err != nil {
		panic("Unable to seed math/rand package with secure random number generator")
	}
	math_rand.Seed(int64(binary.LittleEndian.Uint64(b[:])))
}

func loadPicture(path string) (pixel.Picture, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return pixel.PictureDataFromImage(img), nil
}

type World struct {
	scale   float64
	mpp     float64
	spt     int
	running bool
	elapsed int
	bodies  []*body.Body
	width   int
	height  int
}

type BodyPair struct {
	body1 *body.Body
	body2 *body.Body
}

func (w World) worldToScreen(coords *vector.Vector) vector.Vector {
   return vector.Vector{coords.X / w.mpp * w.scale, coords.Y / w.mpp * w.scale, 0}
}

func (w World) worldTime() string {
	d := w.elapsed / (3600 * 24)
	h := (w.elapsed % (3600 * 24)) / 3600
	m := (w.elapsed % 3600) / 60
	s := w.elapsed % 60
	return fmt.Sprintf("%dd %02dh%02dm%02ds", d, h, m, s)
}

func (w *World) calculateAcceleration(body *body.Body, c chan vector.Vector) {
	deltaA := vector.Vector{0, 0, 0}
	for _, body2 := range w.bodies {
		if body == body2 {
			continue
		}
		d := math.Sqrt(math.Pow(body.Pos.X-body2.Pos.X, 2) + math.Pow(body.Pos.Y-body2.Pos.Y, 2))
		acc := vector.New2DVector((body2.Pos.X-body.Pos.X)/d, (body2.Pos.Y-body.Pos.Y)/d)
		acc.MultScalar(G * body2.Mass / (d * d))
		deltaA.Add(acc)
	}
	c <- deltaA
}

func (w *World) tick() {
	for i := 0; i < w.spt; i++ {
		w.elapsed += 1
		for _, body := range w.bodies {
			go body.CalculateAcceleration(w.bodies)
		}

		// Integrate and check for collisions
		var colliding []*BodyPair
		alreadyColliding := func(body1, body2 *body.Body) bool {
			for i := 0; i < len(colliding); i++ {
				pair := colliding[i]
				if (pair.body1 == body1 && pair.body2 == body2) ||
					(pair.body2 == body1 && pair.body1 == body2) {
					return true
				}
			}
			return false
		}

		for _, body := range w.bodies {
			deltaA := <-body.AccChan
			body.Acc.X = deltaA.X
			body.Acc.Y = deltaA.Y
			body.Vel.Add(body.Acc)
			body.Pos.Add(body.Vel)
			for _, body2 := range w.bodies {
				if body == body2 {
					continue
				}
				if body.Collides(body2) && !alreadyColliding(body, body2) {
					colliding = append(colliding, &BodyPair{body, body2})
				}
			}
		}
		for _, pair := range colliding {
			var keeping *body.Body
			deleting := pair.body1.CollideWith(pair.body2)
			if deleting == pair.body1 {
				keeping = pair.body2
			} else {
				keeping = pair.body1
			}
			fmt.Printf("%v: COLLISION: %v\n", w.worldTime(), keeping)
			newBodies := w.bodies[:0]
			for _, x := range w.bodies {
				if x != deleting {
					newBodies = append(newBodies, x)
				}
			}
			// Clean up remaining
			for i := len(newBodies); i < len(w.bodies); i++ {
				w.bodies[i] = nil
			}
			w.bodies = newBodies
		}
	}
}

func iPow(a, b int) int {
	var result int = 1

	for 0 != b {
		if 0 != (b & 1) {
			result *= a

		}
		b >>= 1
		a *= a
	}

	return result
}

func randomWorld(w, h, n int, pf float64, dense bool) *World {
	world := &World{
		scale:   0.3,
		mpp:     5e5,
		spt:     60,
		running: true,
		elapsed: 0,
		bodies:  make([]*body.Body, n+1),
		width:   w,
		height:  h,
	}
	world.bodies[0] = body.NewBody("Mother", 0, 0, 30*world.mpp, 5e28, 0, 0)
	fmt.Printf("%v\n", world.bodies[0])
	center := world.bodies[0]
	maxDistance := math.Sqrt(float64(iPow(world.width, 2) + iPow(world.height, 2))) / 2.0
	if dense {
		maxDistance *= 0.3
	}
	for i := 1; i < n+1; i++ {
		distance := 200.0 + math_rand.Float64()*maxDistance
		theta := math_rand.Float64() * math.Pi * 2
		pos := vector.New2DVector(-distance*math.Cos(theta)*world.mpp, -distance*math.Sin(theta)*world.mpp)
		circularOrbitVel := math.Sqrt(G * center.Mass / pos.Magnitude())
		u := pos.Unit()
		un := u.Normal2D()
		vel := vector.Vector{un.X, un.Y, un.Z}
		vel.MultScalar(circularOrbitVel)

		vel.X *= (1.0 - (pf / 2.0) + math_rand.Float64()*pf)
		vel.Y *= (1.0 - (pf / 2.0) + math_rand.Float64()*pf)

		baseMass := 1e22
		baseRadius := 10.0
		if i > n/2 {
			baseMass = 1e7
			baseRadius = 4.0
		}
		world.bodies[i] = body.NewBodyVector(fmt.Sprintf("P%v", i), pos, vel,
			(1.0+math_rand.Float64())*baseRadius*world.mpp,
			baseMass*math_rand.Float64())
		fmt.Printf("%v\n", world.bodies[i])
	}
	return world
}

func run() {
	initRand()
	world := randomWorld(800, 800, 40, 0.5, false)
	cfg := pixelgl.WindowConfig{
		Title:  "N-Body Problem",
		Bounds: pixel.R(0, 0, 800, 800),
		VSync:  true,
	}

	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	// initialize font
	basicAtlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
  infoTxt := text.New(pixel.V(600,780), basicAtlas)

  // initialize sprites
	sunPic, err := loadPicture("sun.png")
	if err != nil {
		panic(err)
	}
	planetPic, err := loadPicture("planetsheet.png")
	if err != nil {
		panic(err)
	}
	sunSprite := pixel.NewSprite(sunPic, sunPic.Bounds())
	var planetSprites []*pixel.Sprite
	for x := planetPic.Bounds().Min.X; x < planetPic.Bounds().Max.X; x += 128 {
		for y := planetPic.Bounds().Min.Y; y < planetPic.Bounds().Max.Y; y += 128 {
			planetSprites = append(planetSprites, pixel.NewSprite(planetPic, pixel.R(x, y, x+128, y+128)))
		}
	}
  bodySprites := make(map[string]*pixel.Sprite)
	for bi, body := range world.bodies {
		if bi == 0 {
			bodySprites[body.Id] = sunSprite
		} else {
			idx := math_rand.Intn(len(planetSprites))
			bodySprites[body.Id] = planetSprites[idx]
		}
	}

	for !win.Closed() {
		// zoom
		world.scale *= math.Pow(1.2, win.MouseScroll().Y)
		win.Clear(colornames.Black)
		mat := pixel.IM

		for _, body := range world.bodies {
			var spriteSize float64
			if body.Id == "Mother" {
				spriteSize = 45.0
			} else {
				spriteSize = 128.0
			}
			sprite := bodySprites[body.Id]
			brp := body.Radius * world.scale / world.mpp
			if brp < MinRadius {
			  brp = MinRadius
			}
			sf := brp / spriteSize
			bodyMat := mat.ScaledXY(pixel.ZV, pixel.V(sf, sf))
			screenPos := world.worldToScreen(&body.Pos)
			screenPos.Add(vector.Vector{win.Bounds().Center().X, win.Bounds().Center().Y, 0})
			bodyMat = bodyMat.Moved(pixel.V(screenPos.X, screenPos.Y))
			sprite.Draw(win, bodyMat)
		}
		// Update info text
		infoTxt.Clear()
		fmt.Fprintf(infoTxt, "N: %v\n", len(world.bodies))
		fmt.Fprintf(infoTxt, "t: %v\n", world.worldTime())
		fmt.Fprintf(infoTxt, "S: %4.2f", world.scale)
		infoTxt.Draw(win, pixel.IM)
		win.Update()
		world.tick()
	}
}

func main() {
	pixelgl.Run(run)
}

func testMain() {
	world := randomWorld(800, 800, 40, 0.5, false)
	fmt.Printf("Created world with %v bodies\n", len(world.bodies))
	start := time.Now()
	for j := 0; j < 1440*7; j++ {
		world.tick()
	}
	for _, body := range world.bodies {
		fmt.Printf("%v\n", body)
	}
	elapsed := time.Since(start)
	fmt.Printf("1 week of sim took %v real time", elapsed)
}
