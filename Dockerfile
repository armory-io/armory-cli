FROM golang:1.16-alpine3.12 as builder
ARG GITHUB_TOKEN
ARG VERSION

RUN apk update && apk add git make

WORKDIR /workspace
ADD ./ /workspace

ENV GO111MODULE=on GOOS=linux GOARCH=amd64
RUN git config --global url."https://${GITHUB_TOKEN}:x-oauth-basic@github.com/".insteadOf "https://github.com/"
RUN GOPATH= GOPRIVATE=github.com/armory-io/deploy-engine go get github.com/armory-io/deploy-engine@v0.1.0-snapshot.master.fa1dd4a.0.20210923155317-2a3d611b90f9
RUN make build


FROM alpine:3.12
ENV APP=/usr/local/bin/armory/armory \
    USER_UID=1001 \
    USER_NAME=armory

RUN apk update               \
	&& apk add --no-cache ca-certificates bash   \
	&& adduser -D -u ${USER_UID} ${USER_NAME}

COPY --from=builder /workspace/build/dist/linux_amd64 /usr/local/bin/armory
COPY --from=builder /workspace/entrypoint.sh /usr/local/bin/entrypoint.sh

USER ${USER_NAME}
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]