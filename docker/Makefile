SHELL := /bin/bash

.PHONY: generate serve generatekey

generate: | build generatekey

serve:
	source env.sh && cat keybaseca.config.gen | envsubst > ../example-keybaseca-volume/keybaseca.config
	source env.sh && docker run -e KEYBASE_USERNAME -e PAPERKEY -v $(PWD)/../example-keybaseca-volume:/mnt:rw ca:latest docker/entrypoint-server.sh

build:
	docker build -t ca -f Dockerfile-ca ..

generatekey:
	source env.sh && cat keybaseca.config.gen | envsubst > ../example-keybaseca-volume/keybaseca.config
	source env.sh && docker run -e FORCE_WRITE -e KEYBASE_USERNAME -e PAPERKEY -v $(PWD)/../example-keybaseca-volume:/mnt:rw ca:latest docker/entrypoint-generate.sh
	@echo -e '\nFor each server that you wish to make accessible to the CA bot:\n\n1. Place the public key in `/etc/ssh/ca.pub`\n2. Add the line `TrustedUserCAKeys /etc/ssh/ca.pub` to `/etc/ssh/sshd_config`\n3. Restart ssh `service ssh restart`'

