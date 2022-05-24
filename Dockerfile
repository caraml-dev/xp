# Build xp-management binary
FROM golang:1.16-alpine as api-builder
ARG API_BIN_NAME

ENV GOOS=linux GOARCH=amd64

WORKDIR /app
COPY ./management-service .

RUN go build -mod=vendor -o ./bin/${API_BIN_NAME}
RUN chmod +x ./bin/${API_BIN_NAME}

COPY ./api/*.yaml ./bin/

# Build optimized static xp-ui bundle
FROM node:16 as ui-builder

ARG REACT_APP_ENVIRONMENT
ARG REACT_APP_XP_API="/v1"
ARG REACT_APP_MLP_API
ARG REACT_APP_OAUTH_CLIENT_ID
ARG REACT_APP_SENTRY_DSN
ARG REACT_APP_USER_DOCS_URL

RUN mkdir -p /opt/xp_ui
COPY ./ui /opt/xp_ui
WORKDIR /opt/xp_ui

ENV NODE_OPTIONS "--max_old_space_size=4096"

RUN yarn install --network-concurrency 1
RUN yarn build

# Clean image with xp-management binary and production build of xp-ui
FROM alpine:latest

ARG API_BIN_NAME
ENV API_BIN_NAME ${API_BIN_NAME}
ENV XP_UI_APP_DIRECTORY "./xp-ui"
ENV XP_PORT "8080"

EXPOSE ${XP_PORT}

RUN addgroup -S app && adduser -S app -G app
USER app
WORKDIR /app

COPY --from=api-builder /app/bin/* ./
COPY --from=api-builder /app/database ./database
COPY --from=ui-builder /opt/xp_ui/build ${XP_UI_APP_DIRECTORY}/

ENTRYPOINT [ "./xp-management" ]
