# Set FORCE_WRITE to false if it is not already set
ifdef FORCE_WRITE
FORCE_WRITE := $(FORCE_WRITE)
else
FORCE_WRITE := false
endif
ifdef VOLUME_NAME
VOLUME_NAME := ${VOLUME_NAME}
else
VOLUME_NAME := keybase-sshca-volume
endif

SHELL := /bin/bash

.PHONY: build generate serve clean confirm-clean create-volume env-file-exists ca-key-exists

# Function to fetch containers which are using a specific docker volume.
# In this case, the volume that is mounted to the CA.
GET_CONTAINERS := $(shell docker ps -qa | xargs docker container inspect -f '{{ .Name }} {{ .HostConfig.Binds }}' | grep $(VOLUME_NAME) | awk '{print $$1}')

# Build a new docker image for the CA bot.
build:
	docker build -t ca -f Dockerfile-ca ..

# Generate a new CA key.
generate: env-file-exists build create-volume
	source env.sh && docker run --rm -e FORCE_WRITE=$(FORCE_WRITE) -e KEYBASE_USERNAME -e KEYBASE_PAPERKEY -e TEAM -v $(VOLUME_NAME):/mnt:rw ca:latest docker/entrypoint-generate.sh
	@echo -e "\n------------------------------------------------------------------------------------------------\n"
	@echo -e "Run these commands on each server that you wish to use with the CA chatbot:\n"
	@echo "useradd developer && mkdir -p /home/developer && chown developer:developer /home/developer  # The user that will be used for non-root logins"
	@echo "echo \"`docker run --rm -v $(VOLUME_NAME):/tmp ca cat /tmp/keybase-ca-key.pub`\" > /etc/ssh/ca.pub"
	@echo "echo \"TrustedUserCAKeys /etc/ssh/ca.pub\" >> /etc/ssh/sshd_config"
	@echo "echo \"AuthorizedPrincipalsFile /etc/ssh/auth_principals/%u\" >> /etc/ssh/sshd_config"
	@echo "chmod g-w /etc   # On some distributions /etc is group writable which will cause SSH to refuse to run"
	@echo "service ssh restart"
	@echo -e "\nSee the README for information on how to define which teams are allowed to access which servers\n"
	@echo "------------------------------------------------------------------------------------------------"

# Start the CA chatbot in the background
serve: env-file-exists ca-key-exists build
	source env.sh && docker run -d --restart unless-stopped -e KEYBASE_USERNAME -e KEYBASE_PAPERKEY -e TEAM -v $(VOLUME_NAME):/mnt ca:latest docker/entrypoint-server.sh
	@echo "Started CA bot service in the background... Use 'docker ps' and 'docker logs -f <container name>' to monitor it"

# Wipe all data.
# This is done by first stopping and deleting all containers that uses the volume that the certificates are stored in.
# When the containers are gone, the volume is removed.
clean: confirm-clean
	@echo "Checking for running containers..."
	@[[ -z "$(GET_CONTAINERS)" ]] || (docker stop $(GET_CONTAINERS); docker rm $(GET_CONTAINERS))
	@docker volume ls | grep -q $(VOLUME_NAME) && (docker volume rm $(VOLUME_NAME) 2>&1 && echo "Clean done.") || echo "Already clean!"

# Confirm that the user is okay with deleting their CA key.
confirm-clean:
	@echo -n "Are you sure? This will delete the CA key used to connect to your servers [yes/N] " && read ans && [ $${ans:-N} = yes ]

# Creates a new volume to be used by the keybase ssh-ca server to store certificates in.
# If a volume already exists, this will do nothing.
create-volume:
	@docker volume ls | grep -q $(VOLUME_NAME) || docker volume create $(VOLUME_NAME)

# Asserts that env.sh exists
env-file-exists:
	@test -e "env.sh" || (echo "You must create and fill in env.sh prior to running make" && exit 1)

# Assert that a CA key exists.
# This is done by first checking if there is a docker volume created with the name defined via $VOLUME_NAME,
# If it exists, a check is done to see if there is a public key in the volume or not.
ca-key-exists:
	@docker volume ls | grep -q $(VOLUME_NAME) || (echo "You must run make generate prior to make serve" && exit 1)
	@docker run --rm -v $(VOLUME_NAME):/tmp ca cat /tmp/keybase-ca-key.pub > /dev/null 2>&1 || (echo "You must run make generate prior to make serve" && exit 1)
