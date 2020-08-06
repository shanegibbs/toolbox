build:
	go build ./...
	docker build . -t sham
	docker rm -f toolbox; go run cmd/sham/main.go || docker logs toolbox

	# ln -s ../bin/toolbox ~/.toolbox/stubs/toolbox
	# ln -s toolbox ~/.toolbox/stubs/terraform
	# ln -s toolbox ~/.toolbox/stubs/deployer

test: build
	toolbox terraform version
	terraform version
	echo foo |toolbox cat
	toolbox ssh -T git@github.com || true
	toolbox

install:
	# mkdir -p ~/.sham/bin
	# go build -o ~/.sham/bin/sham cmd/stub/main.go
	go build -o ~/bin/sham cmd/stub/main.go
	# docker run --rm --entrypoint cat toolbox /toolbox-stub > ~/.toolbox/bin/toolbox

nginx:
	docker run --name sham-nginx --rm --net=host -v $(PWD)/build-context:/usr/share/nginx/html:ro nginx

test-build-context:
	docker build --no-cache \
		--build-arg IMAGE='ubuntu' \
		--build-arg SHAM_INIT_OPTIONS='{"Username":"shane.gibbs","Home":"/Users/shane.gibbs","Uid":1084496081,"Gid":1538143563}' \
		--build-arg USER_ID=123 \
		-f build-context/Dockerfile build-context
