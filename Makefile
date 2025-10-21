.PHONY: build run test clean install uninstall

build:
	@./zootype.sh build

run: build
	@ZOOTYPE_BIN_DIR=bin ./zootype.sh

test: build
	@echo "Running smoke tests..."
	@./bin/gophertype --version | grep -q "gophertype" && echo "✓ gophertype"
	@./bin/pythontype --version | grep -q "pythontype" && echo "✓ pythontype"
	@./bin/cameltype --version | grep -q "cameltype" && echo "✓ cameltype"
	@./bin/rattype --version | grep -q "rattype" && echo "✓ rattype"
	@./bin/crabtype --version | grep -q "crabtype" && echo "✓ crabtype"
	@./bin/dinotype --version | grep -q "dinotype" && echo "✓ dinotype"
	@./bin/eggtype --version | grep -q "eggtype" && echo "✓ eggtype"
	@echo "All tests passed!"

versions: build
	@./bin/gophertype --version
	@./bin/pythontype --version
	@./bin/cameltype --version
	@./bin/rattype --version
	@./bin/crabtype --version
	@./bin/dinotype --version
	@./bin/eggtype --version

install:
	@./zootype.sh install

uninstall:
	@./zootype.sh uninstall

clean:
	rm -rf bin
	rm -rf cameltype/_build
	cd rattype && make clean 2>/dev/null || true
	cd crabtype && cargo clean 2>/dev/null || true
	rm -f dinotype/dinotype


