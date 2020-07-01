all: web

#VERSION = $(shell git rev-parse --verify HEAD)
VERSION = "alpha"

web:
	npm run build

run:
	npm start

clean:
	rm -rf build

docker-build:
	docker build . -t democart_web:$(VERSION)
