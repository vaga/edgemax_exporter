# EdgeMAX Exporter for Prometheus

This is a simple server that scrapes EdgeMAX stats and exports them via HTTP for Prometheus consumption.

Rewriting of [mdlayher/edgemax_exporter (archived repo)](https://github.com/mdlayher/edgemax_exporter) with [gorilla/websocket](https://github.com/gorilla/websocket) implementation.

## Getting started

To run it:
```
./edgemax_exporter [flags]
```

Help on flags:
```
./edgemax_exporter --help
```

## Running tests

```
go test ./...
```
