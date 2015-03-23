FROM debian:jessie
MAINTAINER anders@janmyr.com

COPY coreos-docker-listener /
CMD ["/coreos-docker-listener"]
