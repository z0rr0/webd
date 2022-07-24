# WebD

WebD is a wrapper over Go standard library static HTTP server.

## Build

```sh
go build -ldflags "-X main.Tag=`git tags | tail -1`" .
```

## Run

```sh
chmod u+x webd
# show help using: ./webd -h
./webd
```