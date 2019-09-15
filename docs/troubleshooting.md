# Troubleshooting

This file contains some general directions and thoughts on troubleshooting the code in this repo. This is not meant
to be a comprehensive troubleshooting guide and is only a jumping off point. 

## `make generate` refuses to overwrite an existing key

In order to force `make generate` to overwrite the existing CA key (note that this will delete the existing CA
key which means kssh will not work with any servers it currently works with), simply run:

```
FORCE_WRITE=true make generate
```

## kssh is slow, but it works

When kssh starts, it has to search every team you are in for a `kssh-client.config` file which specifies the information
that is needed in order to communicate with the CA chatbot. If you are only in a few teams, this is relatively fast 
(1-2 seconds for <10 teams) but this can become much slower as the number of teams increases (6 seconds for 100 teams
in my benchmarks). This complex start up procedure can be avoided by setting a default bot via 
`kssh --set-default-bot cabotname` which should reduce kssh's startup time considerably. 

## kssh times out

If kssh times out with a message similar to:

```
Generating a new SSH key...
Requesting signature from the CA....
Failed to get a signed key from the CA: timed out while waiting for a response from the CA
```

It means that for whatever reason, kssh is not receiving a response from the CA chatbot when it sends messages in 
Keybase chat. First, ensure that the CA chatbot is currently running. Next, attempt to determine what is happening
by inspecting the chat messages inside of the teams configured with the chatbot. You should see a series of `Ack` and 
`AckRequest` messages going back and forth prior to a `Signature_Request:` and a `Signature_Response:` exchange. Ensure 
that you and the chatbot are in the correct teams such that they can read and respond to the messages. In addition,
review the log output from the keybaseca chatbot. Note that it is required to run the keybaseca chatbot as a different
user than you are using for kssh. 

## SSH rejects the connection

This likely means that you have not configured the SSH server correctly. Confirm that on the SSH server you are trying to access:

* `/etc/ssh/ca.pub` has an SSH public key in it
* `/etc/ssh/auth_principals/username-of-ssh-user` has the name of your Keybase team in it (or multiple comma separated keybase teams)
* `/etc/ssh/sshd_config` has the below two lines somewhere in it:

```
TrustedUserCAKeys /etc/ssh/ca.pub
AuthorizedPrincipalsFile /etc/ssh/auth_principals/%u
```

If that all looks good, review the getting started directions and ensure that you have followed the steps correctly. 
Additionally, it is recommended to compare your sshd_config file with the stock one for your OS to look for any 
non-standard config options. For example, setting `UsePAM no` will prevent the SSH CA from working. 
([sshca.md](./sshca.md) also has some additional information on how SSH CAs work that may
be helpful). If you would like to follow an example, see the code in the `tests/` directory which contains integration 
tests (focus on Dockerfile-sshd for an example SSH server setup). If none of that works, the best strategy is to run
SSH on the server on an alternate port and review the debug information. On the server run `/usr/sbin/sshd -dd -D -p 2222`
and on the client run `kssh -p 2222 user@server` and inspect the debug logs.  

## Keybase is down

If Keybase is down, the bot will not work since it relies on Keybase chat for communication. In this scenario, you can 
manually sign SSH keys with the CA key. This can be done via `keybaseca sign --public-key /path/to/key.pub`. Alternatively,
this can be done manually without relying on any of the tooling in this repository. To do so, place the CA private key 
in `~/cakey` and the CA public key in `~/cakey.pub`. Then run the command:

```bash
ssh-keygen \
  # The location of the ca key:
  -s ~/cakey  \
  # A unique ID for each key. Used to audit key usage
  -I unique-key-id \
  # The comma separated list of principals you wish to sign the key for. Eg "team.ssh.prod,team.ssh.staging,team.ssh.root_everywhere"
  -n "team.ssh.prod,team.ssh.staging,team.ssh.root_everywhere" \
  # How long the signature is valid for. +1d means one day. Valid units are `h` for hour, `d` for day, `w` for week
  -V +1d \
  # Specify the password on the CA key (if exported via `keybaseca backup` there is no password)
  -N "" \
  # The location of the public key you wish to sign
  /path/to/key.pub
```

You can then use the signed SSH key to SSH via `ssh -i /path/to/key.pub user@server`. 

## Default Users and kssh --provision

Default users are implemented using a custom SSH config file that inherits from the default one. This means that if you
run:

```bash
kssh --set-default-user developer
kssh --provision
scp foo server:~/
```

It will not use the default user. There are two ways to work around this. If you do not need the default user to be kssh
specific (eg if kssh is your primary way of accessing certain servers), then you can simply configure this default user
globally by adding the below lines to `~/.ssh/config`

```bash
Host *
  User developer
```

If you do not want to do this, you can run scp with the kssh specific config file via:

```bash
scp -F ~/.ssh/kssh-config foo server:~/
```

Or analogously for rsync:

```bash
rsync -e "ssh -F $HOME/.ssh/kssh-config" foo server:~/
```

It may be useful to define aliases in your bashrc to simplify this:

```bash
alias kscp='kssh --provision && scp -F ~/.ssh/kssh-config'
alias krsync='kssh --provision && rsync -e "ssh -F $HOME/.ssh/kssh-config"'
```

## Other

For any other issues, please open a Github issue or ping @dworken on Keybase! We want to make this project as reliable
as possible so please let us know if there are any ways we can improve the project. 