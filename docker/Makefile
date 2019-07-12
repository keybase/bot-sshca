SHELL := /bin/bash

.PHONY: generate generate-advanced serve serve-advanced generatekey generatekey-advanced clean

generate: | build generatekey
generate-advanced: | build generatekey-advanced

serve:
	source env.sh && docker run -e KEYBASE_USERNAME -e PAPERKEY -v $(PWD)/../example-keybaseca-volume:/mnt:rw ca:latest docker/entrypoint-server.sh

serve-advanced:
	source env-advanced.sh && docker run -e KEYBASE_USERNAME -e PAPERKEY -v $(PWD)/../example-keybaseca-volume:/mnt:rw ca:latest docker/entrypoint-server.sh

build:
	docker build -t ca -f Dockerfile-ca ..

generatekey:
	source env.sh && cat keybaseca.config.gen | envsubst > ../example-keybaseca-volume/keybaseca.config
	source env.sh && docker run -e KEYBASE_USERNAME -e PAPERKEY -v $(PWD)/../example-keybaseca-volume:/mnt:rw ca:latest docker/entrypoint-generate.sh
	@echo -e '\nFor each server that you wish to make accessible to the CA bot:\n\n1. Place the public key in `/etc/ssh/ca.pub`\n2. Add the line `TrustedUserCAKeys /etc/ssh/ca.pub` to `/etc/ssh/sshd_config`\n3. Restart ssh `service ssh restart`'

generatekey-advanced:
	source env-advanced.sh && cat keybaseca-advanced.config.gen | envsubst > ../example-keybaseca-volume/keybaseca.config
	source env-advanced.sh && docker run -e KEYBASE_USERNAME -e PAPERKEY -v $(PWD)/../example-keybaseca-volume:/mnt:rw ca:latest docker/entrypoint-generate.sh

clean:
    # Sudo since it is likely owned by another use since it was written from a docker container
	sudo rm -rf ../example-keybaseca-volume/keybaseca*
	sudo rm -rf ../example-keybaseca-volume/keybase-ca*
	sudo chown -R $$USER ../example-keybaseca-volume/