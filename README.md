## Mini-S3

A mini version of the S3 datastore built with Go.

## Build for production

Modify the paths for the levelDB path and then execute:

	GOOS=linux GOARCH=amd64 ./build.sh

## Build for development

Modify the paths for the levelDB path and then execute:

	./build.sh

### Build profile support

To build profile support in uncomment these lines in `main.go`

	uncomment for profile support

After running quit the program then run

	go tool pprof main /tmp/profile014884158/cpu.pprof

where /tmp/profiel014884158 is the name of the profile built when starting
`main`.

## Setup

	cp config.sample config

## API support

Place a file or information into an object:

	curl -X POST -F file=@wkhtmltopdf testing.mini-s3.dev:8080/testing

Retrieve an object from an bucket:

	curl testing.mini-s3.dev:8080/testing

Delete an object from an bucket:

	curl -X DELETE testing.mini-s3.dev:8080/testing
