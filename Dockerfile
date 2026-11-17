# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM golang:alpine AS builder
WORKDIR /src
COPY . ./
ENV GO111MODULE=on
ENV GOPROXY="https://goproxy.cn,direct"
ARG VERSION=dev
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -trimpath -ldflags "-X project/pkg/global.SYSTEM_VERSION=${VERSION}" -o /out/ThingsPanel-Go .

FROM alpine:latest
ARG VERSION=dev
LABEL description="ThingsPanel Go Backend"
WORKDIR /go/src/app
RUN apk update && apk add --no-cache tzdata
COPY --from=builder /out/ThingsPanel-Go ./ThingsPanel-Go
COPY sql ./sql
# All-in-One 安装器（prepare.py）会从容器里 docker cp 取这些迁移文件与版本号。
COPY pkg/global/global.go ./pkg/global/
COPY encore-api/automation/migrations ./encore-api/automation/migrations
COPY encore-api/devicemetadata/migrations ./encore-api/devicemetadata/migrations
COPY encore-api/l3-core/notification/migrations ./encore-api/l3-core/notification/migrations
COPY encore-api/fileassets/migrations ./encore-api/fileassets/migrations
COPY encore-api/l3-core/connectorcatalog/migrations ./encore-api/l3-core/connectorcatalog/migrations
COPY encore-api/marketinstall/migrations ./encore-api/marketinstall/migrations
COPY static ./static
COPY files/iphone-1.png files/iphone-2.png files/logo.png ./files/
COPY configs/conf.yml configs/casbin.conf configs/messages.yaml configs/messages_str.yaml ./configs/
COPY configs/rsa_key/private_key.pem configs/rsa_key/public.pem ./configs/rsa_key/
RUN mkdir -p files/logs
EXPOSE 9999
RUN chmod +x ThingsPanel-Go && test "$(./ThingsPanel-Go --version)" = "$VERSION"
ENTRYPOINT [ "./ThingsPanel-Go" ]
