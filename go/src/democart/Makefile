all: server

#VERSION = $(shell git rev-parse --verify HEAD)
VERSION = "alpha"

server: generate
	go build -ldflags "-X main.version=$(VERSION)" -o ./democart .

setup:
	go get gopkg.in/spacemonkeygo/dbx.v1 && go mod download

generate: database/schema.dbx.go
	go generate ./database

test: generate
	go test -count=1 ./...

run: generate test server
	./democart --loglevel debug --config config.hcl

runsqlite3: generate test server
	./democart --loglevel debug --config config.hcl \
		--db_url "sqlite3:democart.sqlite3.db?sslmode=disable"

clean:
	rm -f ./democart democart.sqlite3.db*
	go clean

docker:
	docker run --rm \
		-v "$$PWD":/go/src/democart \
		-w /go/src/democart \
		democart:$(VERSION) \
		make server

docker-build:
	docker build . -t democart_server:$(VERSION)
