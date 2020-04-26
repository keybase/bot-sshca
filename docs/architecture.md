# Architecture

The Keybase SSH CA system works according to this diagram:

![Architecture Diagram](https://raw.githubusercontent.com/keybase/bot-sshca/master/docs/Architecture%20Diagram.png "Architecture Diagram")

Note that this means that you do not need to modify your servers in any
way or run any additional processes on your servers other than a standard 
OpenSSH daemon. 

## Network Architecture

Since all communication between the kssh client and the SSH CA server happens over Keybase chat, it is possible (and recommended)
to firewall off the SSH CA server (where this bot is running) so it cannot be reached from the general internet. Additionally, note that the SSH servers
that trust the SSH CA do not need to communicate with Keybase's servers or with the CA server and thus it is also possible
to firewall off the SSH servers from the general internet. Clients running kssh need to have Keybase running locally with
a connection to Keybase's servers. 
