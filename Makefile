build:
	docker build . -t sham
	docker rm -f toolbox; go run cmd/stub/main.go || docker logs toolbox

build-old:
	docker rm -f toolbox || true
	rm -rf ~/.toolbox
	mkdir -p ~/.toolbox/{bin,stubs}

	docker build -t toolbox .
	env GOOS=darwin GOARCH=amd64 go build -o ~/.toolbox/bin/toolbox stub/main.go
	ln -s ../bin/toolbox ~/.toolbox/stubs/toolbox
	ln -s toolbox ~/.toolbox/stubs/terraform
	ln -s toolbox ~/.toolbox/stubs/deployer

test: build
	toolbox terraform version
	terraform version
	echo foo |toolbox cat
	toolbox ssh -T git@github.com || true
	toolbox

install:
	mkdir -p ~/.toolbox/bin
	docker run --rm --entrypoint cat toolbox /toolbox-stub > ~/.toolbox/bin/toolbox

nginx:
	docker run --name sham-nginx --rm --net=host -v $(PWD)/build-context:/usr/share/nginx/html:ro nginx

test-build-context:
	docker build --no-cache \
		--build-arg IMAGE='ubuntu' \
		--build-arg SHAM_INIT_OPTIONS='{"Username":"shane.gibbs","Home":"/Users/shane.gibbs","Uid":1084496081,"Gid":1538143563}' \
		--build-arg USER_ID=123 \
		-f build-context/Dockerfile build-context
