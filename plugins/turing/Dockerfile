FROM golang:1.18-alpine as plugin-builder

RUN apk update && apk add gcc

WORKDIR /app
COPY . .

ENV GOOS=linux \
    GOARCH=amd64 \
    CGO_ENABLED=1

RUN go build -mod=vendor -o ./bin/plugin

FROM alpine:3.15

ARG TURING_USER="turing"
ARG TURING_USER_GROUP="app"

RUN addgroup -S ${TURING_USER_GROUP} \
    && adduser -S ${TURING_USER} -G ${TURING_USER_GROUP} -H

ENV PLUGIN_NAME "xp"
ENV PLUGINS_DIR "/app/plugins"

COPY --chown=${TURING_USER}:${TURING_USER_GROUP} --from=plugin-builder /app/bin/plugin /go/bin/plugin

CMD ["sh", "-c", "cp /go/bin/plugin ${PLUGINS_DIR}/${PLUGIN_NAME:?variable must be set}"]