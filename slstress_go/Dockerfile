# A syslog stress client.
#
# Sample run
# ~~~~~~~~~~
# docker run -it --rm jmencak/slstress /bin/timeout -t 60 /usr/local/bin/slstress -w 10000000

FROM scratch
MAINTAINER Jiri Mencak
COPY ./slstress-root /
COPY ./slstress /usr/local/bin/slstress
COPY ./logger.sh /usr/local/bin/
COPY ./docker-entrypoint.sh /
#RUN sed -i '/^StrictHostKeyChecking.*/d' /etc/ssh/ssh_config && \
#    sed -i '$a StrictHostKeyChecking no' /etc/ssh/ssh_config
CMD ["/docker-entrypoint.sh"]
