FROM golang:alpine AS builder

ARG SERVICE

COPY utils /app/utils
COPY ${SERVICE} /app/${SERVICE}
WORKDIR /app/${SERVICE}

RUN go build -o app

FROM scratch
ARG SERVICE
COPY --from=builder /app/${SERVICE}/app /app
COPY --from=builder /app/${SERVICE}/data.json* /data.json
CMD ["/app"]
