# SSH CA Bot

The idea here is to control SSH access to servers (without needing to install anything on them) based simply on cryptographically provable membership in Keybase teams. 

SSH supports a concept of certificate authorities (CAs) where you can place a single public key on the server, and the SSH server will allow any connections with keys signed by the CA cert. This is how a lot of large companies manage SSH access securely; users can be granted SSH access to servers without having to change the keys that are deployed on the server. 

This repo provides the pieces for anyone to build this workflow:
1. generation scripts and a guide to set up the Keybase team and server ssh configuration
2. a wrapper around ssh (`kssh`) for any end user to get authenticated using the certificate authority
3. a chatbot (`keybaseca`) which listens in a Keybase team for `kssh` requests. If the requester is in the team, the bot will sign the request with an expiring signature (e.g. 1 hour), and then the provisioned server should authenticate as usual. 

Removing a user's ability to access a server is as simple as removing them from the Keybase team. 

This code is currently a work in progress and this project is not yet complete and is not ready to be used. 

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
the integration tests you need to:

```
cp tests/single-environment/env.sh.example tests/single-environment/env.sh; cp tests/multi-environment/env.sh.example tests/multi-environment/env.sh
``` 

and fill in `tests/single-environment/env.sh` and `tests/multi-environment/env.sh`. 

# Getting Started (single environment mode)

```bash
cd docker/
cp env.sh.example env.sh
keybase signup      # Follow the prompts to create a new Keybase users to use for the SSH CA bot
keybase paperkey    # Generate a new paper key
# Create a new Keybase subteam that this user is in along with anyone else you wish to grant SSH access to
nano env.sh         # Fill in the values including the previously generated paper key
make generate
```

This will output the public key for the CA. 
For each server that you wish to make accessible to the CA bot:

1. Place the public key in `/etc/ssh/ca.pub` 
2. Add the line `TrustedUserCAKeys /etc/ssh/ca.pub` to `/etc/ssh/sshd_config`
3. Restart ssh `service ssh restart`

Now start the chatbot itself:

```bash
make serve
```

Now build kssh and start SSHing!

```bash
go build -o bin/kssh cmd/kssh/kssh.go
sudo cp bin/kssh /usr/local/bin/        # Optional
bin/kssh root@server
```

Anyone else in `{TEAM}.ssh` can also run kssh in order to ssh into the server.

# Multi-Environment Mode

kssh supports a multi-environment mode of operation that allows you to define realms of servers where access is granted based off of 
membership in different teams. Imagine that you have a staging environment that everyone should be granted access to and
a production environment that you want to restrict access to a smaller group of people. For this exercise we'll also set
up a third realm that grants root access to all machines. To configure kssh to work with this setup:

1. Create three subteams: `{TEAM}.ssh.staging`, `{TEAM}.ssh.production`, `{TEAM}.ssh.root_everywhere`
2. Add users to those three teams based off of the permissions you want to grant different users

```bash
cd docker/
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

1. Place the public key in `/etc/ssh/ca.pub` 
2. Add the line `TrustedUserCAKeys /etc/ssh/ca.pub` to `/etc/ssh/sshd_config`
3. Add the line `AuthorizedPrincipalsFile /etc/ssh/auth_principals/%u` to `/etc/ssh/sshd_config`
4. Create the file `/etc/ssh/auth_principals/root` with contents `root_everywhere`
5. Create the file `/etc/ssh/auth_principals/user` with contents `staging`
6. Restart ssh `service ssh restart`

For each server in production:

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

# OS Support

It is recommended to run the server component of this bot on linux and running it in other environments is untested. 
`kssh` is tested and works correctly on linux, macOS, and Windows. If running on windows, note that there is a dependency
on the `ssh` binary being in the path. This can be installed in a number of different ways including 
[Chocolatey](https://chocolatey.org/packages/openssh) or the 
[built in version](https://docs.microsoft.com/en-us/windows-server/administration/openssh/openssh_install_firstuse) on 
modern versions of windows. 

# Getting Started (local environment)
###### Recommended only for development work
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

Now you run `go run cmd/kssh/kssh.go root@server` in order to SSH into your server. Anyone else in `{TEAM}.ssh` can
also run that command in order to ssh into the server.
