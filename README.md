# gochat

## Introduction

Gochat is a small IRC-like chat client and server in Go, for practice only. Multiple clients connect to a single server and can chat in a unique channel.

It uses [protobuf](https://github.com/google/protobuf) to serialize the messages.

## Installation

Use `go get`, or clone the repository and install `gochatutil`, `gochatclient` and `gochatserver` separately with `go install`.

Both `gochatserver` and `gochatclient` accept host and port command line variables, for example: `gochaclient -host localhost -port 8081`.

## Dependencies

Only Go is required. The compiled source protobuf file is already included in project.

## TODO list

- Client reconnection in case of connection loss
