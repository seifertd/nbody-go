# nbody-go
Implementation of N-Body problem in Golang, usage below.

## Examples

1. Simulate large central body and 50 bodies in circular orbits:
```bash
$ ./nbody random -n 50 -p 0.0 -r 1.0 -d 768x768
```

The initial circular orbits can be perturbed from perfectly circular by 
specifying a non zero -p flag. The bodies can be clustered closer to the central
body by specifying a -r < 1.0. -r values greater than 1 just spread the distribution
of the bodies out more.

2. Simulate large central body with bodies with moonlets all in circular orbits. Use -n 
   to control the number of bodies and -m to dictate how many moonlets per body.
```bash
$ ./nbody moons -n 10 -m 2 -r 1.2 -d 768x768
```

The -r flag can be used to stretch out the distance from center of the bodies.

3. Simulate the inner solar system
```bash
$ ./nbody solar
```

This mode does not take any other flags.

    Usage: nbody [-hP -d<dimensions> -s=<spt> -p=<pf> -r=<df> -n=<numBodies> -m=<numMoons] MODE
    Run N-Body simulation in mode MODE
    Arguments:
      MODE        mode of the simulation, one of random, moons, solar
    Options:
      -h --help
      -d=<dimensions>, --dimensions=<dimensions>  dimensions of screen in pixels [default: 1024x1024]
      -P        Start paused
      -s=<spt>  Seconds of world time to calculate per UI tick
      -p=<pf>   Perturbation factor for random world generation [default: 0.2]
      -r=<df>   Distance factor for random world generation [default: 1.0]
      -n=<numBodies>, --number=<numBodies>  Number of bodies to start [default: 60]
      -m=<numMoons>, --moons=<numMoons>     Number of moons per body [default: 3]
