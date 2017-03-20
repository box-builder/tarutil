all: test

install_box:
	@sh install_box.sh

install_box_ci:
	@sh install_box_ci.sh

build: 
	PATH=${PATH}:${PWD}/bin box -t unclejack/tarutil build.rb	

run_test:
	docker run unclejack/tarutil

test: install_box build run_test

test-ci: install_box_ci build run_test

.PHONY: build install_box
