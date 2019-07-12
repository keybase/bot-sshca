#!/bin/bash

# Create two accounts and fill in the blanks below
export BOT_USERNAME="dworkenssh"
export BOT_PAPERKEY="soda outside vote key fiber radio cluster guide false help edge vendor income"

export KSSH_USERNAME="dworken2"
export KSSH_PAPERKEY="tank lonely panel outer lion width blame outdoor door curve fault pool miracle"

# Then create a team that that has subteams "teamname.ssh.staging", "teamname.ssh.prod", "teamname.ssh.root_everywhere"
# $BOT_USERNAME should be in all three teams but $KSSH_USERNAME should only be in staging and root_everywhere.
# Set SUBTEAM to "teamname"
export SUBTEAM="dworken_int"
# Create another subteam that both accounts share
export SUBTEAM_SECONDARY="dworken_int.ssh2"