# nbody-go
Implementation of N-Body problem in Golang

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
