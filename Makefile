all: test

install_box:
	@sh install_box.sh

build: install_box
	box -t unclejack/tarutil build.rb	

test: build
	docker run -it unclejack/tarutil

.PHONY: build install_box
