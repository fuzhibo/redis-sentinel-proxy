FROM alpine:3.4
LABEL maintainer="ZhiBo Fu <fuzhibo@tom.com>"

# copy binary
COPY redis-sentinel-proxy-service /usr/local/bin/redis-sentinel-proxy-service

ENTRYPOINT ["/usr/local/bin/redis-sentinel-proxy-service"]
CMD ["-master", "mymaster"]
