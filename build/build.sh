#!/bin/bash

docker build -t gotrace:build --build-arg GOARCH=$($1) .

mkdir ./data

docker run -itd --rm -v $(pwd)/data:/data --net=host gotrace:build

docker rmi -f gotrace:build
echo "恭喜你编译成功在./data目录下就是编译好的二进制文件和config对应的"