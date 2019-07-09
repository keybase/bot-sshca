# SSHCA Bot

This repo contains a work in progress SSH CA bot built on top of Keybase. This project is not yet complete and is not 
ready to be used. 

# Design

There are two binaries contained in this project in the `cmd/` folder. `shared/` is go code that is shared between the 
binaries. 

## keybaseca 

`keybaseca` is the CA server that exposes an interface through Keybase chat. Generate a new CA key by running 
`keybaseca generate`. This will output the CA public key. It also writes a `kssh` (see below) config file to 
`/keybase/team/teamname.ssh/kssh-client.config` such that `kssh` can automatically detect the config file. 
`keybaseca service` starts the CA chatbot service. See `keybaseca/config.go` for a description of the config file. 

## kssh

`kssh` is the replacement SSH binary. It automatically pulls config files from KBFS. 

# Example

```bash
go run cmd/keybaseca/keybaseca.go -c ~/keybaseca.config generate
go run cmd/keybaseca/keybaseca.go -c ~/keybaseca.config service
go run cmd/kssh/kssh.go root@165.22.176.193
```