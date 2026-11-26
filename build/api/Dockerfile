# Build context is the repository root:
#   docker build -f build/api/Dockerfile -t rodritv-api .
FROM golang:1.27-alpine AS builder

ARG TARGETARCH

WORKDIR /build

# Dependencies first, so this layer is cached until go.mod/go.sum change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w" -o /out/startup ./cmd/api

FROM alpine:3.24

RUN addgroup -g 1000 -S rodritv \
    && adduser -u 1000 -S -G rodritv -h /home/rodritv rodritv \
    && install -d -o rodritv -g rodritv \
        /home/rodritv/.config \
        /home/rodritv/.config/api

WORKDIR /home/rodritv

# Owned by root: the service user can run these files but not modify them
COPY --from=builder /out/startup bin/startup
COPY entrypoint.sh entrypoint.sh
RUN chmod 755 entrypoint.sh

USER 1000:1000

EXPOSE 9290

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -q -O /dev/null http://127.0.0.1:9290/healthz || exit 1

ENTRYPOINT ["/home/rodritv/entrypoint.sh"]
