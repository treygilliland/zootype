.PHONY: build run clean install uninstall

build:
	@./build.sh build

run: build
	@ZOOTYPE_CONFIG=zootype.toml ZOOTYPE_BIN_DIR=bin ./zootype.sh

install:
	@./build.sh install

uninstall:
	@./build.sh uninstall

clean:
	rm -rf bin


