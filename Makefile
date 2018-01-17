# do this before running
build:
	docker-compose build

reloadweb:
	docker-compose up --no-deps --build web

upweb:
	docker-compose up --no-deps web

upcassandra:
	docker-compose up cassandra

up:
	docker-compose up
