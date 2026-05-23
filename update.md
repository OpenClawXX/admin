# 运维人员管理平台 - 更新日志

## 项目概述

企业 IT 部门运维人员管理平台，用于管理工单流转、工程师绩效、项目进度等。

**技术栈：** Vue 3 + Arco Design Vue (前端) | Go + Gin + sqlx (后端) | PostgreSQL (数据库)

---

## 功能模块

### 1. 用户与权限管理
- JWT 登录认证
- RBAC 三级角色：管理员、主管、工程师
- 用户 CRUD（创建/编辑/删除/重置密码）
- 团队管理（创建/编辑/删除，主管自动归属团队）
- 主管变更时自动清空原团队主管
- 重置密码支持手动输入或自动生成 8 位强密码
- 编辑用户时状态开关（正常/禁用）

### 2. 工单管理
- 工单全生命周期：创建 → 派单 → 处理 → 完单 → 验收 → 归档
- 工单类型：故障、实施、巡检
- 优先级：紧急、重大、严重、普通（带颜色标签）
- 派单/转派（下拉选择工程师）
- 挂起/恢复
- 进度上报
- 管理员可删除工单
- 工单与项目绑定

### 3. 权限控制
- 管理员：查看/操作所有工单
- 主管：仅查看本团队成员相关工单，指派限本团队
- 工程师：仅查看自己创建/被指派的工单，创建时自动指派给自己
- 菜单按角色动态显示/隐藏

### 4. 完单报告
- 提交表单：解决方案、根因分析、处理结果、影响范围、遗留问题、后续建议、交接备注
- 文件上传（自动上传，支持拖拽）
- 文件类型白名单（文档/图片/日志/压缩包/代码）
- 文件大小限制 50MB
- 路径穿越防护
- 已上传文件删除权限（仅本人/管理员/主管）
- 驳回后重新提交自动填充上次内容和已有文件
- UPSERT 处理重复提交

### 5. 项目管理
- 项目信息：编号（自动生成）、名称、类型、优先级、需求方、负责人、项目成员、预算、描述、备注、日期
- 项目详情抽屉：展示完整信息和关联工单列表
- 项目成员管理（多选）
- 工单列表按项目筛选

### 6. 流转日志
- 工单操作全记录
- 显示操作人姓名和时间

### 7. 监控集成（后端预留）
- Zabbix / Prometheus Webhook 接口预留

---

## 安全措施

| 措施 | 说明 |
|------|------|
| JWT 认证 | 所有 API 需携带 Bearer Token |
| RBAC 权限 | 基于角色的接口和数据权限控制 |
| 文件类型白名单 | 仅允许常见文档/图片/日志/代码格式 |
| 文件大小限制 | 单文件最大 50MB |
| 路径穿越防护 | `filepath.Base()` 清洗文件名 |
| 文件删除权限 | 仅上传者/管理员/主管可删除 |
| 工单删除联动 | 删除工单时自动清理关联文件和磁盘数据 |
| CORS 配置 | 跨域请求控制 |
| 密码加密 | bcrypt 哈希存储 |

---

## 项目结构

```
ops-platform/                  # 后端
├── cmd/server/main.go         # 入口
├── config/                    # 配置
├── internal/
│   ├── handler/               # HTTP 处理器
│   ├── service/               # 业务逻辑
│   ├── repository/            # 数据访问
│   ├── model/                 # 数据模型
│   ├── middleware/             # JWT、RBAC、CORS、日志
│   ├── pkg/auth/              # JWT + 密码工具
│   ├── pkg/response/          # 统一响应
│   └── database/              # 数据库连接
├── migrations/                # 数据库迁移
└── uploads/                   # 上传文件存储

ops-platform-web/              # 前端
├── src/
│   ├── api/                   # API 接口
│   ├── views/                 # 页面
│   │   ├── login/             # 登录
│   │   ├── dashboard/         # 工作台
│   │   ├── ticket/            # 工单管理
│   │   ├── project/           # 项目管理
│   │   ├── engineer/          # 工程师管理
│   │   ├── team/              # 团队管理
│   │   ├── knowledge/         # 知识库
│   │   ├── schedule/          # 排班管理
│   │   ├── asset/             # 资产管理
│   │   └── system/            # 系统设置
│   ├── components/            # 通用组件（Layout）
│   ├── stores/                # Pinia 状态管理
│   ├── router/                # 路由配置
│   └── utils/                 # 工具函数（Axios 封装）
```

---

## 启动方式

### 后端
```bash
cd ops-platform
cp .env.example .env          # 配置数据库连接
go mod tidy
go run cmd/server/main.go     # 启动 :8080
```

### 前端
```bash
cd ops-platform-web
npm install
npm run dev                   # 启动 :3000
```

### 数据库
```bash
# 创建数据库
psql -U postgres -c "CREATE DATABASE ops_platform;"

# 执行迁移
psql -U postgres -d ops_platform -f migrations/001_init.sql
psql -U postgres -d ops_platform -f migrations/002_project_enhance.sql
```

### 测试账号（密码统一 admin123）
| 用户名 | 角色 |
|--------|------|
| admin | 管理员 |
| supervisor1 | 主管 |
| engineer1 | 工程师 |

---

## API 接口列表

| 方法 | 路径 | 说明 | 权限 |
|------|------|------|------|
| POST | /api/login | 登录 | 公开 |
| GET | /api/profile | 个人信息 | 登录 |
| GET | /api/users | 用户列表 | 登录 |
| POST | /api/users | 创建用户 | 管理员 |
| PUT | /api/users/:id | 编辑用户 | 管理员/主管 |
| DELETE | /api/users/:id | 删除用户 | 管理员 |
| POST | /api/users/:id/reset-password | 重置密码 | 管理员 |
| GET | /api/teams | 团队列表 | 登录 |
| POST | /api/teams | 创建团队 | 管理员/主管 |
| PUT | /api/teams/:id | 编辑团队 | 管理员/主管 |
| DELETE | /api/teams/:id | 删除团队 | 管理员/主管 |
| GET | /api/projects | 项目列表 | 登录 |
| GET | /api/projects/:id | 项目详情 | 登录 |
| POST | /api/projects | 创建项目 | 管理员/主管 |
| PUT | /api/projects/:id | 编辑项目 | 管理员/主管 |
| DELETE | /api/projects/:id | 删除项目 | 管理员/主管 |
| GET | /api/tickets | 工单列表 | 登录（按角色过滤） |
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
| POST | /api/tickets/:id/completion | 提交完单报告 | 登录 |
| GET | /api/tickets/:id/completion | 获取完单报告 | 登录 |
| POST | /api/tickets/:id/files | 上传附件 | 登录 |
| GET | /api/tickets/:id/files | 附件列表 | 登录 |
| GET | /api/tickets/:id/files/:file_id/download | 下载附件 | 登录 |
| DELETE | /api/tickets/:id/files/:file_id | 删除附件 | 上传者/管理员/主管 |
