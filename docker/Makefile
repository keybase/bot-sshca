# Set FORCE_WRITE to false if it is not already set
ifdef FORCE_WRITE
FORCE_WRITE := $(FORCE_WRITE)
else
FORCE_WRITE := false
endif

SHELL := /bin/bash

.PHONY: build generate serve clean reset-permissions confirm-clean

# Build a new docker image for the CA bot
build: reset-permissions
	docker build -t ca -f Dockerfile-ca ..

# Generate a new CA key
generate: build
	source env.sh && docker run -e FORCE_WRITE=$(FORCE_WRITE) -e KEYBASE_USERNAME -e KEYBASE_PAPERKEY -v $(CURDIR)/../example-keybaseca-volume:/mnt:rw ca:latest docker/entrypoint-generate.sh

serve: build
	source env.sh && docker run -e KEYBASE_USERNAME -e KEYBASE_PAPERKEY -v $(CURDIR)/../example-keybaseca-volume:/mnt:rw ca:latest docker/entrypoint-server.sh

# Wipe all data
clean: confirm-clean reset-permissions
	@# Sudo since it is likely owned by another use since it was written from a docker container
	sudo rm -rf ../example-keybaseca-volume/keybaseca*
	sudo rm -rf ../example-keybaseca-volume/keybase-ca*

# Confirm that the user is okay with deleting their CA key
confirm-clean:
	@echo -n "Are you sure? This will delete the CA key used to connect to your servers [yes/N] " && read ans && [ $${ans:-N} = yes ]

# Reset the permissions on the shared volume. Sudo since the permissions get messed up from the docker container chown-ing it
reset-permissions:
	sudo chown -R $$USER ../example-keybaseca-volume/