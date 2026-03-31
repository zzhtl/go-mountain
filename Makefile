.PHONY: run build clean migrate frontend-admin help dev install

# 默认目标
help:
	@echo "可用的命令:"
	@echo "  run              - 运行服务器"
	@echo "  build            - 构建项目"
	@echo "  migrate          - 数据库迁移（含创建管理员、初始化角色菜单）"
	@echo "  frontend-admin   - 构建管理后台前端"
	@echo "  clean            - 清理构建文件"
	@echo "  dev              - 开发环境启动"
	@echo "  install          - 安装所有依赖"

# 运行服务器
run:
	@echo "启动服务器..."
	go run cmd/server/main.go

# 构建项目
build:
	@echo "构建项目..."
	go build -o bin/go-mountain cmd/server/main.go

# 数据库迁移（创建表、初始化默认数据、创建管理员）
migrate:
	@echo "执行数据库迁移..."
	go run cmd/migrate/main.go

# 构建管理后台前端
frontend-admin:
	@echo "构建管理后台前端..."
	cd frontend-admin && npm run build

# 清理构建文件
clean:
	@echo "清理构建文件..."
	rm -rf bin/
	rm -rf frontend-admin/dist/

# 开发环境启动
dev:
	@echo "启动开发环境..."
	go run cmd/server/main.go

# 安装依赖
install:
	@echo "安装Go依赖..."
	go mod tidy
	@echo "安装前端依赖..."
	cd frontend-admin && npm install
