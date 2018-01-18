FROM golang:1.8

WORKDIR /go/src/app

# copy the code into the container so it is available at build-time for go-wrapper.
# the docker-compose file mounts the code directory at the same path as well for code-reloading
COPY . .

RUN go-wrapper download   # "go get -d -v ./..."
RUN go-wrapper install    # "go install -v ./..."
RUN go get github.com/db-journey/journey # Grab journey for migrations: This might take awhile

# Use fresh to run the app and auto-recompile on change
# CMD is the command that is run at RUN-TIME when a new container is started from the image
CMD  go get github.com/pilu/fresh && \
  fresh
