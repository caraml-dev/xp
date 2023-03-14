FROM golang:1.18-alpine as api-builder
ARG API_BIN_NAME=xp-treatment

RUN apk update && apk add musl-dev gcc

ENV GO111MODULE=on \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /app
COPY . .

# Build Treatment Service binary
RUN go build \
    -mod=vendor \
    -tags musl \
    -o ./bin/${API_BIN_NAME}

FROM alpine:latest

RUN apk update && apk add git make bash unzip libtool ca-certificates dumb-init tzdata
RUN mkdir -p /opt/xp_treatment

COPY --from=api-builder /app/bin/* /opt/xp_treatment/

WORKDIR /opt/xp_treatment

ARG API_BIN_NAME=xp-treatment
ENV XP_API_BIN "./${API_BIN_NAME}"

ENTRYPOINT ["sh", "-c", "${XP_API_BIN} \"$@\"", "--"]
