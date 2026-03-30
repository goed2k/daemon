# 多阶段构建：静态链接二进制，运行时使用 Alpine
# 构建：docker build -t goed2kd .
# 运行：docker run --rm -p 18080:18080 -p 4661:4661 -p 4662:4662/udp -v goed2kd-data:/app/data goed2kd

FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=0
RUN go build -trimpath -ldflags="-s -w" -o /out/goed2kd ./cmd/goed2kd

FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata \
	&& adduser -D -H -s /sbin/nologin appuser

COPY --from=builder /out/goed2kd /usr/local/bin/goed2kd

USER appuser
WORKDIR /app

# RPC、引擎 TCP/UDP（与默认配置一致，可按需改 publish）
EXPOSE 18080/tcp 4661/tcp 4662/udp

ENTRYPOINT ["/usr/local/bin/goed2kd"]
CMD ["-config", "/app/data/config/config.json"]
