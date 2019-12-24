Modifications to keybase/bot-sshca to support running the ssh ca in a Kubernetes cluster running on ARM.

Summary of changes:
- `./docker/Dockerfile-ca` moved to `./Dockerfile`. Modify Dockerfile to be multi-arch. This includes building `keybase/client` for ourselves.
- Added `docker-compose-ca.yml` for easy testing/deployment.
- Added `sshca.yaml.example` for kubernetes.
- Didn't modify `./docker/Makefile`. Fend for yourselves.

No modifications to kssh were made.

To build Dockerfile, in project root dir:
`docker build . -t ca`

To run docker-compose, in project root dir:
` docker-compose -f docker-compose-ca.yml up`

========================

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

# Documentation

See the [documentation website](https://keybase-ssh-ca-bot.readthedocs.io/en/latest/) for information on getting started,
best practices, the architecture, and contributing. 
