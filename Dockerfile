
# Build in a stock Go builder container
FROM golang:1.9-alpine as builder

ADD . /cryptopepe-bot-api

RUN apk add --no-cache \
  --repository http://dl-3.alpinelinux.org/alpine/edge/testing \
  gcc g++ make libc6-compat

RUN apk add --no-cache \
  --repository http://dl-3.alpinelinux.org/alpine/edge/testing \
  librsvg-dev glib-dev expat-dev libpng-dev fftw-dev


ARG VIPS_VERSION=8.6.3

ENV VIPS_DIR=/vips
ENV PKG_CONFIG_PATH=${VIPS_DIR}/lib/pkgconfig:$PKG_CONFIG_PATH

RUN apk update && apk add --no-cache openssl ca-certificates && mkdir -p ${GOPATH}/src && \
    wget -O- https://github.com/libvips/libvips/releases/download/v${VIPS_VERSION}/vips-${VIPS_VERSION}.tar.gz | tar xzC /tmp

RUN cd /tmp/vips-${VIPS_VERSION} && \
    ./configure \
        --disable-static \
        --disable-dependency-tracking \
        --without-python \
        --prefix=${VIPS_DIR} && \
    make && \
    make install


RUN cd /cryptopepe-bot-api && build/env.sh go build -v -o build/cryptopepe-bot-api .

# Pull into a second stage deploy alpine container
FROM alpine:latest

RUN apk add --no-cache \
  --repository http://dl-3.alpinelinux.org/alpine/edge/testing \
  fftw libpng librsvg expat glib libgsf

# Add the vipslib we compiled from source in the builder
COPY --from=builder /vips/lib/ /usr/local/lib

COPY --from=builder /cryptopepe-bot-api/build/cryptopepe-bot-api /usr/local/bin/

# Copy builder files
COPY --from=builder /cryptopepe-bot-api/vendor/cryptopepe.io/cryptopepe-svg/builder/tmpl /app/tmpl
COPY --from=builder /cryptopepe-bot-api/vendor/cryptopepe.io/cryptopepe-svg/builder/builder.tmpl /app/builder.tmpl
COPY --from=builder /cryptopepe-bot-api/mappings.json /app/mappings.json

# Expose port 80, dokku will pick it up and proxy external port 80/443 to port 80 of this app.
EXPOSE 80

WORKDIR /app
ENTRYPOINT ["cryptopepe-bot-api"]