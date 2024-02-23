#!/bin/bash

docker build -t gotrace:build .

mkdir ./data
echo "cd /GoTrace && go build GOOS=linux GOARCH="$($1) + "-o GoTrace . && cp -arf ./config /data && cp GoTrace /data ">> ./data/build.sh
chmod +x ./data/build.sh
docker run -itd --rm -v $(pwd)/data:/data --net=host gotrace:build

docker rmi -f gotrace:build
rm -rf ./data/build.sh
echo "恭喜你编译成功在./data目录下就是编译好的二进制文件和config对应的"