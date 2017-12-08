# Chaos Monkey

Uses the docker api to randomly stop (or remove) running containers.

## Running

```
docker run -v /var/run/docker.sock:/var/run/docker.sock bluebamboostudios/chaos-monkey [options]
```

## Options

*-version*

Show version and exit

*-debug*

Enable debug logging

*-dry-run*

Don't actually stop or remove containers, simply log expected operations

*-remove*

Also remove stopped containers

*-interval*

Time between chaos

*-volumes*

Also remove attached volumes

*-max-procs*

Set number of CPUs to use simultaneously, defaults to number of available cores.

## Environment Variables

*SKIP_IMAGE*

List of images to not stop, defaults to self. ```bluebamboostudios/chaos-monkey```

*STOP_PROBABILITY*

Int to use for stop probability, defaults to 1000. 