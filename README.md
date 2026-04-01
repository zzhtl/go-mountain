# Go Mountain - 小程序快速开发平台

Go Mountain 是一套基于 Go + Vue 3 的小程序快速开发平台，提供完整的后台管理系统、RBAC 权限控制、微信小程序登录/支付、内容管理、活动报名系统以及代码生成器，帮助开发者快速构建和交付小程序项目。

## 技术栈

| 层级 | 技术选型 |
| --- | --- |
| 后端框架 | Go + Gin |
| ORM | GORM v2 |
| 数据库 | PostgreSQL（推荐）/ SQLite |
| 认证 | JWT + bcrypt |
| 权限 | RBAC（角色-菜单-按钮三级） |
| 微信支付 | PowerWeChat v3（JSAPI） |
| 前端框架 | Vue 3 + Composition API |
| UI 组件库 | Element Plus |
| 状态管理 | Pinia |
| 富文本 | TinyMCE 8 |
| 构建工具 | Vite 6 |
| 配置管理 | Viper（YAML + 环境变量） |

## 项目结构

```
go-mountain/
├── cmd/
│   ├── server/main.go          # 应用入口
│   └── migrate/main.go         # 数据库迁移工具
├── configs/
│   └── config.yaml             # 配置文件
├── internal/
│   ├── config/                 # 配置加载
│   ├── db/                     # 数据库初始化
│   ├── server/                 # HTTP 服务器（优雅关闭）
│   ├── router/                 # 路由定义
│   ├── middleware/             # 中间件（CORS、JWT、RBAC）
│   ├── handler/                # HTTP 处理器
│   ├── service/                # 业务逻辑层
│   ├── model/                  # 数据模型
│   ├── repository/             # 泛型 Repository（BaseRepo[T]）
│   └── pkg/                    # 公共工具（response、crypto、errcode）
├── frontend-admin/             # Vue 3 管理后台
│   ├── src/
│   │   ├── api/                # API 请求封装
│   │   ├── router/             # 动态路由
│   │   ├── store/              # Pinia 状态管理
│   │   ├── views/              # 页面组件
│   │   ├── components/         # 公共组件（RichEditor 等）
│   │   └── directives/         # 自定义指令（v-permission）
│   ├── public/tinymce/         # TinyMCE 自托管资源
│   └── package.json
├── frontend-mp/                # 微信小程序（uni-app）
├── uploads/                    # 上传文件目录
├── Makefile                    # 构建命令
└── go.mod
```

## 环境要求

- Go 1.21+
- Node.js 18+
- PostgreSQL 14+（生产环境）或 SQLite（开发/测试）

## 快速开始

### 1. 克隆项目

```bash
git clone https://github.com/zzhtl/go-mountain.git
cd go-mountain
```

### 2. 安装依赖

```bash
make install
# 等同于：
# go mod tidy
# cd frontend-admin && npm install
```

### 3. 配置

编辑 `configs/config.yaml`：

```yaml
server:
  port: 8080

database:
  # 开发环境使用 SQLite
  driver: sqlite3
  dsn: data.db

  # 生产环境推荐 PostgreSQL
  # driver: postgres
  # dsn: "host=localhost user=postgres password=yourpassword dbname=go_mountain port=5432 sslmode=disable"

jwt:
  secret: your_jwt_secret_at_least_32_chars

wechat:
  app_id: wx_your_appid
  secret: your_app_secret
```

也可以通过环境变量覆盖配置：

```bash
export DATABASE_DRIVER=postgres
export DATABASE_DSN="host=localhost user=postgres password=xxx dbname=go_mountain port=5432 sslmode=disable"
export JWT_SECRET="your-secret"
```

### 4. 数据库迁移

```bash
make migrate
```

此命令会：
- 自动创建所有数据表
- 初始化默认角色（管理员、编辑员、查看者）
- 初始化默认菜单结构
- 创建 admin 账号并输出随机密码

```
========================================
管理员账号创建成功
用户名: admin
密码: xK9mP2qR7w
请妥善保管密码！
========================================
```

### 5. 构建前端

```bash
make frontend-admin
```

### 6. 启动服务

```bash
# 开发模式
make run

# 或构建生产二进制
make build
./bin/go-mountain
```

访问管理后台：http://localhost:8080/web/

## 核心功能

### RBAC 权限管理

