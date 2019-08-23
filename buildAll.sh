#!/bin/bash

export VERSION="`cat VERSION`-`git rev-parse --short HEAD`"

go build -ldflags "-X main.VersionNumber=$VERSION" -o bin/kssh-linux src/cmd/kssh/kssh.go
go build -ldflags "-X main.VersionNumber=$VERSION" -o bin/keybaseca-linux src/cmd/keybaseca/keybaseca.go

GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.VersionNumber=$VERSION" -o bin/kssh-mac src/cmd/kssh/kssh.go
GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.VersionNumber=$VERSION" -o bin/keybaseca-mac src/cmd/keybaseca/keybaseca.go

GOOS=windows GOARCH=amd64 go build -ldflags "-X main.VersionNumber=$VERSION" -o bin/kssh-windows src/cmd/kssh/kssh.go
GOOS=windows GOARCH=amd64 go build -ldflags "-X main.VersionNumber=$VERSION" -o bin/keybaseca-windows src/cmd/keybaseca/keybaseca.go
