FROM golang:1.24-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /out/taskflow-api ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

ENV SERVER_PORT=8080

COPY --from=builder /out/taskflow-api /app/taskflow-api
COPY --from=builder /src/migrations /app/migrations
COPY --from=builder /src/seeds /app/seeds

EXPOSE 8080

ENTRYPOINT ["/app/taskflow-api"]
