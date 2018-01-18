FROM golang:1.8

WORKDIR /go/src/app
COPY . .

RUN go-wrapper download   # "go get -d -v ./..."
RUN go-wrapper install    # "go install -v ./..."

# Use fresh to run the app and auto-recompile on change
CMD go get github.com/pilu/fresh && \ 
  fresh
# CMD ["go-wrapper", "run"] # ["app"]
