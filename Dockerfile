FROM golang:1.12 as builder

RUN go get github.com/keybase/client/go/keybase
RUN go install -tags production github.com/keybase/client/go/keybase
RUN go build -tags production github.com/keybase/client/go/kbfs/kbfsfuse
RUN cp kbfsfuse /usr/bin/kbfsfuse

WORKDIR /sshca
COPY . ./
RUN go build -o bin/keybaseca src/cmd/keybaseca/keybaseca.go

FROM ubuntu:18.04

RUN useradd -ms /bin/bash keybase
RUN chown keybase:keybase /usr/bin
# RUN chown keybase:keybase /home/keybase

USER keybase
WORKDIR /home/keybase
COPY --from=builder /go/bin/keybase /usr/bin/keybase
COPY --from=builder /go/src/github.com/keybase/client/packaging/linux/run_keybase /usr/bin/run_keybase
COPY --from=builder /usr/bin/kbfsfuse /usr/bin/kbfsfuse
COPY --from=builder /sshca/bin/keybaseca bin/keybaseca
COPY --from=builder /sshca/docker docker
