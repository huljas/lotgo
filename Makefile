all: vet build-example test

vet:
	go vet *.go
	go vet ws/*.go

build-example:
	cd example && go build
	cd ws && go build

test:
	go test

glide:
	glide up