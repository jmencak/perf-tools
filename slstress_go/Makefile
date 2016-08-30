PROGRAM=slstress

all: $(PROGRAM)

$(PROGRAM): $(PROGRAM).go
	go build $<

run: $(PROGRAM).go
	go run $<

fmt format: $(PROGRAM).go
	go fmt $<

strip:
	strip $(PROGRAM)

clean:
	go clean
