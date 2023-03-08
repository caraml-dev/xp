FROM alpine:latest

RUN apk update && apk add git make bash unzip libtool gcc ca-certificates dumb-init tzdata libc6-compat

RUN mkdir -p /opt/xp_treatment

COPY . /opt/xp_treatment/
COPY ./bin/* /opt/xp_treatment/
RUN chmod +x /opt/xp_treatment/xp-treatment

WORKDIR /opt/xp_treatment
ENTRYPOINT [ "./xp-treatment" ]
