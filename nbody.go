package main

import (
	crypto_rand "crypto/rand"
	"encoding/binary"
	"fmt"
	"github.com/docopt/docopt-go"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"github.com/seifertd/go/vector"
	"github.com/seifertd/nbody-go/body"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
	"image"
	_ "image/png"
	"math"
	math_rand "math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	G         = 6.674e-11
	MinRadius = 4.0
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

var (
	sprites          map[string]*pixel.Sprite
	numPlanetSprites int
	circleMode       bool
)

func loadSprite(name, path string) *pixel.Sprite {
	if sprites == nil {
		sprites = make(map[string]*pixel.Sprite)
		numPlanetSprites = 0
	}
	if sprite, contains := sprites[name]; contains {
		return sprite
	} else {
		pic, err := loadPicture(path)
		if err != nil {
			panic(err)
		}
		sprite = pixel.NewSprite(pic, pic.Bounds())
		sprites[name] = sprite
		return sprite
	}
}

func randomPlanetSprite() *pixel.Sprite {
	if circleMode {
		return sprites["circle"]
	}
	i := math_rand.Intn(numPlanetSprites + 1) // include an extra for Earth
	if i == numPlanetSprites {
		return sprites["earth"]
	} else {
		name := fmt.Sprintf("planet%v", math_rand.Intn(numPlanetSprites))
		return sprites[name]
	}
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
	mag     float64
}

func (w World) worldToScreen(coords *vector.Vector) vector.Vector {
	return vector.Vector{coords.X / w.mpp * w.scale * w.mag, coords.Y / w.mpp * w.scale * w.mag, 0}
}

func (w World) worldTime() string {
	d := w.elapsed / (3600 * 24)
	h := (w.elapsed % (3600 * 24)) / 3600
	m := (w.elapsed % 3600) / 60
	s := w.elapsed % 60
	return fmt.Sprintf("%dd %02dh%02dm%02ds", d, h, m, s)
}

func (w World) escaped(body *body.Body) bool {
	sun := w.bodies[0]
	radius := body.Pos.DistanceTo(sun.Pos)
	maxDistance := math.Sqrt(float64(iPow(w.width, 2)+iPow(w.height, 2))) * 10.0 * w.mag * w.mpp

	return radius > maxDistance && body.Vel.Magnitude() > math.Sqrt(2.0*G*sun.Mass/radius)
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

func (w *World) removeBody(toRemove *body.Body) {
	newBodies := w.bodies[:0]
	for _, x := range w.bodies {
		if x != toRemove {
			newBodies = append(newBodies, x)
		}
	}
	// Clean up remaining
	for i := len(newBodies); i < len(w.bodies); i++ {
		w.bodies[i] = nil
	}
	w.bodies = newBodies
}

func (w *World) tick() {
	if !w.running {
		return
	}

	for i := 0; i < w.spt; i++ {
		w.elapsed += 1

		// Store initial positions and velocities
		initialPos := make([]vector.Vector, len(w.bodies))
		initialVel := make([]vector.Vector, len(w.bodies))

		// RK4 requires 4 sets of derivatives (k1, k2, k3, k4)
		k1Vel := make([]vector.Vector, len(w.bodies))
		k1Acc := make([]vector.Vector, len(w.bodies))
		k2Vel := make([]vector.Vector, len(w.bodies))
		k2Acc := make([]vector.Vector, len(w.bodies))
		k3Vel := make([]vector.Vector, len(w.bodies))
		k3Acc := make([]vector.Vector, len(w.bodies))
		k4Vel := make([]vector.Vector, len(w.bodies))
		k4Acc := make([]vector.Vector, len(w.bodies))

		// Save initial state
		for i, body := range w.bodies {
			initialPos[i] = vector.Vector{body.Pos.X, body.Pos.Y, body.Pos.Z}
			initialVel[i] = vector.Vector{body.Vel.X, body.Vel.Y, body.Vel.Z}
		}

		// Step 1: Calculate k1 (derivatives at the beginning of the interval)
		w.calculateAllAccelerations(k1Acc)
		for i, body := range w.bodies {
			k1Vel[i] = vector.Vector{body.Vel.X, body.Vel.Y, body.Vel.Z}
		}

		// Step 2: Calculate k2 (derivatives at the midpoint using k1)
		w.applyHalfStep(initialPos, initialVel, k1Vel, k1Acc)
		w.calculateAllAccelerations(k2Acc)
		for i, body := range w.bodies {
			k2Vel[i] = vector.Vector{body.Vel.X, body.Vel.Y, body.Vel.Z}
		}

		// Step 3: Calculate k3 (derivatives at the midpoint using k2)
		w.resetToInitial(initialPos, initialVel)
		w.applyHalfStep(initialPos, initialVel, k2Vel, k2Acc)
		w.calculateAllAccelerations(k3Acc)
		for i, body := range w.bodies {
			k3Vel[i] = vector.Vector{body.Vel.X, body.Vel.Y, body.Vel.Z}
		}

		// Step 4: Calculate k4 (derivatives at the end using k3)
		w.resetToInitial(initialPos, initialVel)
		w.applyFullStep(initialPos, initialVel, k3Vel, k3Acc)
		w.calculateAllAccelerations(k4Acc)
		for i, body := range w.bodies {
			k4Vel[i] = vector.Vector{body.Vel.X, body.Vel.Y, body.Vel.Z}
		}

		// Apply the weighted sum to get final positions and velocities
		w.resetToInitial(initialPos, initialVel)

		for i, body := range w.bodies {
			// Update velocity: v(t+dt) = v(t) + (1/6) * (k1 + 2*k2 + 2*k3 + k4)
			velChange := vector.Vector{0, 0, 0}
			velChange.Add(k1Acc[i])

			temp := vector.Vector{k2Acc[i].X, k2Acc[i].Y, k2Acc[i].Z}
			temp.MultScalar(2)
			velChange.Add(temp)

			temp = vector.Vector{k3Acc[i].X, k3Acc[i].Y, k3Acc[i].Z}
			temp.MultScalar(2)
			velChange.Add(temp)

			velChange.Add(k4Acc[i])
			velChange.MultScalar(1.0 / 6.0)

			body.Vel.Add(velChange)

			// Update position: x(t+dt) = x(t) + (1/6) * (k1 + 2*k2 + 2*k3 + k4)
			posChange := vector.Vector{0, 0, 0}
			posChange.Add(k1Vel[i])

			temp = vector.Vector{k2Vel[i].X, k2Vel[i].Y, k2Vel[i].Z}
			temp.MultScalar(2)
			posChange.Add(temp)

			temp = vector.Vector{k3Vel[i].X, k3Vel[i].Y, k3Vel[i].Z}
			temp.MultScalar(2)
			posChange.Add(temp)

			posChange.Add(k4Vel[i])
			posChange.MultScalar(1.0 / 6.0)

			body.Pos.Add(posChange)
		}

		// Handle collisions and escapes
		var escaping []*body.Body
		var colliding []map[*body.Body]bool

		addCollision := func(body1, body2 *body.Body) {
			added := false
			for _, groups := range colliding {
				if _, ok := groups[body1]; ok {
					groups[body2] = true
					added = true
				}
				if _, ok := groups[body2]; ok {
					groups[body1] = true
					added = true
				}
			}
			if !added {
				newmap := make(map[*body.Body]bool)
				newmap[body1] = true
				newmap[body2] = true
				colliding = append(colliding, newmap)
			}
		}

		for _, body := range w.bodies {
			// Check if body is 1) higher than escape velocity and 2) is more more
			// than 2X screens from center.
			if w.escaped(body) {
				escaping = append(escaping, body)
			} else {
				for _, body2 := range w.bodies {
					if body == body2 {
						continue
					}
					if body.Collides(body2) {
						addCollision(body, body2)
					}
				}
			}
		}

		// Handle escaping bodies
		for _, escaper := range escaping {
			fmt.Printf("%v: ESCAPED: %v\n", w.worldTime(), escaper)
			w.removeBody(escaper)
		}

		// Handle collisions
		for _, group := range colliding {
			var big *body.Body
			for b, _ := range group {
				if big == nil || b.Radius > big.Radius {
					big = b
				}
			}
			for small, _ := range group {
				if small != big {
					big.CollideWith(small)
					fmt.Printf("%v: COLLISION: %v\n", w.worldTime(), big)
					w.removeBody(small)
				}
			}
		}
	}
}

// New helper functions for RK4 implementation
// Calculate accelerations for all bodies and store in the provided slice
func (w *World) calculateAllAccelerations(accelerations []vector.Vector) {
	channels := make([]chan vector.Vector, len(w.bodies))

	for i, body := range w.bodies {
		channels[i] = make(chan vector.Vector)
		go w.calculateAcceleration(body, channels[i])
	}

	for i, ch := range channels {
		acc := <-ch
		if !math.IsNaN(acc.X) && !math.IsNaN(acc.Y) {
			accelerations[i] = acc
			w.bodies[i].Acc = acc  // Update body's acceleration
		} else {
			// Handle NaN case
			accelerations[i] = vector.Vector{0, 0, 0}
		}
	}
}

// Apply half step for midpoint calculations
func (w *World) applyHalfStep(initialPos, initialVel, velDelta, accDelta []vector.Vector) {
	for i, body := range w.bodies {
		// Position = initial + 0.5 * velocity * dt
		body.Pos.X = initialPos[i].X + 0.5 * velDelta[i].X
		body.Pos.Y = initialPos[i].Y + 0.5 * velDelta[i].Y
		body.Pos.Z = initialPos[i].Z + 0.5 * velDelta[i].Z

		// Velocity = initial + 0.5 * acceleration * dt
		body.Vel.X = initialVel[i].X + 0.5 * accDelta[i].X
		body.Vel.Y = initialVel[i].Y + 0.5 * accDelta[i].Y
		body.Vel.Z = initialVel[i].Z + 0.5 * accDelta[i].Z
	}
}

// Apply full step for final k4 calculation
func (w *World) applyFullStep(initialPos, initialVel, velDelta, accDelta []vector.Vector) {
	for i, body := range w.bodies {
		// Position = initial + velocity * dt
		body.Pos.X = initialPos[i].X + velDelta[i].X
		body.Pos.Y = initialPos[i].Y + velDelta[i].Y
		body.Pos.Z = initialPos[i].Z + velDelta[i].Z

		// Velocity = initial + acceleration * dt
		body.Vel.X = initialVel[i].X + accDelta[i].X
		body.Vel.Y = initialVel[i].Y + accDelta[i].Y
		body.Vel.Z = initialVel[i].Z + accDelta[i].Z
	}
}

// Reset all bodies to their initial state
func (w *World) resetToInitial(initialPos, initialVel []vector.Vector) {
	for i, body := range w.bodies {
		body.Pos.X = initialPos[i].X
		body.Pos.Y = initialPos[i].Y
		body.Pos.Z = initialPos[i].Z

		body.Vel.X = initialVel[i].X
		body.Vel.Y = initialVel[i].Y
		body.Vel.Z = initialVel[i].Z
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

func solarSystem(w, h int) *World {
	world := &World{
		scale:   1.0,
		mpp:     5.5e8,
		spt:     600,
		running: true,
		elapsed: 0,
		bodies:  make([]*body.Body, 6),
		width:   w,
		height:  h,
		mag:     1.0,
	}

	world.bodies[0] = body.NewBody("Sol", 0, 0, 696_340_000, 1.9885e30, 0.0, 0.0, sprites["sun"])
	world.bodies[0].Id = "Mother"
	world.bodies[1] = body.NewBody("Mercury", 46e9, 0, 2_439_700, 0.33011e24, 0.0, 58.98e3, sprites["mercury"])
	world.bodies[2] = body.NewBody("Venus", 0, 107.48e9, 6_051_800, 4.86750e24, -35.26e3, 0.0, sprites["venus"])
	world.bodies[3] = body.NewBody("Mars", 0, -206.62e9, 3_389_500, 0.64171e24, 26.50e3, 0.0, sprites["mars"])
	earth := body.NewBody("Earth", -147.09e9, 0, 6_371_000, 5.9724e24, 0.0, -30.29e3, sprites["earth"])
	world.bodies[4] = earth
	luna := body.NewBody("Luna", earth.Pos.X-0.3633e9, 0, 1_737_400, 0.07346e24, 0.0, earth.Vel.Y-1.082e3, sprites["luna"])
	world.bodies[5] = luna

	for _, body := range world.bodies {
		fmt.Printf("%v\n", body)
	}
	return world
}

func randomWithMoons(w, h, n, m int, df float64) *World {
	fmt.Printf("Making %v planets with %v moons each\n", n, m)
	world := &World{
		scale:   0.1,
		mpp:     5e5,
		spt:     1,
		running: true,
		elapsed: 0,
		bodies:  make([]*body.Body, n*m+n+1),
		width:   w,
		height:  h,
		mag:     1.0,
	}
	world.bodies[0] = body.NewBody("Mother", 0, 0, 30*world.mpp, 5e28, 0, 0, sprites["sun"])
	center := world.bodies[0]
	fmt.Printf("%v\n", center)
	maxDistance := math.Sqrt(float64(iPow(world.width, 2)+iPow(world.height, 2))) * 2.0
	bi := 1
	for i := 0; i < n; i++ {
		distance := 200.0 + math_rand.Float64()*maxDistance*df
		theta := math_rand.Float64() * math.Pi * 2
		pos := vector.New2DVector(-distance*math.Cos(theta)*world.mpp, -distance*math.Sin(theta)*world.mpp)
		circularOrbitVel := math.Sqrt(G * center.Mass / pos.Magnitude())
		u := pos.Unit()
		un := u.Normal2D()
		vel := vector.Vector{un.X, un.Y, un.Z}
		vel.MultScalar(circularOrbitVel)
		mass := math_rand.Float64() * 1e26
		radius := float64(8+math_rand.Intn(8)) * world.mpp
		world.bodies[bi] = body.NewBody(fmt.Sprintf("P%v", i), pos.X, pos.Y, radius, mass, vel.X, vel.Y, randomPlanetSprite())
		fmt.Printf("%v\n", world.bodies[bi])
		bi += 1
		for j := 0; j < m; j++ {
			//moon
			d := radius + float64(10+math_rand.Intn(40))*world.mpp
			// moon vel
			moonOrbVel := math.Sqrt(G * mass / d)
			var sign float64
			if math_rand.Intn(2) == 1 {
				sign = 1
			} else {
				sign = -1
			}
			mu := vector.Vector{0, sign, 0}
			mv := vector.Vector{vel.X, vel.Y, vel.Z}
			mu.MultScalar(moonOrbVel)
			mv.Add(mu)
			mm := 1e5 * math_rand.Float64()
			mr := float64(1+math_rand.Intn(4)) * world.mpp
			world.bodies[bi] = body.NewBody(fmt.Sprintf("P%vM%v", i, j), pos.X-sign*d, pos.Y, mr, mm, mv.X, mv.Y, randomPlanetSprite())
			fmt.Printf("%v\n", world.bodies[bi])
			bi += 1
		}
	}
	return world
}

func randomWorld(w, h, n int, pf float64, df float64) *World {
	world := &World{
		scale:   0.3,
		mpp:     5e5,
		spt:     1,
		running: true,
		elapsed: 0,
		bodies:  make([]*body.Body, n+1),
		width:   w,
		height:  h,
		mag:     1.0,
	}
	world.bodies[0] = body.NewBody("Mother", 0, 0, 30*world.mpp, 5e28, 0, 0, sprites["sun"])
	fmt.Printf("%v\n", world.bodies[0])
	center := world.bodies[0]
	maxDistance := math.Sqrt(float64(iPow(world.width, 2)+iPow(world.height, 2))) / 2.0
	maxDistance *= df
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
			baseMass*math_rand.Float64(), randomPlanetSprite())
		fmt.Printf("%v\n", world.bodies[i])
	}
	return world
}

func usage() string {
	return `Usage:
	nbody-go [-hPC -d<dimensions> -s=<spt> -p=<pf> -r=<df> -n=<numBodies> -m=<numMoons> -M=<mf>] MODE
Run N-Body simulation in mode MODE
Arguments:
  MODE        mode of the simulation, one of random, moons, solar
Options:
	-h --help
	-d=<dimensions>, --dimensions=<dimensions>  dimensions of screen in pixels [default: 1024x1024]
	-P        Start paused
	-C        Use plain white circle as planet graphic instead of random ones in moons and random MODE
	-s=<spt>  Seconds of world time to calculate per UI tick
	-p=<pf>   Perturbation factor for random world generation [default: 0.2]
	-r=<df>   Distance factor for random world generation [default: 1.0]
	-M=<mf>   For high DPI screens, scale up window by this amount [default: 1.0]
	-n=<numBodies>, --number=<numBodies>      Number of bodies to start [default: 60]
	-m=<numMoons>, --moons=<numMoons>         Number of moons per body [default: 3]
`
}

func run() {
	options, opterr := docopt.ParseDoc(usage())
	if opterr != nil {
		panic(opterr)
	}
	dims, _ := options.String("--dimensions")
	width, height := func() (int, int) {
		elems := strings.Split(dims, "x")
		w, _ := strconv.Atoi(elems[0])
		h, _ := strconv.Atoi(elems[1])
		return w, h
	}()
	numBodies, _ := options.Int("--number")
	numMoons, _ := options.Int("--moons")
	pf, _ := options.Float64("-p")
	df, _ := options.Float64("-r")
	mode, _ := options.String("MODE")
	spt, _ := options.Int("-s")
	paused, _ := options.Bool("-P")
	circleMode, _ = options.Bool("-C")
	mf, _ := options.Float64("-M")

	initRand()

	// initialize all the sprites
	loadSprite("sun", "./images/sun.png")
	loadSprite("earth", "./images/earth.png")
	loadSprite("jupiter", "./images/earth.png")
	loadSprite("luna", "./images/luna.png")
	loadSprite("mars", "./images/mars.png")
	loadSprite("venus", "./images/venus.png")
	loadSprite("mercury", "./images/mercury.png")
	loadSprite("circle", "./images/circle.png")
	planetPic, err := loadPicture("./images/planetsheet.png")
	if err != nil {
		panic(err)
	}
	for x := planetPic.Bounds().Min.X; x < planetPic.Bounds().Max.X; x += 128 {
		for y := planetPic.Bounds().Min.Y; y < planetPic.Bounds().Max.Y; y += 128 {
			name := fmt.Sprintf("planet%v", numPlanetSprites)
			sprites[name] = pixel.NewSprite(planetPic, pixel.R(x, y, x+128, y+128))
			numPlanetSprites += 1
		}
	}

	var world *World
	if mode == "random" {
		world = randomWorld(width, height, numBodies, pf, df)
	} else if mode == "solar" {
		world = solarSystem(width, height)
	} else if mode == "moons" {
		totalBodies := numBodies
		for (numBodies*numMoons + numBodies) > totalBodies {
			numBodies -= 1
		}
		world = randomWithMoons(width, height, numBodies, numMoons, df)
	} else {
		fmt.Printf("MODE %v is not valid\n", mode)
		fmt.Print(usage())
		os.Exit(2)
	}

	world.mag = mf

	if spt > 0 {
		world.spt = spt
	}
	if paused {
		world.running = false
	}

	cfg := pixelgl.WindowConfig{
		Title:  "N-Body Problem",
		Bounds: pixel.R(0, 0, float64(width)*world.mag, float64(height)*world.mag),
		VSync:  true,
	}

	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	// initialize font
	basicAtlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	infoTxt := text.New(pixel.V(win.Bounds().Max.X-200*world.mag, win.Bounds().Max.Y-20*world.mag), basicAtlas)

	followBody := -1
	center := vector.Vector{win.Bounds().Center().X, win.Bounds().Center().Y, 0}
	offset := center
	var closest *body.Body

	for !win.Closed() {

		// toggle running flag
		if win.JustPressed(pixelgl.KeySpace) {
			world.running = !world.running
		}

		// switch center from body to body
		if win.JustPressed(pixelgl.KeyN) {
			if followBody == -1 {
				followBody = 1
			} else {
				followBody += 1
				if followBody >= len(world.bodies) {
					followBody = 0
				}
			}
		}

		// Recenter
		if win.JustPressed(pixelgl.KeyC) {
			followBody = -1
			offset = center
		}

		// Turn off closest vec, accel and info display
		if win.JustPressed(pixelgl.MouseButtonRight) {
			closest = nil
		}

		// Get details on a body
		if win.JustPressed(pixelgl.MouseButtonLeft) {
			closest = nil
			mouseCoords := win.MousePosition()
			mouseCoordsVec := vector.Vector{mouseCoords.X, mouseCoords.Y, 0}
			var closestDistance float64 = 0.0
			for _, body := range world.bodies {
				bodyScreen := world.worldToScreen(&body.Pos)
				bodyScreen.Add(offset)
				if closest == nil {
					closest = body
					closestDistance = bodyScreen.DistanceTo(mouseCoordsVec)
				} else {
					bodyDistance := bodyScreen.DistanceTo(mouseCoordsVec)
					if bodyDistance < closestDistance {
						closestDistance = bodyDistance
						closest = body
					}
				}
			}
		}

		// Speed sim up or down
		if win.Pressed(pixelgl.KeyI) {
			world.spt += 1
		}
		if win.Pressed(pixelgl.KeyK) {
			world.spt -= 1
			if world.spt == 0 {
				world.spt = 1
			}
		}
		// zoom in/out
		world.scale *= math.Pow(1.2, win.MouseScroll().Y)
		win.Clear(colornames.Black)
		mat := pixel.IM

		if followBody >= 0 && followBody < len(world.bodies) {
			offset = vector.Vector{center.X, center.Y, center.Z}
			offset.Sub(world.worldToScreen(&world.bodies[followBody].Pos))
		}
		if len(world.bodies) <= 0 {
			fmt.Println("There are no more bodies, ending sim...")
			os.Exit(3)
		}
		for _, body := range world.bodies {
			sprite := body.Sprite
			if sprite == nil {
				panic(fmt.Sprintf("NO SPRITE FOR BODY %v", body))
			}
			spriteSize := float64(sprite.Frame().Max.X)
			brp := body.Radius * world.scale / world.mpp
			if brp < MinRadius {
				brp = MinRadius
			}
			sf := brp / spriteSize
			bodyMat := mat.ScaledXY(pixel.ZV, pixel.V(sf*world.mag, sf*world.mag))
			screenPos := world.worldToScreen(&body.Pos)
			screenPos.Add(offset)
			bodyMat = bodyMat.Moved(pixel.V(screenPos.X, screenPos.Y))
			sprite.Draw(win, bodyMat)
		}
		// Update info text
		infoTxt.Clear()
		fmt.Fprintf(infoTxt, "N: %v\n", len(world.bodies))
		fmt.Fprintf(infoTxt, "t: %v\n", world.worldTime())
		fmt.Fprintf(infoTxt, "S: %4.2f\n", world.scale)
		fmt.Fprintf(infoTxt, "dt: %v\n", world.spt)
		// Add on clicked body info
		if closest != nil {
			// Add Vel and Acc vectors
			closestPos := world.worldToScreen(&closest.Pos)
			closestPos.Add(offset)
			imd := imdraw.New(nil)
			imd.Color = colornames.Red
			imd.EndShape = imdraw.SharpEndShape
			// velocity
			vel := vector.MultScalar(closest.Vel.Unit(), 40)
			endVel := vector.Add(closestPos, vel)
			imd.Push(pixel.V(closestPos.X, closestPos.Y), pixel.V(endVel.X, endVel.Y))
			imd.Line(2)
			imd.Draw(win)
			// acceleration
			imd.Color = colornames.Green
			acc := vector.MultScalar(closest.Acc.Unit(), 40)
			endAcc := vector.Add(closestPos, acc)
			imd.Push(pixel.V(closestPos.X, closestPos.Y), pixel.V(endAcc.X, endAcc.Y))
			imd.Line(2)
			imd.Draw(win)

			fmt.Fprintf(infoTxt, "\n%v:\n", closest.Name)
			fmt.Fprintf(infoTxt, "P: (%5.2e,%5.2e)\n", closest.Pos.X, closest.Pos.Y)
			fmt.Fprintf(infoTxt, "V: (%5.2e,%5.2e)\n", closest.Vel.X, closest.Vel.Y)
			fmt.Fprintf(infoTxt, "A: (%5.2e,%5.2e)\n", closest.Acc.X, closest.Acc.Y)
		}
		infoTxt.Draw(win, pixel.IM.Scaled(infoTxt.Orig, world.mag))
		win.Update()
		world.tick()
	}
}

func main() {
	pixelgl.Run(run)
}

func testMain() {
	world := randomWorld(1024, 1024, 60, 0.5, 1.0)
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
