#!/bin/bash

docker build -t gotrace:build .

mkdir data

docker run -it --rm -v $(pwd):/data --net=host GoTrace:build

docker rmi -f GoTrace:build
echo "恭喜你编译成功在./data目录下就是编译好的二进制文件和config对应的"
