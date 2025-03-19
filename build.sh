#!/bin/bash
set -e

# 获取当前目录路径
CURRENT_DIR=$(pwd)

# 输出目录
PROTO_DIR="examples/greeter/proto"
GO_OUT_DIR="${PROTO_DIR}"
THOR_OUT_DIR="${PROTO_DIR}"

# 编译 protoc-gen-thor 插件
echo "正在编译 protoc-gen-thor 插件..."
mkdir -p cmd/protoc-gen-thor
go build -o cmd/protoc-gen-thor/protoc-gen-thor cmd/protoc-gen-thor/main.go

# 确保插件是可执行的
chmod +x cmd/protoc-gen-thor/protoc-gen-thor

# 将插件目录添加到 PATH
export PATH=$PATH:${CURRENT_DIR}/cmd/protoc-gen-thor

# 检查 protoc 是否已安装
if ! command -v protoc &> /dev/null; then
    echo "错误: 未找到 protoc，请安装 Protocol Buffers 编译器"
    exit 1
fi

# 检查 protoc-gen-go 是否已安装
if ! command -v protoc-gen-go &> /dev/null; then
    echo "正在安装 protoc-gen-go..."
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
fi

echo "正在生成代码..."

# 确保输出目录存在
mkdir -p ${GO_OUT_DIR}
mkdir -p ${THOR_OUT_DIR}

# 运行 protoc 命令生成代码
protoc --proto_path=${PROTO_DIR} \
       --go_out=${GO_OUT_DIR} \
       --go_opt=paths=source_relative \
       --thor_out=${THOR_OUT_DIR} \
       --thor_opt=paths=source_relative \
       ${PROTO_DIR}/greeter.proto

echo "代码生成完成!"

# 为了调试，显示生成的文件
echo "生成的文件:"
ls -la ${GO_OUT_DIR}/greeter.pb.go ${THOR_OUT_DIR}/greeter.thor.go 2>/dev/null || echo "文件未找到，请检查错误"

echo "编译完成"