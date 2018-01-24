# do this before running
build:
	docker-compose build

up:
	docker-compose up

reload:
	docker-compose up --build

upweb:
	docker-compose up --no-deps web

reloadweb:
	docker-compose up --no-deps --build web

upjob:
	docker-compose up --no-deps job

reloadjob:
	docker-compose up --no-deps --build job

updb:
	docker-compose up db

upquote:
	docker-compose up --no-deps quote
