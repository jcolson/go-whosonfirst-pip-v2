FROM sncrpc/spatialite:latest as build-env

RUN apk add go \
    wget

ADD . /go-whosonfirst-pip-v2
RUN cd /go-whosonfirst-pip-v2 && make tools

# ENV dataset whosonfirst-data-admin-us-latest
# real dataset, huge!
ENV dataset whosonfirst-data-latest
# small data set to test with
RUN mkdir /usr/local/data && cd /usr/local/data && wget https://dist.whosonfirst.org/sqlite/${dataset}.db.bz2 && bunzip2 ${dataset}.db.bz2

# Create a minimal instance
FROM alpine

# copy binaries
COPY --from=build-env /go-whosonfirst-pip-v2/bin/wof-pip-server /bin/wof-pip-server
COPY --from=build-env /go-whosonfirst-pip-v2/docker/entrypoint.sh /bin/entrypoint.sh
COPY --from=build-env /go-whosonfirst-pip-v2/docker/spatialite.sql /bin/spatialite.sql
COPY --from=build-env /usr/local/data /usr/local/data

# document the fact that we use port 8080
EXPOSE 8080

WORKDIR /bin/

ENTRYPOINT /bin/entrypoint.sh
