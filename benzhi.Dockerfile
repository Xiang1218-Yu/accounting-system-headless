# 官方 Go 镜像，自带完整工具链
FROM golang:1.26

WORKDIR /app

# 先复制依赖文件并下载依赖，保证容器内离线可用
COPY go.mod go.sum ./
RUN go mod download

# 复制所有项目文件
COPY . .

# 预编译，确认基础代码健康
RUN go build ./...

# 容器启动后进入 shell，方便模型操作
CMD ["bash"]
