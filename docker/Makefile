ifdef FORCE_WRITE
FORCE_WRITE := $(FORCE_WRITE)
else
FORCE_WRITE := false
endif

SHELL := /bin/bash

.PHONY: build generate serve clean reset-permissions confirm-clean config

build: reset-permissions
	docker build -t ca -f Dockerfile-ca ..

config:
	source env.sh && cat keybaseca.config.gen | envsubst > ../example-keybaseca-volume/keybaseca.config

generate: build config
	source env.sh && docker run -e FORCE_WRITE=$(FORCE_WRITE) -e KEYBASE_USERNAME -e PAPERKEY -v $(CURDIR)/../example-keybaseca-volume:/mnt:rw ca:latest docker/entrypoint-generate.sh

serve: build config
	source env.sh && docker run -e KEYBASE_USERNAME -e PAPERKEY -v $(CURDIR)/../example-keybaseca-volume:/mnt:rw ca:latest docker/entrypoint-server.sh

clean: confirm-clean reset-permissions
	@# Sudo since it is likely owned by another use since it was written from a docker container
	sudo rm -rf ../example-keybaseca-volume/keybaseca*
	sudo rm -rf ../example-keybaseca-volume/keybase-ca*

confirm-clean:
	@echo -n "Are you sure? This will delete the CA key used to connect to your servers [yes/N] " && read ans && [ $${ans:-N} = yes ]

reset-permissions:
	sudo chown -R $$USER ../example-keybaseca-volume/