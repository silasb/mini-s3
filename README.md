## Mini-S3

A mini version of the S3 datastore built with Go.

## Build for production

	GOOS=linux GOARCH=amd64 GOPATH=`pwd`:$GOPATH go build main.go

## Build for development

	GOPATH=`pwd`:$GOPATH go run main.go

## Setup

cp config.sample config

## API support

Place a file or information into an object:

	curl -X POST -F file=@wkhtmltopdf testing.mini-s3.dev:8080/testing

Retrieve an object from an bucket:

	curl testing.mini-s3.dev:8080/testing

Delete an object from an bucket:

	curl -X DELETE testing.mini-s3.dev:8080/testing
