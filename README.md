# toolbox

![Build](https://github.com/shanegibbs/toolbox/workflows/Build/badge.svg)

* <https://github.com/shanegibbs/toolbox/packages>
* <https://hub.docker.com/r/shanegibbs/toolbox>

1. rename to "sham"
1. Running "pre-configured" docker commands
1. Run adhoc container, or single run
1. Bind options. What directories to bind? Add more?
1. Different HOME from host?
1. should we build new image from base? or add into existing at runtime?

## stand-alone tool

## yaml config

Search order:

1. current directory
1. user home
1. in image

```yaml
- name: ubuntu
  image: ubuntu
  keep: true
```
