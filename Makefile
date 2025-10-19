.PHONY: build run clean install uninstall

build:
	@./zootype.sh build

run: build
	@ZOOTYPE_BIN_DIR=bin ./zootype.sh

install:
	@./zootype.sh install

uninstall:
	@./zootype.sh uninstall

clean:
	rm -rf bin


