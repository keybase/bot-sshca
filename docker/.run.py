# Run by make in order to generate keybaseca config files based on values from the environment
import os
import sys

GENERATE_SINGLE_INSTRUCTIONS = """
For each server that you wish to make accessible to the CA bot:

1. Place the public key in `/etc/ssh/ca.pub`
2. Add the line `TrustedUserCAKeys /etc/ssh/ca.pub` to `/etc/ssh/sshd_config`
3. Restart ssh `service ssh restart`"""


def is_multi_environment_mode(subteams):
    return "," in subteams

def generate_config(subteams):
    if is_multi_environment_mode(subteams):
        configTemplate = "keybaseca-multi-environment.config.gen"
    else:
        configTemplate = "keybaseca-single-environment.config.gen"
    os.system("cat %s | envsubst > ../example-keybaseca-volume/keybaseca.config" % configTemplate)


def generate(subteams, username, paperkey, argv):
    generate_config(subteams)
    ret = os.system("bash -c 'docker run -e FORCE_WRITE=${FORCE_WRITE:-false} -e KEYBASE_USERNAME -e PAPERKEY -v $(pwd)/../example-keybaseca-volume:/mnt:rw ca:latest docker/entrypoint-generate.sh'")
    if ret != 0:
        print("\nFailed to generate key!")
        exit(ret)
    if is_multi_environment_mode(subteams):
        print("\nSee README.md for instructions on how to install this key on your servers")
    else:
        print(GENERATE_SINGLE_INSTRUCTIONS)

def serve(subteams, username, paperkey, argv):
    generate_config(subteams)
    os.system("bash -c 'docker run -e KEYBASE_USERNAME -e PAPERKEY -v $(pwd)/../example-keybaseca-volume:/mnt:rw ca:latest docker/entrypoint-server.sh'")

if __name__ == "__main__":
    subteams = os.environ.get('SUBTEAMS', None)
    username = os.environ.get('KEYBASE_USERNAME', None)
    paperkey = os.environ.get('PAPERKEY', None)
    if not subteams or not username or not paperkey:
        print("Missing environment variables! Did you fill in env.sh correctly?")
        exit(1)

    if sys.argv[1] == "generate":
        generate(subteams, username, paperkey, sys.argv[2:])
    elif sys.argv[1] == "serve":
        serve(subteams, username, paperkey, sys.argv[2:])
    else:
        print("Cannot handle argument %s" % sys.argv[1])
