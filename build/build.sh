#!/bin/bash

# 如果没有提供架构参数，则默认为 amd64
ARCH="${1:-amd64}"

# 构建 Docker 镜像
docker build -t gotrace:build .

# 创建一个用于存储数据的目录
mkdir -p ./data

# 在 data 目录中创建构建脚本
echo "cd /GoTrace && GOOS=linux GOARCH=${ARCH} go build -o GoTrace . && cp -arf ./config /data && cp GoTrace /data" > ./data/build.sh
chmod +x ./data/build.sh

# 使用构建的镜像在 Docker 容器中运行，并等待容器执行完毕
CONTAINER_ID=$(docker run -itd --rm -v $(pwd)/data:/data --net=host gotrace:build)
docker wait $CONTAINER_ID

# 清理工作
docker rmi -f gotrace:build
rm -rf ./data/build.sh


# 打印成功消息
echo "恭喜！编译成功。在 ./data 目录中找到二进制文件和相应的配置文件。"