- **三级权限模型**：目录 → 菜单 → 按钮/API
- **动态菜单**：前端根据用户角色动态生成侧边栏和路由
- **按钮级权限**：`v-permission` 指令控制按钮显隐
- **API 级鉴权**：RBAC 中间件自动校验请求路径和方法

### 内容管理

- **文章管理**：TinyMCE 富文本编辑器，支持图片/视频上传、中文界面
- **栏目管理**：文章分类，支持排序

### 活动报名系统

- **活动管理**：创建活动，设置报名时间窗口、人数上限、费用
- **活动状态流转**：草稿 → 报名中 → 报名截止 → 进行中 → 已结束
- **报名管理**：报名校验（状态、时间窗口、人数、去重），免费活动自动确认
- **支付管理**：微信 JSAPI 支付，回调处理，退款

### 微信支付集成

通过 [PowerWeChat v3](https://github.com/ArtisanCloud/PowerWeChat) 实现：

- JSAPI 统一下单
- 支付回调签名验证 + AES-256-GCM 解密
- 退款接口
- **所有支付配置通过后台管理界面维护**（系统配置 → 微信支付分组），无需修改配置文件

### 代码生成器

核心商业化功能，根据数据库表自动生成管理后台 CRUD 代码：

1. **选择数据表**：自动读取 PostgreSQL 表结构（列名、类型、注释）
2. **配置字段**：设置每个字段的显示名称、表单类型（文本框/数字/下拉框/日期/图片上传/富文本/开关等）、是否列表显示、可搜索、必填
3. **预览代码**：预览将要生成的 Go Model、Service、Handler 代码
4. **一键生成**：
   - 生成 `internal/model/gen_xxx.go`
   - 生成 `internal/service/gen_xxx_svc.go`
   - 生成 `internal/handler/gen_xxx_handler.go`
   - 自动在数据库中创建菜单项（含 CRUD 权限按钮）并分配给管理员角色
   - 输出 Router 和 API 代码片段，手动粘贴到对应文件后重新编译即可

5. **前端自动适配**：生成的模块通过 `DynamicCrud` 通用组件自动渲染，无需编写前端页面

### 系统配置

- 数据库驱动的 Key-Value 配置存储，带内存缓存
- 按分组管理配置项
- 支持 string/number/boolean/json 类型
- 敏感配置（密钥等）在界面上以密码框显示

## API 接口

### 公开接口

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/ping` | 健康检查 |
| POST | `/api/mp/login` | 小程序微信登录 |
| POST | `/api/mp/register` | 小程序用户注册 |
| GET | `/api/mp/columns/` | 栏目列表 |
| GET | `/api/mp/articles/column/:columnId` | 栏目文章 |
| GET | `/api/mp/articles/:id` | 文章详情 |
| GET | `/api/mp/activities/` | 活动列表 |
| GET | `/api/mp/activities/:id` | 活动详情 |
| POST | `/api/payment/wechat/notify` | 微信支付回调 |

### 小程序认证接口（需 JWT）

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/api/mp/registrations` | 报名 |
| PUT | `/api/mp/registrations/:id/cancel` | 取消报名 |
| GET | `/api/mp/registrations/mine` | 我的报名 |
| POST | `/api/mp/payments/create` | 创建支付订单 |
| GET | `/api/mp/payments/query` | 查询支付状态 |

### 管理后台接口（需 JWT + RBAC）

| 模块 | 路由前缀 | 支持操作 |
| --- | --- | --- |
| 后台用户 | `/api/admin/backend-users` | CRUD + 状态 + 重置密码 |
| 小程序用户 | `/api/admin/users` | 列表 + 详情 + 更新 + 删除 |
| 文章 | `/api/admin/articles` | CRUD + 状态 |
| 栏目 | `/api/admin/columns` | CRUD |
| 角色 | `/api/admin/roles` | CRUD + 状态 + 菜单分配 |
| 菜单 | `/api/admin/menus` | CRUD + 树形结构 |
| 活动 | `/api/admin/activities` | CRUD + 状态 |
| 报名 | `/api/admin/registrations` | 列表 + 详情 |
| 支付 | `/api/admin/payments` | 列表 + 详情 + 退款 |
| 系统配置 | `/api/admin/system-configs` | 列表 + 分组 + 保存 + 批量保存 + 删除 |
| 代码生成 | `/api/admin/codegen` | 配置 CRUD + 表/列查询 + 预览 + 生成 |
| 文件上传 | `/api/admin/upload` | 图片 + 视频 |

## 生产部署

### 方式一：二进制部署

```bash
# 构建
make build
make frontend-admin

# 部署目录结构
deploy/
├── bin/go-mountain          # 后端二进制
├── configs/config.yaml      # 配置文件
├── frontend-admin/dist/     # 前端构建产物
└── uploads/                 # 上传文件目录

# 启动
cd deploy && ./bin/go-mountain
```

### 方式二：systemd 服务

创建 `/etc/systemd/system/go-mountain.service`：

```ini
[Unit]
Description=Go Mountain Server
After=network.target postgresql.service

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/go-mountain
ExecStart=/opt/go-mountain/bin/go-mountain
Restart=always
RestartSec=5
Environment=DATABASE_DRIVER=postgres
Environment=DATABASE_DSN=host=localhost user=postgres password=xxx dbname=go_mountain port=5432 sslmode=disable
Environment=JWT_SECRET=your-production-secret

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable go-mountain
sudo systemctl start go-mountain
```

### 方式三：Docker 部署

创建 `Dockerfile`：

```dockerfile
# 构建后端
FROM golang:1.21-alpine AS backend
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o bin/go-mountain cmd/server/main.go

# 构建前端
FROM node:18-alpine AS frontend
WORKDIR /app/frontend-admin
COPY frontend-admin/package*.json ./
RUN npm ci
COPY frontend-admin/ .
RUN npm run build

# 最终镜像
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai
WORKDIR /app
COPY --from=backend /app/bin/go-mountain .
COPY --from=frontend /app/frontend-admin/dist ./frontend-admin/dist
COPY configs/ ./configs/
RUN mkdir -p uploads
EXPOSE 8080
CMD ["./go-mountain"]
```

```bash
docker build -t go-mountain .
docker run -d -p 8080:8080 \
  -e DATABASE_DRIVER=postgres \
  -e DATABASE_DSN="host=db user=postgres password=xxx dbname=go_mountain port=5432 sslmode=disable" \
  -e JWT_SECRET="your-secret" \
  -v /data/uploads:/app/uploads \
  go-mountain
```

### Nginx 反向代理

```nginx
server {
    listen 80;
    server_name your-domain.com;

    client_max_body_size 50m;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## Makefile 命令

| 命令 | 说明 |
| --- | --- |
| `make help` | 显示可用命令 |
| `make install` | 安装所有依赖（Go + npm） |
| `make migrate` | 数据库迁移 + 初始化默认数据 + 创建管理员 |
| `make run` | 开发模式运行 |
| `make build` | 构建生产二进制到 `bin/go-mountain` |
| `make frontend-admin` | 构建管理后台前端 |
| `make clean` | 清理构建产物 |

## 代码生成器使用指南

以创建一个「商品管理」模块为例：

### 1. 创建数据库表

```sql
CREATE TABLE products (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    price DECIMAL(10,2) NOT NULL DEFAULT 0,
    thumbnail TEXT,
    category TEXT,
    status INT DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

COMMENT ON TABLE products IS '商品表';
COMMENT ON COLUMN products.name IS '商品名称';
COMMENT ON COLUMN products.description IS '商品描述';
COMMENT ON COLUMN products.price IS '价格';
COMMENT ON COLUMN products.thumbnail IS '缩略图';
COMMENT ON COLUMN products.category IS '分类';
COMMENT ON COLUMN products.status IS '状态(1:上架 0:下架)';
```

### 2. 在管理后台配置

1. 进入 **代码生成** → **新建配置**
2. 选择 `products` 表，系统自动读取列信息
3. 填写模块名 `product`，显示名称 `商品管理`
4. 配置每个字段的表单类型（如 thumbnail 选择「图片上传」，status 选择「下拉框」并添加选项）
5. 保存配置

### 3. 预览并生成

1. 点击 **预览** 查看将生成的代码
2. 点击 **生成代码**
3. 系统自动生成 3 个 Go 文件并创建菜单

### 4. 集成到项目

将弹窗中显示的 **Router 代码片段** 添加到 `internal/router/router.go`，然后重新编译：

```bash
go build -o bin/go-mountain cmd/server/main.go
```

重启服务后，刷新管理后台，侧边栏自动出现「商品管理」菜单，CRUD 功能即刻可用。

## License

MIT
