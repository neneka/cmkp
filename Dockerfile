FROM golang:1.18.3-alpine AS server-build
RUN apk add --no-cache git
WORKDIR /go/src/github.com/wtks/cmkp/backend
COPY ./backend/go.* ./
RUN go mod download
COPY ./backend .
RUN CGO_ENABLED=0 go build -o /cmkp

# runtime image
FROM alpine:3.10
RUN apk add --no-cache ca-certificates openssl
ENV DOCKERIZE_VERSION v0.6.1
RUN wget https://github.com/jwilder/dockerize/releases/download/$DOCKERIZE_VERSION/dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && tar -C /usr/local/bin -xzvf dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && rm dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz
ARG MYSQL_DATABASE
ARG MYSQL_HOSTNAME
ARG MYSQL_PASSWORD
ARG MYSQL_USERNAME
EXPOSE 5000
COPY --from=server-build /cmkp /cmkp
ENTRYPOINT ["/cmkp"]
