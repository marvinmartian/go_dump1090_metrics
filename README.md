# go_dump1090_exporter

This project contains the code to export Prometheus metrics from a running ADSB dump1090 instance.

## Clone the project

```
$ git clone https://github.com/marvinmartian/go_dump1090_metrics.git
```

## Build Instructions
```
$ cd go_dump1090_metrics
$ CGO_ENABLED=false GOOS=linux GOARCH=arm GOARM=6 go build -a -tags netgo -ldflags '-w' -o go_dump1090_exporter src/main.go
```
Target is intended to be a Raspberry Pi. Change the ARM version if required
