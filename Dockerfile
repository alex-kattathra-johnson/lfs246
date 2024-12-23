FROM golang:alpine AS builder

ARG SERVICE

WORKDIR /app/${SERVICE}

COPY utils/go.* /app/utils/
COPY ${SERVICE}/go.* /app/${SERVICE}/
RUN go mod download

COPY utils /app/utils
COPY ${SERVICE} /app/${SERVICE}
RUN go build -o app

FROM scratch
ARG SERVICE
COPY --from=builder /app/${SERVICE}/app /app
COPY --from=builder /app/${SERVICE}/data.json* /data.json
CMD ["/app"]
