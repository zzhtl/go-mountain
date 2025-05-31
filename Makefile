.PHONY: run build clean create-admin frontend-admin frontend-mp help

# 默认目标
help:
	@echo "可用的命令:"
	@echo "  run              - 运行服务器"
	@echo "  build            - 构建项目"
	@echo "  create-admin     - 创建初始管理员用户"
	@echo "  frontend-admin   - 构建管理后台前端"
	@echo "  frontend-mp      - 构建小程序前端"
	@echo "  clean            - 清理构建文件"

# 运行服务器
run:
	@echo "启动服务器..."
	go run cmd/app/main.go

# 构建项目
build:
	@echo "构建项目..."
	go build -o bin/go-mountain cmd/app/main.go

# 创建初始管理员用户
create-admin:
	@echo "创建初始管理员用户..."
	go run cmd/create_admin/main.go

# 构建管理后台前端
frontend-admin:
	@echo "构建管理后台前端..."
	cd frontend-admin && npm run build

# 构建小程序前端（如果需要的话）
frontend-mp:
	@echo "小程序前端不需要构建"

# 清理构建文件
clean:
	@echo "清理构建文件..."
	rm -rf bin/
	rm -rf frontend-admin/dist/

# 开发环境启动
dev:
	@echo "启动开发环境..."
	go run cmd/app/main.go

# 安装依赖
install:
	@echo "安装Go依赖..."
	go mod tidy
	@echo "安装前端依赖..."
	cd frontend-admin && npm install 