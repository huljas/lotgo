all: vet build-example test

vet:
	go vet *.go
	go vet http/*.go

build-example:
	cd example && go build
	cd http && go build

test:
	go test

glide:
	glide up