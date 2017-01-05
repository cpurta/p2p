FROM ubuntu:latest

MAINTAINER Chris Purta cpurta@gmail.com

RUN apt-get update && \
    apt-get install -y ca-certificates && \
    mkdir -p /opt

RUN chmod +x ./p2p
ADD ./p2p /opt

EXPOSE 8888

ENTRYPOINT ["/opt/p2p"]
