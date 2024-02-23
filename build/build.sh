#!/bin/bash

# 默认架构
ARCHITECTURES=("arm" "arm64" "amd64" "386")

# 如果没有提供参数，则构建所有架构
if [ $# -eq 0 ]; then
    for ARCH in "${ARCHITECTURES[@]}"; do
        ./build.sh "$ARCH"
    done
    exit 0
fi

# 如果提供了参数，则检查是否有效
for ARCH in "$@"; do
    valid=false
    for valid_arch in "${ARCHITECTURES[@]}"; do
        if [ "$ARCH" = "$valid_arch" ]; then
            valid=true
            break
        fi
    done
    if ! $valid; then
        echo "无效的架构: ${ARCH}"
        echo "有效的架构包括: " + "${ARCHITECTURES[@]}"
        exit 1
    fi
done

# 构建提供的架构
for ARCH in "$@"; do
    ./build.sh "$ARCH"
done
