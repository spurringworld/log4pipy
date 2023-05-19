FROM golang:1.18.3 AS builder
MAINTAINER leon "lqhome@126.com"
# docker中的工作目录
WORKDIR /app
# 将当前目录同步到docker工作目录下，也可以只配置需要的目录和文件（配置目录、编译后的程序等）
COPY . /app
# 由于所周知的原因，某些包会出现下载超时。这里在docker里也使用go module的代理服务
ENV GO111MODULE=on
ENV GOPROXY="https://goproxy.io"
# 指定编译完成后的文件名，可以不设置使用默认的，最后一步要执行该文件名
RUN cd /app && CGO_ENABLED=0 go build -o log4pipy

FROM alpine
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/log4pipy /app/log4pipy

EXPOSE 9876
# 这里跟编译完的文件名一致
ENTRYPOINT  ["./log4pipy"]
