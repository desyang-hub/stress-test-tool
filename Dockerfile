FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY . .

RUN apk add --no-cache git
RUN go build -ldflags="-s -w" -o /stresstest ./cmd/stresstest/

FROM alpine:3.20

RUN apk add --no-cache ca-certificates
COPY --from=builder /stresstest /usr/local/bin/stresstest

ENTRYPOINT ["stresstest"]
