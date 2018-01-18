# do this before running
build:
	docker-compose build

up:
	docker-compose up

reload:
	docker-compose up --build

reloadweb:
	docker-compose up --no-deps --build web

upweb:
	docker-compose up --no-deps web

reloadtx:
	docker-compose up --no-deps --build tx

uptx:
	docker-compose up --no-deps tx

upcassandra:
	docker-compose up cassandra
