IMAGE_NAME=`cat IMAGE_NAME`
CONTAINER_NAME=`cat CONTAINER_NAME`

compile:
	docker build -t $(IMAGE_NAME) .


run: compile
	docker run -it --rm --name $(CONTAINER_NAME) $(IMAGE_NAME)
