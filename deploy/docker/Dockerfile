# Stage 1: Build Go binaries
FROM golang:1.22-alpine AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /oss-bot ./cmd/bot
RUN CGO_ENABLED=0 go build -o /oss-cli ./cmd/oss

# Stage 2: Final image
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=go-builder /oss-bot /usr/local/bin/oss-bot
COPY --from=go-builder /oss-cli /usr/local/bin/oss
COPY prompts/ /prompts/
ENV OSS_PROMPTS_DIR=/prompts
EXPOSE 8090
ENTRYPOINT ["oss-bot"]
