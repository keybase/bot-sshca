FROM golang:1.12 as builder

RUN go get github.com/keybase/client/go/keybase
RUN go install -tags production github.com/keybase/client/go/keybase
RUN go build -tags production github.com/keybase/client/go/kbfs/kbfsfuse
RUN cp kbfsfuse /usr/bin/kbfsfuse

WORKDIR /bot-sshca
COPY . ./
RUN go build -o bin/keybaseca src/cmd/keybaseca/keybaseca.go

FROM alpine:latest

# add bash for entrypoint scripts, ssh for ssh-keygen used by the bot, sudo for stepping down to keybase user
RUN apk update && apk add --no-cache bash openssh sudo

# add the keybase user
RUN adduser -s /bin/bash -h /home/keybase -D keybase

# this folder is needed for kbfsfuse
RUN mkdir /keybase && chown -R keybase:keybase /keybase

RUN chown keybase:keybase /usr/local/bin
RUN chown keybase:keybase /home/keybase

USER keybase
WORKDIR /home/keybase

# copy the keybase binaries from previous build step 
COPY --from=builder /go/bin/keybase /usr/local/bin/
COPY --from=builder /go/bin/kbfsfuse /usr/local/bin/
COPY --from=builder /bot-sshca/bin/keybaseca bin/

# copy in entrypoint scripts and fix permissions
COPY ./docker/entrypoint-generate.sh .
COPY ./docker/entrypoint-server.sh .
RUN chown -R keybase:keybase /home/keybase

# Run container as root but only to be able to chown the Docker bind-mount, 
# then immediatetly step down to the keybase user via sudo in the entrypoint scripts
# USER root
