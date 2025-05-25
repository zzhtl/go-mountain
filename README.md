# 远山公益项目

远山公益是一个综合公益平台，包含管理后台和微信小程序，用于发布公益文章和管理公益活动。

## 项目架构

- **后端**：基于Go + Gin的API服务
- **管理后台**：基于Vue 3 + Element Plus的后台管理系统
- **小程序前端**：微信小程序客户端

## 环境要求

- Go 1.16+
- Node.js 14+
- npm 6+
- 微信开发者工具

## 安装与运行

### 后端

1. 进入项目根目录，编译并运行后端：

```bash
# 开发环境运行
go run cmd/app/main.go

# 或构建可执行文件
go build -o go-mountain cmd/app/main.go
./go-mountain
```

2. 配置文件位于 `configs/config.yaml`，支持环境变量覆盖：

```bash
export SERVER_PORT=8080
export DATABASE_DRIVER=sqlite3
export DATABASE_DSN="./data.db"
export JWT_SECRET="your-secret-key"
```

### 管理后台前端

1. 进入管理后台项目目录并安装依赖：

```bash
cd frontend-admin
npm install
```

2. 开发环境启动：

```bash
npm run dev
```

3. 构建生产环境版本：

```bash
npm run build
```

4. 访问地址：
   - 开发环境：http://localhost:5173 (或Vite输出的端口)
   - 生产环境：http://your-server:8080/web/

### 小程序前端

1. 进入小程序项目目录：

```bash
cd frontend-mp
```

2. 使用微信开发者工具打开此目录，编译运行小程序

## 主要功能

### 管理后台

- **用户管理**：查看、编辑、删除用户信息
- **栏目管理**：创建、编辑、删除文章栏目
- **文章管理**：
  - 富文本编辑器支持图片和视频上传
  - 文章发布、编辑、下架、删除
  - 按栏目和状态筛选文章

### 小程序端

- **用户登录**：微信授权登录
- **文章浏览**：按栏目查看文章列表
- **文章详情**：阅读完整文章内容

## API接口说明

### 认证相关

- `POST /api/admin/auth/login` - 管理员登录，返回JWT令牌
- `POST /api/mp/login` - 微信小程序登录，获取openid
- `POST /api/mp/register` - 小程序用户注册绑定手机号

### 用户管理

- `GET /api/admin/users` - 获取用户列表
- `GET /api/admin/users/:id` - 获取用户详情
- `PUT /api/admin/users/:id` - 更新用户信息
- `DELETE /api/admin/users/:id` - 删除用户

### 栏目管理

- `GET /api/admin/columns` - 获取栏目列表
- `POST /api/admin/columns` - 创建新栏目
- `PUT /api/admin/columns/:id` - 更新栏目信息
- `DELETE /api/admin/columns/:id` - 删除栏目

### 文章管理

- `GET /api/admin/articles` - 获取文章列表
- `POST /api/admin/articles` - 创建新文章
- `GET /api/admin/articles/:id` - 获取文章详情
- `PUT /api/admin/articles/:id` - 更新文章
- `PUT /api/admin/articles/:id/status` - 更新文章状态
- `DELETE /api/admin/articles/:id` - 删除文章

### 小程序API

- `GET /api/mp/columns` - 获取栏目列表
- `GET /api/mp/articles/column/:columnId` - 获取指定栏目的文章
- `GET /api/mp/articles/:id` - 获取文章详情

### 文件上传

- `POST /api/admin/upload/image` - 上传图片
- `POST /api/admin/upload/video` - 上传视频

## 技术栈

### 后端

- Go语言
- Gin Web框架
- SQLite/MySQL数据库
- JWT认证

### 管理后台

- Vue 3
- Vue Router
- Element Plus
- Axios
- Quill富文本编辑器

### 小程序端

- 微信小程序原生开发

## 未来规划

1. 用户权限与角色管理系统
2. 活动、捐赠等更多业务模块
3. 集成微信支付实现在线捐赠
4. 数据统计与可视化报表
5. 性能优化与测试完善
