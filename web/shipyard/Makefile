VERSION ?= $(shell git describe --always --abbrev=16 --dirty)

install:
	npm ci

build:
	npm run build

run:
	npm start

clean:
	rm -rf build

docker-build:
	docker build . -t samolds/shipyard-web:$(VERSION)

fresh:
	make clean
	rm -rf node_modules
	make install
	make build
