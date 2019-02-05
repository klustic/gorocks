# GOROX
Golang implementation of [AROX](https://github.com/klustic/AlmondRocks)

## Building
### Executable
```
$ mkdir -p $GOPATH/github.com/klustic
$ git clone https://github.com/klustic/gorocks
$ go build github.com/klustic/gorocks/server
$ go build github.com/klustic/gorocks/relay
```

## Usage
```
$ ./relay -h
$ ./server -h
```

See AlmondRocks documentation for more info on `server` vs `relay`

## Credits
- [yamux](https://github.com/hashicorp/yamux)
- [go-socks5](https://github.com/armon/go-socks5)
