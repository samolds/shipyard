VERSION ?= $(shell git describe --always --abbrev=16 --dirty)

build:
	VERSION=$(VERSION) docker-compose build

up:
	docker-compose up

down:
	docker-compose down

push-docker-images:
	docker push samolds/democart-api
	docker push samolds/democart-web

clean-docker: stop down
	-docker container rm -f democart_api democart_db democart_web democart_grafana democart_prometheus democart_nginx_proxy democart_nginx_proxy_letsencrypt
	-docker image rm -f samolds/democart-api samolds/democart-web democart_nginx_proxy
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
	docker exec -ti democart_api /bin/sh

shell-db: start
	docker exec -ti democart_db /bin/sh

clean:
	cd go/src/democart && make clean
	cd web/democart && make clean

devup:
	cd go/src/democart && (make clean runsqlite3 &)
	cd web/democart && make clean run

fresh:
	make clean
	make clean-docker
	make build
	make up
