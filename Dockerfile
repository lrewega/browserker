FROM golang:1.14.2-alpine3.11 AS builder

RUN apk add build-base gcc abuild binutils binutils-doc gcc-doc

RUN apk update && apk add --no-cache nmap && \
    echo @edge http://nl.alpinelinux.org/alpine/edge/community >> /etc/apk/repositories && \
    echo @edge http://nl.alpinelinux.org/alpine/edge/main >> /etc/apk/repositories && \
    apk update && \
    apk add --no-cache \
      chromium \
      harfbuzz \
      "freetype>2.8" \
      ttf-freefont \
      nss

WORKDIR /browserker
COPY . .

RUN go build

ENTRYPOINT []
