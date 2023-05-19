# Step 1: Build
FROM golang:1.19-alpine AS build

ENV OUT_D /out

RUN mkdir -p /out
RUN apk add --update  --no-cache \
    bash \
    coreutils \
    ca-certificates \
    git

RUN mkdir -p /go/src/github.com/capyfile/capyfile
ADD . /go/src/github.com/capyfile/capyfile/

RUN cd /go/src/github.com/capyfile/capyfile/capysvr && \
    go build -o $OUT_D

# Step 2: App
FROM alpine:3.12

RUN apk add --update --no-cache \
    ca-certificates \
    exiftool

WORKDIR /app
COPY --from=build /out/capysvr /app/capysvr

RUN adduser -D user
USER user:user

EXPOSE 8024

ENTRYPOINT ["/app/capysvr"]