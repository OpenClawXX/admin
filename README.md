# 运维人员管理平台

企业 IT 部门运维人员管理平台，用于管理工单流转、工程师绩效、项目进度、团队协作等。

## 技术栈

| 层级 | 技术 |
|------|------|
| 前端 | Vue 3 + Vite + Arco Design Vue + ECharts + Pinia |
| 后端 | Go + Gin + sqlx |
| 数据库 | PostgreSQL |
| 缓存 | Redis |
| 文件存储 | 本地存储（可扩展 MinIO） |
| 认证 | JWT (golang-jwt) |

## 功能特性

### 用户与权限
- JWT 登录认证
- RBAC 三级角色：管理员 / 主管 / 工程师
- 用户 CRUD、重置密码（手动或自动生成强密码）
- 团队管理（主管自动归属、唯一主管约束）
- 菜单按角色动态显示/隐藏

### 工单管理
- 全生命周期：创建 → 派单 → 处理 → 完单 → 验收 → 归档
- 工单类型：故障、实施、巡检
- 优先级：紧急、重大、严重、普通
- 派单/转派（下拉选择本团队工程师）
- 挂起/恢复、进度上报、流转日志
- 工单与项目绑定

### 完单报告
- 提交表单：解决方案、根因分析、处理结果、影响范围、遗留问题、后续建议、交接备注
- 文件上传（自动上传、拖拽、类型白名单、大小限制）
- 文件权限控制（仅本人/管理员/主管可删除）
- 驳回后自动填充上次提交内容和已有文件

### 项目管理
- 项目信息：自动编号、类型、优先级、需求方、负责人、成员、预算、日期
- 项目详情抽屉：展示完整信息和关联工单
- 工单列表按项目筛选

### 数据权限
- 管理员：查看/操作所有数据
- 主管：仅查看本团队成员相关工单，指派限本团队
- 工程师：仅查看自己创建/被指派的工单，创建时自动指派给自己

### 监控集成（预留）
- Zabbix / Prometheus Webhook 接口

## 项目结构

```
admin/
├── cmd/server/              # 后端入口
│   └── main.go
├── config/                  # 配置加载
│   └── config.go
├── internal/
│   ├── handler/             # HTTP 处理器
│   │   ├── user.go
│   │   ├── team.go
│   │   ├── ticket.go
│   │   ├── project.go
│   │   └── completion.go
│   ├── service/             # 业务逻辑
│   │   ├── user.go
│   │   ├── team.go
│   │   ├── ticket.go
│   │   ├── project.go
│   │   └── completion.go
│   ├── repository/          # 数据访问层
│   │   ├── user.go
│   │   ├── team.go
│   │   ├── ticket.go
│   │   ├── project.go
│   │   └── completion.go
│   ├── model/               # 数据模型
│   ├── middleware/           # JWT、RBAC、CORS、日志
│   ├── pkg/auth/            # JWT + 密码工具
│   ├── pkg/response/        # 统一响应格式
│   └── database/            # 数据库连接
├── migrations/              # 数据库迁移脚本
├── uploads/                 # 上传文件存储
├── web/                     # 前端项目
│   ├── src/
│   │   ├── api/             # API 接口封装
│   │   ├── views/           # 页面组件
│   │   ├── components/      # 通用组件
│   │   ├── stores/          # Pinia 状态管理
│   │   ├── router/          # 路由配置
│   │   └── utils/           # 工具函数
│   ├── package.json
│   └── vite.config.ts
├── .env.example             # 环境变量模板
├── go.mod
├── Makefile
└── update.md                # 更新日志
```

## 快速开始

### 环境要求

- Go 1.22+
- Node.js 18+
- PostgreSQL 14+

### 1. 克隆项目

```bash
git clone https://github.com/Mcloud136/admin.git
cd admin
```

### 2. 配置数据库

```bash
# 创建数据库
psql -U postgres -c "CREATE DATABASE ops_platform;"

# 执行迁移
psql -U postgres -d ops_platform -f migrations/001_init.sql
psql -U postgres -d ops_platform -f migrations/002_project_enhance.sql
```

### 3. 启动后端

```bash
cp .env.example .env
# 编辑 .env 修改数据库连接信息

go mod tidy
go run cmd/server/main.go
```

后端启动于 http://localhost:8080

