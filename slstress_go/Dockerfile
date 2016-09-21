# A syslog stress client.
#
# Sample run
# ~~~~~~~~~~
# docker run -it --rm jmencak/slstress /bin/timeout -t 60 /usr/local/bin/slstress -w 10000000

FROM scratch
MAINTAINER Jiri Mencak
COPY ./slstress-root /
COPY ./slstress /usr/bin/slstress
COPY ./logger.sh /usr/local/bin/
COPY ./docker-entrypoint.sh /
CMD ["/docker-entrypoint.sh"]
