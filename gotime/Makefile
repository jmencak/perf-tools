PROGRAM=gotime

all: $(PROGRAM)

$(PROGRAM): $(PROGRAM).go
	go build $<

static: $(PROGRAM).go
	CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags '-s' $<

run: $(PROGRAM).go
	go run $<

fmt format: $(PROGRAM).go
	go fmt $<

strip:
	strip $(PROGRAM)

clean:
	go clean
