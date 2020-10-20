VERSION ?= $(shell git describe --always --abbrev=16 --dirty)

build:
	VERSION=$(VERSION) docker-compose build

up:
	docker-compose up

down:
	docker-compose down

push-docker-images:
	docker push samolds/shipyard-api
	docker push samolds/shipyard-web

clean-docker: stop down
	-docker container rm -f shipyard_api shipyard_db shipyard_web shipyard_grafana shipyard_prometheus shipyard_nginx_proxy shipyard_nginx_proxy_letsencrypt
	-docker image rm -f samolds/shipyard-api samolds/shipyard-web shipyard_nginx_proxy
	-docker container prune -f
	-docker image prune -f
	-docker volume prune -f
	-docker network prune -f
	-rm -rf monitor/grafana monitor/grafana_data

start:
	docker-compose start

stop:
	docker-compose stop

shell-server: start
	docker exec -ti shipyard_api /bin/sh

shell-db: start
	docker exec -ti shipyard_db /bin/sh

clean:
	cd ../go/src/shipyard && make clean
	cd ../web/shipyard && make clean

devup:
	cd ../go/src/shipyard && (make clean runsqlite3 &)
	cd ../web/shipyard && make clean run

fresh:
	make clean
	make clean-docker
	make build
	make up
