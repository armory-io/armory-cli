FROM alpine:3.12
ARG BUILD_PATH
ENV APP=/usr/local/bin/armory \
    USER_UID=1001 \
    USER_NAME=armory

RUN apk update               \
	&& apk add --no-cache ca-certificates bash   \
	&& adduser -D -u ${USER_UID} ${USER_NAME}

WORKDIR /usr/local/bin
COPY ${BUILD_PATH} ./
COPY ./entrypoint.sh ./

USER ${USER_NAME}
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]