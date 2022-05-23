FROM debian:buster-slim

RUN apt-get update && apt-get install -y ca-certificates && update-ca-certificates

RUN mkdir -p /opt/xp_treatment

COPY . /opt/xp_treatment/
COPY ./bin/* /opt/xp_treatment/

WORKDIR /opt/xp_treatment
ENTRYPOINT [ "./xp-treatment" ]
