FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /ehr-server ./cmd/ehr-server

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /ehr-server /usr/local/bin/ehr-server

EXPOSE 8000

ENTRYPOINT ["ehr-server"]
CMD ["serve"]
