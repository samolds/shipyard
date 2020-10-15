VERSION ?= $(shell git describe --always --abbrev=16 --dirty)

# TODO(sam): dbx is no longer stable. go get .../dbx.v1 fails due to unfound
# dependencies, so the docker image wont be able to use "make setup" and have a
# dbx binary to generate fresh. so the generate calls can only run during dev.

server-no-dbx:
	go build -ldflags "-X main.version=$(VERSION)" -o ./democart .

server: generate
	make server-no-dbx

setup-no-dbx:
	go mod download

setup:
	go get gopkg.in/spacemonkeygo/dbx.v1 && go mod download

generate: database/schema.dbx.go
	go generate ./database

test-no-dbx:
	go test -count=1 ./...

test: generate
	make test-no-dbx

runsqlite3: generate test server
	./democart --loglevel debug --configs config.hcl \
		--db_url "sqlite3:democart.sqlite3.db?sslmode=disable&cache=shared" \
		--public_api_url "http://localhost:8080" \
		--public_idp_url "http://localhost:8081" \
		--client_hosts "http://localhost:3000"

clean:
	rm -f ./democart democart.sqlite3.db*
	go clean

docker-build:
	docker build . -t samolds/democart-api:$(VERSION)

version:
	@echo "$(VERSION)"
