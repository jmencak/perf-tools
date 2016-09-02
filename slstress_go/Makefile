PROGRAM=slstress
BIN_DIR=bin

all: $(PROGRAM)

$(PROGRAM): $(PROGRAM).go
	CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags '-s' $<

run: $(PROGRAM).go
	go run $<

fmt format: $(PROGRAM).go
	go fmt $<

docker: $(PROGRAM) Dockerfile
	(docker build -t $(PROGRAM) .)

strip:
	strip $(PROGRAM)

clean:
	go clean
