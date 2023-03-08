FROM ubuntu:latest

RUN mkdir -p /opt/xp_treatment

COPY . /opt/xp_treatment/
COPY ./bin/* /opt/xp_treatment/
RUN chmod +x /opt/xp_treatment/xp-treatment

WORKDIR /opt/xp_treatment
ENTRYPOINT [ "./xp-treatment" ]
