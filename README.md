# go-mountain
远山公益小程序项目

## 使用说明

### 后端
1. 进入项目根目录，编译并运行后端：
   ```bash
   go run cmd/app/main.go
   go build -o go-mountain cmd/app/main.go
   ```
2. 配置文件位于 `configs/config.yaml`，支持环境变量覆盖：
   ```bash
   export SERVER_PORT=8081
   export DATABASE_DRIVER=postgres
   export DATABASE_DSN="user=... dbname=... sslmode=disable"
   ```

### 前端

#### 后台管理前端
1. 进入后台管理项目目录并安装依赖：
   ```bash
   cd frontend-admin
   npm install
   ```
2. 启动前端开发服务器：
   ```bash
   npm run dev
   ```
3. 通过浏览器访问 http://localhost:5173 （或 Vite 输出的端口）并操作后台管理界面

#### 小程序前端
1. 进入小程序项目目录：
   ```bash
   cd frontend-mp
   ```
2. 打开微信开发者工具，导入该目录并编译运行小程序

## 功能说明

- 后端 API：
  - GET  /api/ping                健康检查接口，返回 `pong`。
  - Items 模块：支持 `/api/items` 的增删改查。
  - 小程序用户：
    - POST /api/mp/login          微信 Code2Session 登录，获取 `openid`。
    - POST /api/mp/register       手机号绑定注册。
  - 后台管理：
    - POST /api/admin/auth/login  管理员登录，返回 JWT。
    - GET  /api/admin/users       列出所有用户。
    - GET  /api/admin/users/:id   查看单个用户。
    - PUT  /api/admin/users/:id   更新用户信息。
    - DELETE /api/admin/users/:id 删除用户。

- 前端界面：
  - `frontend-admin`：基于 Vue3 + Vite 的管理后台 SPA。
  - `frontend-mp`：微信小程序端，支持自动登录、手机号绑定、首页展示。

## 未来功能规划

1. 用户权限与角色管理：为不同角色分配不同操作权限。
2. 活动、捐赠、评论等更多业务模块插件化接入。
3. 集成微信支付/支付宝支付，实现在线捐赠与订单管理。
4. 后台报表与数据可视化，统计捐赠、用户活跃度等指标。
5. 多语言与国际化支持，拓展更多市场。
6. 性能优化：引入 pprof 调优、缓存层和消息队列提高吞吐。
7. 持续完善单元/集成测试，达到 90%+ 覆盖率。
