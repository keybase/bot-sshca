# SSH CA Bot

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

# Integration Tests

This project contains integration tests that can be run via `./integrationTest.sh`. Note that prior to running
the integration tests you need to `cp tests/env.sh.example tests/env.sh` and fill in `tests/env.sh`. 

# Getting Started (docker)

```bash
mkdir ~/keybaseca-volume
export SUBTEAM="teamname.subteam_for_ssh"
export KEYBASE_USERNAME="username_of_ca_bot"
export PAPERKEY="paper key for the ca bot"
cd docker/
cat keybaseca.config.gen | envsubst > keybaseca.config
docker build -t ca -f Dockerfile-ca ..
docker run -e KEYBASE_USERNAME -e PAPERKEY -v ~/keybaseca-volume:/mnt:rw ca:latest docker/entrypoint-generate.sh
```

This will output the public key for the CA. 
For each server that you wish to make accessible to the CA bot:

1. Place the public key in `/etc/ssh/ca.pub` 
2. Add the line `TrustedUserCAKeys /etc/ssh/ca.pub` to `/etc/ssh/sshd_config`
3. Restart ssh `service ssh restart`

Now start the chatbot itself:

```bash
docker run -e KEYBASE_USERNAME -e PAPERKEY -v ~/keybaseca-volume:/mnt:rw ca:latest docker/entrypoint-server.sh
```

Now you just run `go run cmd/kssh/kssh.go root@server` in order to SSH into your server. Anyone else in `{TEAM}.ssh` can
also run that command in order to ssh into the server.

# Getting Started (local environment)

In all of these directions, replace `{USER}` with your username and `{TEAM}` with the name of the team that you wish to 
configure this bot for. 

Create a new subteam, `{TEAM}.ssh`. Anyone that is added to this subteam will be granted SSH access. 

Create a new Keybase user named `{TEAM}sshca`. This user will be the bot user that provisions new SSH certificates. 
Export a paper key for this user. Now create a config file at `~/keybaseca.config`:

```yaml
# The ssh user you want to use
ssh_user: root
# The name of the subteam used for granting SSH access
teams: 
- {TEAM}.ssh

# Whether to use an alternate account. Only useful if you are running the chatbot on an account other than the one you are currently using
# Mainly useful for dev work
keybase_home_dir: /tmp/keybase/
keybase_paper_key: "{Put the paper key here}"
keybase_username: {TEAM}sshca
```

Now run `go run cmd/keybaseca/keybaseca.go -c ~/keybaseca.config generate`. This will output the public key for the CA. 
For each server that you wish to make accessible to the CA bot:

1. Place the public key in `/etc/ssh/ca.pub` 
2. Add the line `TrustedUserCAKeys /etc/ssh/ca.pub` to `/etc/ssh/sshd_config`
3. Restart ssh `service ssh restart`

Now start the chatbot itself: `keybase --home /tmp/keybase service & go run cmd/keybaseca/keybaseca.go -c ~/keybaseca.config service` and leave it running.

Now you just run `go run cmd/kssh/kssh.go root@server` in order to SSH into your server. Anyone else in `{TEAM}.ssh` can
also run that command in order to ssh into the server.
