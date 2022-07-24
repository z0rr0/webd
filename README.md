# WebD

![Version](https://img.shields.io/github/tag/z0rr0/webd.svg)
![License](https://img.shields.io/github/license/z0rr0/webd.svg)

WebD is a wrapper over Go standard library static HTTP server.

## Build

```sh
go build -ldflags "-X main.Tag=`git tags | tail -1`" .
```

## Run

```sh
chmod u+x webd
./webd
```

Parameters

```
Usage of ./webd:
  -host string
        host to listen on (default "127.0.0.1")
  -password string
        password for basic auth
  -port uint
        port to listen on (default 8080)
  -root string
        root directory to serve (default ".")
  -timeout duration
        timeout for requests (default 5s)
  -user string
        username for basic auth
  -version
        show version
```

## License

This source code is governed by a Apache v2 license that can be found
in the [LICENSE](https://github.com/z0rr0/webd/blob/main/LICENSE) file.