### 4. 启动前端

```bash
cd web
npm install
npm run dev
```

前端启动于 http://localhost:3000，API 请求自动代理到后端。

### 5. 登录

| 用户名 | 密码 | 角色 |
|--------|------|------|
| admin | admin123 | 管理员 |
| supervisor1 | admin123 | 主管 |
| engineer1 | admin123 | 工程师 |

## API 接口

<details>
<summary>点击展开完整 API 列表</summary>

### 认证
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/login | 登录 |
| GET | /api/profile | 个人信息 |

### 用户管理
| 方法 | 路径 | 说明 | 权限 |
|------|------|------|------|
| GET | /api/users | 用户列表 | 登录 |
| POST | /api/users | 创建用户 | 管理员 |
| PUT | /api/users/:id | 编辑用户 | 管理员/主管 |
| DELETE | /api/users/:id | 删除用户 | 管理员 |
| POST | /api/users/:id/reset-password | 重置密码 | 管理员 |

### 团队管理
| 方法 | 路径 | 说明 | 权限 |
|------|------|------|------|
| GET | /api/teams | 团队列表 | 登录 |
| POST | /api/teams | 创建团队 | 管理员/主管 |
| PUT | /api/teams/:id | 编辑团队 | 管理员/主管 |
| DELETE | /api/teams/:id | 删除团队 | 管理员/主管 |

### 项目管理
| 方法 | 路径 | 说明 | 权限 |
|------|------|------|------|
| GET | /api/projects | 项目列表 | 登录 |
| GET | /api/projects/:id | 项目详情 | 登录 |
| POST | /api/projects | 创建项目 | 管理员/主管 |
| PUT | /api/projects/:id | 编辑项目 | 管理员/主管 |
| DELETE | /api/projects/:id | 删除项目 | 管理员/主管 |

### 工单管理
| 方法 | 路径 | 说明 | 权限 |
|------|------|------|------|
| GET | /api/tickets | 工单列表（支持 status/priority/type/project_id/keyword 筛选） | 登录 |
| GET | /api/tickets/:id | 工单详情 | 登录 |
| POST | /api/tickets | 创建工单 | 登录 |
| PUT | /api/tickets/:id | 编辑工单 | 登录 |
| DELETE | /api/tickets/:id | 删除工单 | 管理员 |
| POST | /api/tickets/:id/assign | 派单 | 管理员/主管 |
| POST | /api/tickets/:id/transfer | 转派 | 管理员/主管 |
| POST | /api/tickets/:id/suspend | 挂起 | 登录 |
| POST | /api/tickets/:id/resume | 恢复 | 登录 |
| POST | /api/tickets/:id/progress | 进度上报 | 登录 |
| POST | /api/tickets/:id/logs | 添加日志 | 登录 |
| GET | /api/tickets/:id/logs | 流转日志 | 登录 |
| POST | /api/tickets/:id/complete | 完单 | 登录 |
| POST | /api/tickets/:id/review | 验收 | 管理员/主管 |
| POST | /api/tickets/:id/archive | 归档 | 管理员/主管 |

### 完单报告与附件
| 方法 | 路径 | 说明 | 权限 |
|------|------|------|------|
| POST | /api/tickets/:id/completion | 提交/更新完单报告 | 登录 |
| GET | /api/tickets/:id/completion | 获取完单报告 | 登录 |
| POST | /api/tickets/:id/files | 上传附件 | 登录 |
| GET | /api/tickets/:id/files | 附件列表 | 登录 |
| GET | /api/tickets/:id/files/:file_id/download | 下载附件 | 登录 |
| DELETE | /api/tickets/:id/files/:file_id | 删除附件 | 上传者/管理员/主管 |

</details>

## 工单状态流转

```
创建 → 待派发 → 已派发 → 处理中 → 待验收 → 已完单 → 已归档
                ↓          ↓          ↓
             挂起 ←→    挂起中      驳回(→处理中)
```

## 安全措施

- JWT 认证 + RBAC 权限控制
- 文件上传类型白名单（文档/图片/日志/压缩包/代码）
- 文件大小限制（50MB）
- 路径穿越防护
- 文件删除权限校验
- 密码 bcrypt 加密存储
- CORS 跨域控制

## 更新日志

详见 [update.md](update.md)

## License

MIT
