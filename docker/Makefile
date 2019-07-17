SHELL := /bin/bash

.PHONY: build generate serve clean reset-permissions confirm-clean

build: reset-permissions
	docker build -t ca -f Dockerfile-ca ..

generate: build
	source env.sh && python3 .run.py generate

serve: build
	source env.sh && python3 .run.py serve

clean: confirm-clean reset-permissions
	@# Sudo since it is likely owned by another use since it was written from a docker container
	sudo rm -rf ../example-keybaseca-volume/keybaseca*
	sudo rm -rf ../example-keybaseca-volume/keybase-ca*

confirm-clean:
	@echo -n "Are you sure? This will delete the CA key used to connect to your servers [yes/N] " && read ans && [ $${ans:-N} = yes ]

reset-permissions:
	sudo chown -R $$USER ../example-keybaseca-volume/