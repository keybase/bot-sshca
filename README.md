# SSH CA Bot

# This code is currently a work in progress and this project is not yet complete and is not ready to be used. 

See [keybase.io/blog/keybase-ssh-ca](https://keybase.io/blog/keybase-ssh-ca) for a full announcement and description
of the code in this repository. 

The idea here is to control SSH access to servers (without needing to install anything on them) based simply on 
cryptographically provable membership in Keybase teams. 

SSH supports a concept of certificate authorities (CAs) where you can place a single public key on the server, 
and the SSH server will allow any connections with keys signed by the CA cert. This is how a lot of large companies 
manage SSH access securely; users can be granted SSH access to servers without having to change the keys that are 
deployed on the server. 

# Getting Started 

kssh allows you to define realms of servers where access is granted based off of 
membership in different teams. Imagine that you have a staging environment that everyone should be granted access to and
a production environment that you want to restrict access to a smaller group of people. For this exercise we'll also set
up a third realm that grants root access to all machines. To configure kssh to work with this setup:

1. Create three subteams: `{TEAM}.ssh.staging`, `{TEAM}.ssh.production`, `{TEAM}.ssh.root_everywhere`
2. Add users to those three teams based off of the permissions you want to grant different users

On a secured server that you wish to use to run the CA chatbot:

```bash
git clone git@github.com:keybase/bot-sshca.git
cd bot-sshca/docker/
cp env.sh.example env.sh
keybase signup      # Follow the prompts to create a new Keybase users to use for the SSH CA bot
keybase paperkey    # Generate a new paper key
# Create `{TEAM}.ssh.staging`, `{TEAM}.ssh.production`, `{TEAM}.ssh.root_everywhere` as new Keybase subteams
# and add the bot to those subteams. Add users to those subteams based off of the permissions you wish to grant
# different users
nano env.sh         # Fill in the values including the previously generated paper key
make generate
```

This will output the public key for the CA. 

For each server in staging:

0. Create a user named `user`
1. Place the public key in `/etc/ssh/ca.pub` 
2. Add the line `TrustedUserCAKeys /etc/ssh/ca.pub` to `/etc/ssh/sshd_config`
3. Add the line `AuthorizedPrincipalsFile /etc/ssh/auth_principals/%u` to `/etc/ssh/sshd_config`
4. Create the file `/etc/ssh/auth_principals/root` with contents `root_everywhere`
5. Create the file `/etc/ssh/auth_principals/user` with contents `staging`
6. Restart ssh `service ssh restart`

For each server in production:

0. Create a user named `user`
1. Place the public key in `/etc/ssh/ca.pub` 
2. Add the line `TrustedUserCAKeys /etc/ssh/ca.pub` to `/etc/ssh/sshd_config`
3. Add the line `AuthorizedPrincipalsFile /etc/ssh/auth_principals/%u` to `/etc/ssh/sshd_config`
4. Create the file `/etc/ssh/auth_principals/root` with contents `root_everywhere`
5. Create the file `/etc/ssh/auth_principals/user` with contents `production`
6. Restart ssh `service ssh restart`

Now start the chatbot itself:

```bash
make serve
```

Now build kssh and start SSHing!

```bash
go build -o bin/kssh cmd/kssh/kssh.go
sudo cp bin/kssh /usr/local/bin/        # Optional
bin/kssh user@staging-server-ip         # If in {TEAM}.ssh.staging
bin/kssh user@production-server-ip      # If in {TEAM}.ssh.production
bin/kssh root@server                    # If in {TEAM}.ssh.root_everywhere
```

We recommend building kssh yourself and distributing it among your team. 

# OS Support

It is recommended to run the server component of this bot on linux and running it in other environments is untested. 
`kssh` is tested and works correctly on linux, macOS, and Windows. If running on windows, note that there is a dependency
on the `ssh` binary being in the path. This can be installed in a number of different ways including 
[Chocolatey](https://chocolatey.org/packages/openssh) or the 
[built in version](https://docs.microsoft.com/en-us/windows-server/administration/openssh/openssh_install_firstuse) on 
modern versions of windows. 

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

`kssh` is the replacement SSH binary. kssh handles provisioning new SSH keys and is meant to be installed on each
user's laptop. 

```
NAME:
   kssh - A replacement ssh binary using Keybase SSH CA to provision SSH keys

USAGE:
   kssh [kssh options] [ssh arguments...]

VERSION:
   0.0.1

GLOBAL OPTIONS:
   --help,               Show help
   --provision           Provision a new SSH key and add it to the ssh-agent. Useful if you need to run another 
                         program that uses SSH auth (eg scp, rsync, etc)
   --set-default-team    Set the default team to be used for kssh. Not necessary if you are only in one team that
                         is using Keybase SSH CA
   --clear-default-team  Clear the default team
   --team                Specify a specific team to be used for kssh. Not necessary if you are only in one team that
                         is using Keybase SSH CA
```

## Architecture

#### Config

Keybaseca is configured using environment variables (see env.md for information on all of the options). When keybaseca 
starts, it writes a client config file to `/keybase/team/{teamname for teamname in $TEAMS}/kssh-client.config`. This 
client config file is how kssh determines which teams are using kssh and the needed information about the bot (eg the
channel name, the name of the bot, etc). When keybaseca stops, it deletes the client config file. 

kssh reads the client config file in order to determine how to interact with a bot. kssh does not have any user controlled
config files. It does have one local config file stored in `~/.ssh/cssh.cache` that is used to store the default team
if the `--set-default-team` flag is set. This config file is not meant to be manually edited and is only meant to be 
interacted with via the `--set-default-team` and `--clear-default-team` flags. 

#### Communication

kssh and keybaseca communicate with each other over Keybase chat. If the `CHAT_CHANNEL` environment variable is specified,
keybaseca will only accept communication in the specified team and channel. If the `CHAT_CHANNEL` environment variable
is not specified then keybaseca will accept messages in any channel of any team listed in the `TEAMS` environment variable.
All communication happens via the Go chatbot library. 

Prior to sending a `SignatureRequest`, kssh sends a series of AckRequest messages. An AckRequest message is sent until 
kssh receives an Ack from keybaseca. This is done in order to ensure that kssh has correctly connected to the chat channel
and that the bot is responding to messages. In order to ensure that kssh is receiving an Ack in response to the AckRequests
that it sent, the AckRequest includes the username of the user using kssh. Afterwards, a SignatureRequest packet is sent
and keybaseca parses it and returns a signed key. Note that only public keys and signatures are sent over Keybase chat
and private keys never leave the devices they were generated on. 

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