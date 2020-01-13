# Set FORCE_WRITE to false if it is not already set
ifdef FORCE_WRITE
FORCE_WRITE := $(FORCE_WRITE)
else
FORCE_WRITE := false
endif

SHELL := /bin/bash

.PHONY: build generate serve clean reset-permissions confirm-clean env-file-exists ca-key-exists

# Build a new docker image for the CA bot
build: reset-permissions
	docker build -t ca -f Dockerfile-ca ..

# Generate a new CA key
generate: env-file-exists build
	docker run -e FORCE_WRITE=$(FORCE_WRITE) --env-file ./env.list -v $(CURDIR)/example-keybaseca-volume:/mnt:rw ca:latest ./entrypoint-generate.sh
	@echo -e "\nRun these commands on each server that you wish to use with the CA chatbot\n"
	@echo "useradd developer && mkdir -p /home/developer && chown developer:developer /home/developer  # The user that will be used for non-root logins"
	@echo "echo \"`cat $(CURDIR)/example-keybaseca-volume/keybase-ca-key.pub`\" > /etc/ssh/ca.pub"
	@echo "echo \"TrustedUserCAKeys /etc/ssh/ca.pub\" >> /etc/ssh/sshd_config"
	@echo "echo \"AuthorizedPrincipalsFile /etc/ssh/auth_principals/%u\" >> /etc/ssh/sshd_config"
	@echo "chmod g-w /etc   # On some distributions /etc is group writable which will cause SSH to refuse to run"
	@echo "service ssh restart"
	@echo -e "\nSee the getting started docs for information on how to define which teams are allowed to access which servers"

# Start the CA chatbot in the background
serve: env-file-exists ca-key-exists
	docker run -d --restart unless-stopped --env-file ./env.list -v $(CURDIR)/example-keybaseca-volume:/mnt:rw ca:latest ./entrypoint-server.sh
	@echo "Started CA bot service in the background... Use `docker ps` and `docker logs` to monitor it"

# Stop the service
stop:
	docker kill `docker ps -q --filter ancestor=ca`

# Restart the service (useful if you updated env.sh)
restart: stop serve

# Wipe all data
clean: confirm-clean reset-permissions
	@# Sudo since it is likely owned by another use since it was written from a docker container
	sudo rm -rf example-keybaseca-volume/keybaseca*
	sudo rm -rf example-keybaseca-volume/keybase-ca*

# Confirm that the user is okay with deleting their CA key
confirm-clean:
	@echo -n "Are you sure? This will delete the CA key used to connect to your servers [yes/N] " && read ans && [ $${ans:-N} = yes ]

# Reset the permissions on the shared volume. Sudo since the permissions get messed up from the docker container chown-ing it
reset-permissions:
	# Avoid prompting for sudo unless the permissions actually need to be chnaged by piping find to xargs
	find example-keybaseca-volume/ -not -user $$USER | xargs -I {} -- sudo chown -R $$USER {}

# Asserts that env.sh exists
env-file-exists:
	@test -e "env.sh" || (echo "You must create and fill in env.sh prior to running make" && exit 1)

# Assert that a CA key exists
ca-key-exists:
	@test -e "example-keybaseca-volume/keybase-ca-key" || (echo "You must run make generate prior to make serve" && exit 1)
