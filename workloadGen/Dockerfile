FROM golang:1.8

WORKDIR /go/src/app

# copy the code into the container so it is available at build-time for go-wrapper.
# the docker-compose file mounts the code directory at the same path as well for code-reloading
COPY . .

RUN go-wrapper download   # "go get -d -v ./..."
RUN go-wrapper install    # "go install -v ./..."
RUN go get github.com/pilu/fresh

RUN ["apt-get", "update"]
RUN ["apt-get", "install", "-y", "vim"]

# Use fresh to run the app and auto-recompile on change
# CMD is the command that is run at RUN-TIME when a new container is started from the image.
# This also prevents the container from exiting after the go program exits, since fresh is still running
CMD fresh
