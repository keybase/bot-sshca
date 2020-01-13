# Getting Started

kssh allows you to define realms of servers where access is granted based off of
membership in different teams. Imagine that you have a staging environment that everyone should be granted access to and
a production environment that you want to restrict access to a smaller group of people. For this exercise we'll also set
up a third realm that grants root access to all machines. 

Start by creating a new Keybase user to use for the CA chatbot:

```bash
keybase signup      # Creates a new Keybase user to use for the SSH CA bot
keybase paperkey    # Generate a new paper key
```

Note that this system will not work if you attempt to use the same user for the CA chatbot as for kssh. It is required
to use distinct users. 

Then create `{TEAM}.ssh.staging`, `{TEAM}.ssh.production`, `{TEAM}.ssh.root_everywhere` as new Keybase subteams
and add the bot to those subteams. Add users to those subteams based off of the permissions you wish to grant
different users

On a secured server (note that this server only needs docker installed) that you wish to use to run the CA chatbot:

```bash
git clone git@github.com:keybase/bot-sshca.git
cd bot-sshca/docker/
cp env.list.example env.list
nano env.list       # Fill in the values including the previously generated paper key
make generate       # Generate a new CA key
```

Running `make generate` will output a list of configuration steps to run on each server you wish to use with the CA chatbot. 
These commands create a new user to use with kssh (the `developer` user), add the CA's public key to the server, and 
configure the server to trust the public key. 

Now you must define a mapping between Keybase teams and the users on the servers that members of those teams are
allowed to access. If you wish to make the user `foo` on your server available to anyone in `team.ssh.bar`,
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

## Updating environment variables

If you update any environment variables, it is necessary to restart the keybaseca service. This can be done 
by running `make restart`. Note that it is not required to re-run `make generate`. 

Note that this means `kssh` will not work for a brief period of time while the container restarts.

