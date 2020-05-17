# gobuild - Multi container builds from the shell

## tl;dr

Execute builds with multiple jobs performed in different containers from the shell.

```
gobuild [config=.gobuild.yaml]
```

## Install

```
go get github.com/Useurmind/gobuild
```

## .gobuild.yaml

The `.gobuild.yaml` represents the configuration for your multi container build. When executed `gobuild` will look for it in the current working directory.

Example:
```yaml
# define some environment variables that 
# apply to all jobs
env:
  # use simple values
  VAR1: value 1
  # or use variables from the environment you run gobuild in
  # expand is done according to https://golang.org/pkg/os/#Expand
  VAR2: value 2 $VAR4

# you define a set of jobs that will be executed
# each in its own container
jobs:
  # each job has 
  # - a name for identifying it in the logs
  # - an image that is run via docker
  # - a set of scripts that are executed in the container
  # - a set of environment variables that only apply to the job
  - name: job1
    image: ubuntu
    scripts: 
      - echo "hallo job"
    env:
      # simple value
      VAR3: value 3
      # expand from os environment
      VAR4: $VAR4
      # expand from global env in yaml
      VAR5: $VAR1

  - name: job2
    image: ubuntu
    scripts: 
      - echo "hallo job 2"
```

All jobs are performed sequentially. Its like calling `docker run` multiple times. The logs of all containers will be outputed to the shell.
