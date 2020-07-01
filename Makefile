new:
	make clean-docker
	make build
	make up

clean:
	cd go/src/democart && make clean
	cd web/democart && make clean

build:
	docker-compose -f docker-compose.yml build

up:
	docker-compose up

start:
	docker-compose start

stop:
	docker-compose stop

up-bg:
	docker-compose up -d

shell-server: start
	docker exec -ti democart_server /bin/sh

shell-db: start
	docker exec -ti democart_db /bin/sh

clean-docker: stop
	rm -rf go/src/democart/postgres/data
	-docker container rm -f democart_server democart_web democart_db
	-docker image rm -f democart_server democart_web democart_db
	-docker container prune -f
	-docker image prune -f

devup:
	cd go/src/democart && (make clean runsqlite3 &)
	cd web/democart && make clean run
