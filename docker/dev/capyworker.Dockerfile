# Step 1: Build
FROM golang:1.19-alpine AS build

ENV OUT_D /out

RUN mkdir -p /out
RUN apk add --update  --no-cache \
    bash \
    coreutils \
    ca-certificates \
    git \
    gcc g++ \
    pkgconfig \
    vips-dev

RUN mkdir -p /go/src/github.com/capyfile/capyfile
ADD . /go/src/github.com/capyfile/capyfile/

RUN cd /go/src/github.com/capyfile/capyfile/capyworker && \
    go build -o $OUT_D

# Step 2: App
FROM alpine:3.12

ARG VIPS_VERSION=8.12.2
RUN set -x -o pipefail \
    && wget -O- https://github.com/libvips/libvips/releases/download/v${VIPS_VERSION}/vips-${VIPS_VERSION}.tar.gz | tar xzC /tmp \
    && apk update \
    && apk upgrade \
    && apk add \
    exiftool \
    zlib libxml2 glib gobject-introspection \
    libjpeg-turbo libexif lcms2 fftw giflib libpng \
    libwebp orc tiff poppler-glib librsvg libgsf openexr \
    libheif libimagequant pango \
    && apk add --virtual vips-dependencies build-base \
    zlib-dev libxml2-dev glib-dev gobject-introspection-dev \
    libjpeg-turbo-dev libexif-dev lcms2-dev fftw-dev giflib-dev libpng-dev \
    libwebp-dev orc-dev tiff-dev poppler-dev librsvg-dev libgsf-dev openexr-dev \
    libheif-dev libimagequant-dev pango-dev \
    && cd /tmp/vips-${VIPS_VERSION} \
    && ./configure --prefix=/usr \
                   --disable-static \
                   --disable-dependency-tracking \
                   --enable-silent-rules \
    && make -s install-strip \
    && cd $OLDPWD \
    && rm -rf /tmp/vips-${VIPS_VERSION} \
    && apk del --purge vips-dependencies \
    && rm -rf /var/cache/apk/*

WORKDIR /app
COPY --from=build /out/capyworker /app/capyworker

RUN adduser -D user
USER user:user

ENTRYPOINT ["/app/capyworker"]