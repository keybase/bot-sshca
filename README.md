# SSH CA Bot

[![License](https://img.shields.io/badge/license-BSD-success.svg)](https://opensource.org/licenses/BSD-3-Clause)
[![CircleCI](https://circleci.com/gh/keybase/bot-sshca.svg?style=shield)](https://circleci.com/gh/keybase/bot-sshca)
[![Go ReportCard](https://goreportcard.com/badge/github.com/keybase/bot-sshca)](https://goreportcard.com/report/github.com/keybase/bot-sshca)

See [keybase.io/blog/keybase-ssh-ca](https://keybase.io/blog/keybase-ssh-ca) for a full announcement and description
of the code in this repository. 

This repository provides the tooling to control SSH access to servers (without needing to install anything 
on them) based simply on cryptographically provable membership in Keybase teams. 

SSH supports a concept of certificate authorities (CAs) where you can place a single public key on the server, 
and the SSH server will allow any connections with keys signed by the CA cert. This is how a lot of large companies 
manage SSH access securely; users can be granted SSH access to servers without having to change the keys that are 
deployed on the server. 

This repo provides the pieces for anyone to build this workflow on top of Keybase:
1. generation scripts and a guide to set up the Keybase team and server ssh configuration
2. a wrapper around ssh (`kssh`) for any end user to get authenticated using the certificate authority
3. a chatbot (`keybaseca`) which listens in a Keybase team for `kssh` requests. If the requester is in the team, the bot will sign the request with an expiring signature (e.g. 1 hour), and then the provisioned server should authenticate as usual.

Removing a user's ability to access a server is as simple as removing them from the Keybase team.

# Getting Started

kssh allows you to define realms of servers where access is granted based off of
membership in different teams. Imagine that you have a staging environment that everyone should be granted access to and
a production environment that you want to restrict access to a smaller group of people. For this exercise we'll also set
up a third realm that grants root access to all machines. To configure kssh to work with this setup, we will set it up 
according to this diagram:

![Architecture Diagram](https://raw.githubusercontent.com/keybase/bot-sshca/master/docs/Architecture%20Diagram.png "Architecture Diagram")

On a secured server that you wish to use to run the CA chatbot:

```bash
git clone git@github.com:keybase/bot-sshca.git
cd bot-sshca/docker/
cp env.sh.example env.sh
keybase signup      # Creates a new Keybase user to use for the SSH CA bot
keybase paperkey    # Generate a new paper key
# Create `{TEAM}.ssh.staging`, `{TEAM}.ssh.production`, `{TEAM}.ssh.root_everywhere` as new Keybase subteams
# and add the bot to those subteams. Add users to those subteams based off of the permissions you wish to grant
# different users
nano env.sh         # Fill in the values including the previously generated paper key
make generate       # Generate a new CA key
```

Running `make generate` will output a list of configuration steps to run on each server you wish to use with the CA chatbot. 
These commands create a new user to use with kssh (the `developer` user), add the CA's public key to the server, and 
configure the server to trust the public key. 

Now you must define a mapping between Keybase teams the users and servers that they are
allowed to access. If you wish to make the user foo available to anyone in team.ssh.bar,
create the file `/etc/ssh/auth_principals/foo` with contents `team.ssh.bar`. 

More concretely following the current example setup:

* For each server in your staging environment:
  1. Create the file `/etc/ssh/auth_principals/root` with contents `{TEAM}.ssh.root_everywhere`
  2. Create the file `/etc/ssh/auth_principals/developer` with contents `{TEAM}.ssh.staging`
* For each server in your production environment:
  1. Create the file `/etc/ssh/auth_principals/root` with contents `{TEAM}.ssh.root_everywhere`
  2. Create the file `/etc/ssh/auth_principals/developer` with contents `{TEAM}.ssh.production`

Now on the server where you wish to run the chatbot, start the chatbot itself:

```bash
make serve    # Runs inside of docker for ease of use
```

Now download the kssh binary and start SSHing! See https://github.com/keybase/bot-sshca/releases to download the most 
recent version of kssh for your platform. 

```bash
sudo mv kssh-{platform} /usr/local/bin/kssh 
sudo chmod +x /usr/local/bin/kssh

kssh developer@staging-server-ip        # If in {TEAM}.ssh.staging
kssh developer@production-server-ip     # If in {TEAM}.ssh.production
kssh root@server                        # If in {TEAM}.ssh.root_everywhere
```

We recommend building kssh yourself and distributing the binary among your team (perhaps in Keybase Files!). 

# OS Support

It is recommended to run the server component of this bot on linux and running it in other environments is untested.
`kssh` is tested and works correctly on linux, macOS, and Windows. If running on windows, note that there is a dependency
on the `ssh` binary being in the path. This can be installed in a number of different ways including 
[Chocolatey](https://chocolatey.org/packages/openssh) or the 
[built in version](https://docs.microsoft.com/en-us/windows-server/administration/openssh/openssh_install_firstuse) on 
modern versions of windows. 

# Using kssh with jumpboxes and bastion hosts

kssh should work correctly with jumpboxes and bastion hosts as long as they are configured to trust the SSH CA and the usernames are correct. For example:

```
kssh -J developer@jumpbox.example.com developer@server.internal
```

This can also be made easier by setting the kssh default ssh-username locally, then you won't have to specify it for each server. 

```
kssh --set-default-user developer
kssh -J jumpbox.example.com server.internal
```

# Contributing

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

On nix style systems (linux and MacOS), kssh generates and uses ed25519 private keys using the ssh-keygen binary. In 
order to remove the dependency on ssh-keygen for windows systems (since it is not always installed on windows), kssh 
will generate a 2048 bit RSA key when running on Windows. 

keybaseca uses the ssh-keygen binary in order to complete all key signing operations. 

#### KBFS

In order to ensure that keybaseca can run inside of docker (which does not support FUSE filesystems without adding
the CAP_SYS_ADMIN permission), all KBFS interactions are done via `keybase fs ...` commands. This makes it so that 
keybaseca can run in unprivileged docker containers. 

# Integration Tests

This project contains integration tests that can be run via `./integrationTest.sh`. The integration tests depend on 
docker and docker-compose. The first time you run them, they will walk you through creating two new accounts to be 
used in the tests. The credentials for these accounts will be stored in `tests/env.sh`. 
