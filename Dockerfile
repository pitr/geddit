FROM golang:1.15-alpine as builder

RUN apk add --no-cache build-base git

WORKDIR /app
COPY go.mod go.sum /app/
RUN go mod download

ADD . /app
RUN make clean build.local


FROM alpine:latest

RUN apk add --no-cache sqlite

COPY --from=builder /app/build/geddit /
ADD geddit-prod.crt /geddit.crt
ADD geddit-prod.key /geddit.key

CMD ["/geddit"]