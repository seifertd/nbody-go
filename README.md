# nbody-go
Implementation of N-Body problem in Golang, usage below.

## Building & Running
1. git clone https://github.com/seifertd/nbody-go.git
2. go build
3. ./nbody-go random

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

## While Sim is Running

A info display of total number of bodies in the simulation, elapsed world time, zoom and seconds per
tick is shown in the upper right of the window.

* Press Space to pause and unpause the simulation
* Press the `I` key to speed up the simulation (increases seconds of world time per UI tick)
* Press the `K` key to slow the simulation down (decreases seconds of world time per UI tick)
* Press the `N` key repeatedly to cycle through the bodies and center them on the screen
* Press the `C` key to re-center the display
* Use scroll wheel or 2-finger swipe to zoom in and out.
* Press the left mouse button to select a body and show the following:
  * The body's name, velocity and acceleration in the info display
  * A green velocity direction vector.
  * A red acceleration direction vector
* Press the right mouse button to turn off the closest body display

## Usage

	   nbody-go [-hPC -d<dimensions> -s=<spt> -p=<pf> -r=<df> -M=<magFact> -n=<numBodies> -m=<numMoons] MODE
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
	 -M=<magFact> For high DPI screens, scale up window by this amount [default: 1.0]
         -n=<numBodies>, --number=<numBodies>  Number of bodies to start [default: 60]
         -m=<numMoons>, --moons=<numMoons>     Number of moons per body [default: 3]
