# Build stage
FROM golang:alpine3.19 as build

RUN apk add --no-cache --no-progress tesseract-ocr-dev leptonica-dev g++ gcc make

ADD ./src /app

WORKDIR /app

RUN make tidy

RUN make build

# App Stage
FROM alpine:3.19

RUN apk add --no-cache --no-progress tesseract-ocr tesseract-ocr-data-slk poppler-utils

COPY --from=build ../tmp/menucko /app/menucko

ADD ./static /app/static

WORKDIR /app

ENTRYPOINT ./menucko
