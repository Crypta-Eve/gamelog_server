FROM golang:1.18-alpine as builder
LABEL maintainer="Crypta Electrica <crypta@crypta.tech>"
RUN apk update && apk add alpine-sdk git && rm -rf /var/cache/apk/*
RUN mkdir /build
COPY . /build
WORKDIR /build
ENV GO111MODULE on
ENV GOPROXY https://proxy.golang.org
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o main .


FROM alpine
RUN adduser -S -D -H -h /app appuser && apk --no-cache add ca-certificates
USER appuser
COPY --from=builder /build/main /app/
WORKDIR /app
ENV GIN_MODE=release
EXPOSE 8080
ENTRYPOINT ["./main"]
