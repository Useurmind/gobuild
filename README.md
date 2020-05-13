# gobuild - Multi container builds from the shell

![Go](https://github.com/Useurmind/gobuild/workflows/Go/badge.svg?branch=master)

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
# you define a set of jobs that will be executed
# each in its own container
jobs:
  # each job has 
  # - a name for identifying it in the logs
  # - an image that is run via docker
  # - a set of scripts that are executed in the container
  - name: job1
    image: ubuntu
    scripts: 
      - echo "hallo job"

  - name: job2
    image: ubuntu
    scripts: 
      - echo "hallo job 2"
```

All jobs are performed sequentially. Its like calling `docker run` multiple times. The logs of all containers will be outputed to the shell.
