# toolbox

![Build](https://github.com/shanegibbs/toolbox/workflows/Build/badge.svg)

* <https://github.com/shanegibbs/toolbox/packages>
* <https://hub.docker.com/r/shanegibbs/toolbox>

1. rename to "sham"
1. Running "pre-configured" docker commands
1. Run adhoc container, or single run
1. Bind options. What directories to bind? Add more?
1. Different HOME from host?
1. ~~should we build new image from base? or add into existing at runtime?~~
1. `sham run ubuntu` to run with defaults and no config
1. replicate log level to in-container
1. save state of running containers so we can cleanup shams
1. bin is 8MB. rust? docker client is golang though
1. add all groups the user belongs to?
1. log not loading correctly in container
1. pass go deps into build container, remove local go build
1. add app version info
1. pull sham image from github, optional locally for dev
1. replace docker exec
1. only pass in workdir by default. add to labels, check when running sham

## stand-alone tool

## yaml config

Search order:

1. current directory
1. previous directories
1. user home
1. in image

```yaml
name: toolbox
image: toolbox
keep: true
shams:
- terraform
- deployer
```

* <https://superuser.com/questions/521657/zsh-automatically-set-environment-variables-for-a-directory>

## caller use cases

1. `sham run ubuntu` - single use by default?
1. `sham run ubuntu ls` - single use by default?
1. `sham` - maybe? tricky with multiple sham containers
1. `ubuntu`
1. `terraform`
1. `sham start ubuntu`
1. `sham stop ubuntu`
1. `sham update ubuntu`
1. `sham clean` - remove sham containers and images

## issues

```shell
bash-5.0$ ls -al |grep host
drwxr-xr-x   4 root root    4096 Aug  7 09:09 host%!(EXTRA string=
```

## problem statement

Example:

* <https://golangci-lint.run/usage/install/#local-installation>

```shell
docker run --rm -v $(pwd):/app -w /app golangci/golangci-lint:v1.30.0 golangci-lint run -v
```

Could be:
* `sham run golangci/golangci-lint run -v`
* `golangci-lint run -v` with `sham install golangci/golangci-lint`
