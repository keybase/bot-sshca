FROM alpine:3.11 AS builder

# add dependencies
RUN apk update && apk add --no-cache go curl git musl-dev gcc

# build keybase binary
WORKDIR /go
ENV GOPATH=/go
ENV KEYBASE_VERSION=5.0.0
RUN go get -d github.com/keybase/client/go/keybase
RUN cd src/github.com/keybase/client/go/keybase && git checkout v$KEYBASE_VERSION
RUN go install -tags production github.com/keybase/client/go/keybase

# build kbfsfuse binary (we won't use FUSE but the bot needs KBFS for exchanging Team config files)
RUN go install -tags production github.com/keybase/client/go/kbfs/kbfsfuse

# build keybaseca
WORKDIR /bot-sshca
COPY . ./
RUN go build -o bin/keybaseca src/cmd/keybaseca/keybaseca.go

FROM alpine:3.11

# add bash for entrypoint scripts, ssh for ssh-keygen used by the bot, sudo for stepping down to keybase user
RUN apk update && apk add --no-cache bash openssh sudo

# add the keybase user
RUN adduser -s /bin/bash -h /home/keybase -D keybase
RUN chown keybase:keybase /home/keybase

# this folder is needed for kbfsfuse
RUN mkdir /keybase && chown -R keybase:keybase /keybase

USER keybase
WORKDIR /home/keybase

# copy the keybase binaries from previous build step 
COPY --from=builder --chown=keybase:keybase /go/bin/keybase /usr/local/bin/
COPY --from=builder --chown=keybase:keybase /go/bin/kbfsfuse /usr/local/bin/
COPY --from=builder --chown=keybase:keybase /bot-sshca/bin/keybaseca bin/

# copy in entrypoint scripts and fix permissions
COPY ./docker/entrypoint-generate.sh .
COPY ./docker/entrypoint-server.sh .
