# Introduction

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
