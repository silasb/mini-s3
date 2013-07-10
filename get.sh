LEVELDB_PREFIX=/Users/silas/opt/opt/leveldb

CGO_CFLAGS="-I$LEVELDB_PREFIX/include" CGO_LDFLAGS="-L$LEVELDB_PREFIX/lib" go get github.com/jmhodges/levigo
