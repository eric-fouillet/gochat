# gochat

## Introduction
Gochat is a small chat client and server in Go, for practice only.

It uses [protobuf](https://github.com/google/protobuf) to serialize the messages.

## Installation

Use `go get`, or clone the repository and install `gochatutil`, `gochatclient` and `gochatserver` separately with `go install`.

## Dependencies

Only Go is required. The compiled source protobuf file is already included in project.

## TODO list

- Client reconnection in case of connection loss
