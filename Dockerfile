# syntax = docker/dockerfile:experimental
# Build Container
FROM golang:1.22 as builder

ENV GO111MODULE on
ENV GOPRIVATE=github.com/latonaio
ENV CGO_ENABLED=0
WORKDIR /go/src/github.com/latonaio

COPY . .
RUN go mod download
RUN go build -o data-platform-api-article-deletes-rmq-kube ./

# Runtime Container
FROM alpine:3.19
RUN apk add --no-cache libc6-compat
ENV SERVICE=data-platform-api-article-deletes-rmq-kube \
    APP_DIR="${AION_HOME}/${POSITION}/${SERVICE}"

WORKDIR ${AION_HOME}

COPY --from=builder /go/src/github.com/latonaio/data-platform-api-article-deletes-rmq-kube .

CMD ["./data-platform-api-article-deletes-rmq-kube"]
