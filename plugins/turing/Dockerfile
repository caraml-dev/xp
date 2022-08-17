FROM alpine:latest

ARG TURING_USER="turing"
ARG TURING_USER_GROUP="app"

RUN addgroup -S ${TURING_USER_GROUP} \
    && adduser -S ${TURING_USER} -G ${TURING_USER_GROUP} -H

RUN mkdir -p /go/bin
RUN mkdir -p /app/plugins

ENV PLUGIN_NAME "xp"
ENV PLUGINS_DIR "/app/plugins"

COPY --chown=${TURING_USER}:${TURING_USER_GROUP} ./bin/plugin /go/bin/plugin

CMD ["sh", "-c", "cp /go/bin/plugin ${PLUGINS_DIR}/${PLUGIN_NAME:?variable must be set}"]