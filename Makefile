build:
	docker rm -f toolbox || true
	rm -rf ~/.toolbox
	mkdir -p ~/.toolbox/{bin,stubs}

	docker build -t toolbox .
	env GOOS=darwin GOARCH=amd64 go build -o ~/.toolbox/bin/toolbox stub/main.go
	ln -s ../bin/toolbox ~/.toolbox/stubs/toolbox
	ln -s toolbox ~/.toolbox/stubs/terraform

test: build
	toolbox terraform version
	terraform version
	echo foo |toolbox cat
	toolbox ssh -T git@github.com || true
	toolbox

install:
	mkdir -p ~/.toolbox/bin
	docker run --rm --entrypoint cat toolbox /toolbox-stub > ~/.toolbox/bin/toolbox
