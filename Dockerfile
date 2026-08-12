# ===== 阶段 1：构建前端 =====
FROM node:22-alpine AS web-build
WORKDIR /web
# 先拷依赖清单，利用 Docker 层缓存
COPY web/package.json web/package-lock.json* ./
RUN npm ci || npm install
COPY web/ .
RUN npm run build

# ===== 阶段 2：编译 Go 后端（嵌入前端产物）=====
FROM golang:1.26-alpine AS go-build
WORKDIR /src
# 后端纯标准库，无外部依赖，直接拷源码
COPY go.mod ./
COPY api/ ./api/
COPY main.go ./
COPY middleware/ ./middleware/
# 嵌入前端构建产物（必须与 main.go 的 go:embed 目录一致）
COPY --from=web-build /web/dist ./web/dist
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/vibe-server .

# ===== 阶段 3：精简运行时 =====
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata \
    && adduser -D -u 10001 app \
    && mkdir -p /data \
    && chown -R app:app /data
# 运行用户（非 root）
COPY --from=go-build /out/vibe-server /usr/local/bin/vibe-server
USER app
ENV PORT=8080 \
    VIBE_GUID_FILE=/data/.device-guid
WORKDIR /app
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --retries=3 \
  CMD wget -q -O /dev/null http://127.0.0.1:8080/ || exit 1
CMD ["vibe-server"]
