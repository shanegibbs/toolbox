export SHAM_LOG=trace

UNAME_S:=$(shell uname -s)
ifeq ($(UNAME_S),Linux)
	OS = linux
endif
ifeq ($(UNAME_S),Darwin)
	OS = darwin
endif

build: clean-containers clean-images
	printenv |grep SHAM

	docker build . -t sham
	# docker run --rm sham ls -al /sham
	docker rm -f toolbox 2>/dev/null || true
	
	rm -rf output && mkdir output
	docker run --rm sham sh -c "cd /sham && tar -cf- sham.$(OS)" |tar -xvf- -C output
	mv output/sham.$(OS) output/sham
	
	env PATH="$(PWD)/output:$(PATH)" sham bash -c "cat /etc/lsb-release; id"

run-sham:
	go run cmd/sham/main.go bash -c "cat /etc/lsb-release; id"

docker-build:
	docker run --rm -v $(PWD):/work --workdir /work  -v $(PWD)/.cache:/root/.cache golang:1.14 go build ./...

test: build
	toolbox terraform version
	terraform version
	echo foo |toolbox cat
	toolbox ssh -T git@github.com || true
	toolbox

install:
	# mkdir -p ~/.sham/bin
	# go build -o ~/.sham/bin/sham cmd/stub/main.go
	go build -o ~/bin/sham cmd/sham/main.go
	# docker run --rm --entrypoint cat toolbox /toolbox-stub > ~/.toolbox/bin/toolbox
	# ln -s ../bin/toolbox ~/.toolbox/stubs/toolbox
	# ln -s toolbox ~/.toolbox/stubs/terraform
	# ln -s toolbox ~/.toolbox/stubs/deployer

nginx:
	docker run --name sham-nginx --rm --net=host -v $(PWD)/build-context:/usr/share/nginx/html nginx

test-build-context:
	docker build --no-cache \
		--build-arg IMAGE='ubuntu' \
		--build-arg SHAM_INIT_OPTIONS='{"Username":"shane.gibbs","Home":"/Users/shane.gibbs","Uid":1084496081,"Gid":1538143563}' \
		--build-arg USER_ID=123 \
		-f build-context/Dockerfile build-context

.PHONY: toolbox
toolbox:
	docker build --pull -t toolbox toolbox

clean-containers:
	docker rm -f $(shell docker ps -qf label=com.gibbsdevops.sham) || true

clean-images:
	docker image prune --all --filter label=com.gibbsdevops.sham --force

clean: clean-containers clean-images
	rm -rf output
