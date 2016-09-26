PROGRAM=slstress
TAG=jmencak/$(PROGRAM)

all: $(PROGRAM)

$(PROGRAM): $(PROGRAM).go
	go build $<

static: $(PROGRAM).go
	CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags '-s' $<

run: $(PROGRAM).go
	go run $<

fmt format: $(PROGRAM).go
	go fmt $<

docker: $(PROGRAM) Dockerfile
	(docker build -t $(TAG) .)

push: 
	(docker push $(TAG))

strip:
	strip $(PROGRAM)

clean:
	go clean
