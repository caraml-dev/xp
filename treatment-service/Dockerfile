FROM golang:1.18-alpine as api-builder
ARG API_BIN_NAME=xp-treatment

RUN apk update && apk add musl-dev gcc

ENV GO111MODULE=on \
    GOOS=linux \
    GOARCH=amd64

ENV PROJECT_ROOT=github.com/caraml-dev/xp/treatment-service

WORKDIR /app
COPY . .

# Build Treatment Service binary
RUN go build \
    -mod=vendor \
    -tags musl \
    -o ./bin/${API_BIN_NAME} \
    -v ${PROJECT_ROOT}/cmd

FROM alpine:latest

RUN apk update && apk add git make bash unzip libtool gcc musl-dev ca-certificates dumb-init tzdata
RUN mkdir -p /opt/xp_treatment

COPY --from=api-builder /app/bin/* /opt/xp_treatment/
RUN chmod +x /opt/xp_treatment/xp-treatment

WORKDIR /opt/xp_treatment
ENTRYPOINT [ "./xp-treatment" ]
