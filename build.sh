LEVELDB_PREFIX=/Users/silas/opt/opt/leveldb

export GOPATH="`pwd`:$GOPATH"
export CGO_CFLAGS="-I$LEVELDB_PREFIX/include"
export CGO_LDFLAGS="-L$LEVELDB_PREFIX/lib"
go build -o mini-s3 main
