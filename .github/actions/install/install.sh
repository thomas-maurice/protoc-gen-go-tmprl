#!/bin/bash

VERSION="${VERSION:-$(curl -s https://api.github.com/repos/thomas-maurice/protoc-gen-go-tmprl/releases/latest | jq .name -r)}"
RUNNER_ARCH="${RUNNER_ARCH:=X64}"
ARCH_NAME="amd64"
OS_NAME="linux"

case "$RUNNER_ARCH" in
    "X86")
    ARCH_NAME="x86_64"
    ;;
    "X64")
    ARCH_NAME="x86_64"
    ;;
    "ARM")
    ARCH_NAME="arm"
    ;;
    "ARM64")
    ARCH_NAME="arm64"
    ;;
esac

case "$RUNNER_OS" in
    "Linux")
    OS_NAME="linux"
    ;;
    "macOS")
    OS_NAME="darwin"
    ;;
esac

TARGET_URL="https://github.com/thomas-maurice/protoc-gen-go-tmprl/releases/download/${VERSION}/protoc-gen-go-tmprl_${OS_NAME}_${ARCH_NAME}.tar.gz"
echo "Downloading ${TARGET_URL}"
wget "${TARGET_URL}" -qO /tmp/protoc-gen-go-tmprl.tar.gz
tar zxvf /tmp/protoc-gen-go-tmprl.tar.gz -C /bin protoc-gen-go-tmprl

if [ -f /bin/protoc-gen-go-tmprl ]; then
    echo "Installed protoc-gen-go-tmprl to /bin"
fi;
