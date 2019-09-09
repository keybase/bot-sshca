# Advanced Configuration

The SSH CA bot is configured via environment variables. This documents lists the different environment variables 
used by the bot and their purpose. 

## Environment Variables

### TEAMS

The `TEAMS` environment variable configures which teams the SSH CA bot will use to grant SSH access. 

Examples:

```bash
export TEAMS="team.ssh"
export TEAMS="team.ssh.prod"
export TEAMS="team.ssh.prod,team.ssh.staging,team.ssh.root_everywhere"
```

### CA_KEY_LOCATION

The `CA_KEY_LOCATION` environment variable configures where the CA bot will store the CA key. It is recommended to 
ensure that the CA key is stored in a secure location. Defaults to `/mnt/keybase-ca-key`. 

Examples:

```bash
export CA_KEY_LOCATION="/etc/cakey"
export CA_KEY_LOCATION="~/secure/cakey"
```

### KEY_EXPIRATION

The `KEY_EXPIRATION` environment variable configures the validity length of keys signed by the bot. A key provisioned
via kssh is valid for this length of time before kssh will automatically reprovision another key. It is recommended
to keep the key expiration window to a relatively short period of time. By default, signed key s expire after one 
hour. Valid formats are +30m, +1h, +5h, +1d, +3d, +1w, etc

Examples:

```bash
export KEY_EXPIRATION="+2h"
export KEY_EXPIRATION="+10m"
export KEY_EXPIRATION="+1w"     # not recommended to set it to a time period this long
```

### LOG_LOCATION

The `LOG_LOCATION` environment variable configures where logs from the CA bot will be stored. It is recommended to store these logs in a
secure location for audit purposes. One potential option is storing logs in a KBFS subteam dedicated to admins.
If not set, logs go to stdout.

Examples:

```bash
export LOG_LOCATION="/keybase/team/teamname.ssh.admin/keybaseca_audit.log"
```

### STRICT_LOGGING

The `STRICT_LOGGING` environment variable defines the behavior of the bot if it fails to save an audit log entry.
By default, if the CA bot fails to write a log to a file it will simply send it to stdout. If it is critical to 
maintain correct audit logs, the `STRICT_LOGGING` option will cause the CA bot to panic and shutdown if it is 
unable to save logs.

Examples:

```bash
export STRICT_LOGGING="true"
export STRICT_LOGGING="false"
```

### CHAT_CHANNEL

The `CHAT_CHANNEL` environment variable controls where communication between the bot and users will take place.
It specifies a specific team and channel. This may be useful if subteams are being used for more purposes
than just SSH access. For example, one could use team.prod to grant SSH production access and for a secret
sharing bot used to share other credentials. One would then want to configure the CA bot to use a separate
channel (eg #ssh-provision) for provision requests so that the general channel could be used for lower volume
purposes. Note that this means that all users of the SSH bot must have access to this channel.

Examples:

```bash
export CHAT_CHANNEL="team.prod#ssh-provision"
export CHAT_CHANNEL="team.ssh_bot#general"
```

### Announcement

The `ANNOUNCEMENT` environment variable contains a string that will be announced in all of the configured teams when
the bot is started. This is useful if you would like the bot to announce the fact it has started and granted access in 
a given team. The `ANNOUNCEMENT` environment variable supports a number of templating variables that will be instantiated
based off of the current config. These are:

* `$USERNAME` will be replaced with the username of the bot
* `$CURRENT_TEAM` will be replaced with the team that the message is being sent in
* `$TEAMS` will be replaced with the comma separated list of teams that the bot is running in

Examples:

```bash
export ANNOUNCEMENT="SSH CA bot starting up..."
export ANNOUNCEMENT="Hello! I'm $USERNAME and I'm an SSH bot! See github.com/keybase/bot-sshca for information on using Keybase for SSH."
export ANNOUNCEMENT="Hello! I'm $USERNAME and I'm an SSH bot! I'm currently listening in $TEAMS."
export ANNOUNCEMENT="Hello! I'm $USERNAME and I'm an SSH bot! Being in $CURRENT_TEAM will grant you SSH access to certain servers. Reach out to @dworken for more information."
```

## Developer Options

These environment variables are mainly useful for dev work. For security reasons, it is recommended always to run a 
production CA chat bot on an isolated machine. These options make it possible to run a CA chat bot on a machine where 
you currently are logged into another account. 

Examples:

```bash
KEYBASE_HOME_DIR: /tmp/keybase/
KEYBASE_PAPERKEY: "paper key goes here"
KEYBASE_USERNAME: teamname-sshca-bot
```
