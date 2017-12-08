FROM golang:alpine AS build

ARG CGO_ENABLED=1
ARG GOOS=linux
ARG GOARCH=amd64

WORKDIR /usr/src
COPY . /usr/src

RUN apk add --no-cache git build-base
RUN go get -d ./... \
    && go build --ldflags="-s" -o chaos-monkey
RUN apk del --no-cache git build-base

FROM scratch

COPY --from=build /usr/src/chaos-monkey /chaos-monkey

ENTRYPOINT ["/chaos-monkey"]