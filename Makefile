all: vet build-example test

vet:
	go vet *.go

build-example:
	cd example && go build

test:
	go test