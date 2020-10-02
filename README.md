# nbody-go
Implementation of N-Body problem in Golang, usage below.

## Building & Running
1.  git clone https://github.com/seifertd/nbody-go.git
1.  [Install pre-requisites for faiface/pixel go 2D graphics library](https://github.com/faiface/pixel#requirements) -- I've run this on Mac OS X and Linux, tried to run it on a Raspberry Pi 4, but the library support is lacking.
1.  go build
1.  ./nbody-go -h

## Examples

1. Simulate large central body and 50 bodies in circular orbits:
```bash
$ ./nbody-go random -n 50 -p 0.0 -r 1.0 -d 768x768
```

The initial circular orbits can be perturbed from perfectly circular by 
specifying a non zero -p flag. The bodies can be clustered closer to the central
body by specifying a -r < 1.0. -r values greater than 1 just spread the distribution
of the bodies out more.

2. Simulate large central body with planetoids with moonlets all in circular orbits. Use -n 
   to control the total number of bodies and -m to dictate how many moonlets per body.
   The sim will create enough planetoids to ensure the total number of planetoids and
   moonlets created is less than the value of the -n flag.

   This will create 5 planetoids, each with 2 moonlets for 15 bodies total:
```bash
$ ./nbody-go moons -n 15 -m 2 -r 1.2 -d 768x768
```

The -r flag can be used to stretch out the distance from center of the bodies.

3. Simulate the inner solar system
```bash
$ ./nbody-go solar
```

This mode does not take any other flags.

### High DPI Screens

On Linux Mint running on an old Mac Book Pro with a retina display, I found the GUI text was so small as to be hard to read. Provide `-M 2.0` or such to magnify the window by that much and make the text easier to read.

## While Sim is Running

A info display of total number of bodies in the simulation, elapsed world time, zoom and seconds per
tick is shown in the upper right of the window. As bodies collide, the sim attempts to preserve momentum.
The body in a colliding group with the largest radius is kept and absorbs the mass of the other bodies
in the group, increasing radius to keep original density the same (dubious). The remaining body's momentum
is set equal to the group's momentum at time of the collision and a message will be printed to the console
giving details on the resulting body's parameters. If a body gets far enough away from the center and has
reached escape velocity, it will be removed from the sim and a message so indicating is printed to the console.

### Controls

* Press Space to pause and unpause the simulation
* Press the `I` key to speed up the simulation (increases seconds of world time per UI tick)
* Press the `K` key to slow the simulation down (decreases seconds of world time per UI tick)
* Press the `N` key repeatedly to cycle through the bodies and center them on the screen
* Press the `C` key to re-center the display
* Use mouse scroll wheel or 2-finger drag to zoom in and out.
* Press the left mouse button to select a body and show the following:
  * The body's name, velocity and acceleration in the info display
  * A green velocity direction vector.
  * A red acceleration direction vector
* Press the right mouse button to turn off the closest body display

## Usage

```
> nbody-go [-hPC -d<dimensions> -s=<spt> -p=<pf> -r=<df> -n=<numBodies> -m=<numMoons> -M=<mf>] MODE
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
```
