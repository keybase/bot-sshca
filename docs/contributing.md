# Contributing and Additional Info

There are two separate binaries built from the code in this repo:

## keybaseca 

`keybaseca` is the CA server that exposes an interface through Keybase chat. This binary is meant to be run in a secure
location. 

```
NAME:
   keybaseca - An SSH CA built on top of Keybase

USAGE:
   keybaseca [global options] command [command options] [arguments...]

VERSION:
   0.0.1

COMMANDS:
     backup    Print the current CA private key to stdout for backup purposes
     generate  Generate a new CA key
     service   Start the CA service in the foreground
     help, h   Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
```

## kssh

`kssh` is the replacement SSH binary. kssh handles provisioning (via the keybaseca-bot) new temporary SSH keys and is meant to be installed on each
user's laptop. 

```
NAME:
   kssh - A replacement ssh binary using Keybase SSH CA to provision SSH keys

USAGE:
   kssh [kssh options] [ssh arguments...]

VERSION:
   0.0.1

GLOBAL OPTIONS:
   --help                Show help
   -v                    Enable kssh and ssh debug logs
   --provision           Provision a new SSH key and add it to the ssh-agent. Useful if you need to run another 
                         program that uses SSH auth (eg scp, rsync, etc)
   --set-default-bot     Set the default bot to be used for kssh. Not necessary if you are only in one team that
                         is using Keybase SSH CA
   --clear-default-bot   Clear the default bot
   --bot                 Specify a specific bot to be used for kssh. Not necessary if you are only in one team that
                         is using Keybase SSH CA
   --set-default-user    Set the default SSH user to be used for kssh. Useful if you use ssh configs that do not set 
					     a default SSH user 
   --clear-default-user  Clear the default SSH user
   --set-keybase-binary  Run kssh with a specific keybase binary rather than resolving via $PATH 
```

## Architecture

#### Config

Keybaseca is configured using environment variables (see docs/env.md for information on all of the options). When keybaseca 
starts, it writes a client config file to `/keybase/team/{teamname for teamname in $TEAMS}/kssh-client.config`. This 
client config file is how kssh determines which teams are using kssh and the needed information about the bot (eg the
channel name, the name of the bot, etc). When keybaseca stops, it deletes all of the client config files. 

kssh reads the client config file in order to determine how to interact with a bot. kssh does not have any user controlled
configuration. It does have one local config file stored in `~/.ssh/kssh-config.json` that is used to store a few settings 
for kssh. By default, this config file is not used. It is only created and meant to be interacted with via the 
`--set-default-bot`, `--clear-default-bot`, `--set-default-user`, `--clear-default-user` flags. 

#### Communication

kssh and keybaseca communicate with each other over Keybase chat. If the `CHAT_CHANNEL` environment variable is 
specified in keybaseca's environment, keybaseca will only accept communication in the specified team and channel. 
This configuration is passed to kssh clients via the client config file(s) stored in KBFS. If the `CHAT_CHANNEL` environment variable
is not specified then keybaseca will accept messages in any channel of any team listed in the `TEAMS` environment variable.
All communication happens via the Go chatbot library. 

Prior to sending a `SignatureRequest`, kssh sends a series of `AckRequest` messages. An `AckRequest` message is sent until 
kssh receives an `Ack` from keybaseca. This is done in order to ensure that kssh has correctly connected to the chat channel
and that the bot is responding to messages. Afterwards, a `SignatureRequest` packet is sent and keybaseca parses it and 
returns a signed key. Note that only public keys and signatures are sent over Keybase chat and private keys never 
leave the devices they were generated on. 

#### SSH Operations

When the ssh-keygen command is available, ssh keys are generated via the ssh-keygen command. In this case, generated
SSH keys are ed25519 keys. If the ssh-keygen command is not available, SSH keys are generated in pure go code and are 
ecdsa keys. 

keybaseca uses the ssh-keygen binary in order to complete all key signing operations. 

#### KBFS

In order to ensure that keybaseca can run inside of docker (which does not support FUSE filesystems without adding
the CAP_SYS_ADMIN permission), all KBFS interactions are done via `keybase fs ...` commands. This makes it so that 
keybaseca can run in unprivileged docker containers. 

## Integration Tests

This project contains integration tests that can be run via `./integrationTest.sh`. The integration tests depend on 
docker and docker-compose. The first time you run them, they will walk you through creating two new live keybase accounts to be 
used in the tests. The credentials for these accounts will be stored in `tests/env.sh`. 
