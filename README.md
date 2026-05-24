# 运维人员管理平台

企业 IT 部门运维人员管理平台，用于管理工单流转、工程师绩效、项目进度、团队协作等。

## 技术栈

| 层级 | 技术 |
|------|------|
| 前端 | Vue 3 + Vite + Arco Design Vue + ECharts + Pinia + TipTap |
| 后端 | Go + Gin + sqlx |
| 数据库 | PostgreSQL 17 |
| 缓存 | Redis |
| 文件存储 | 本地存储（可扩展 MinIO） |
| 认证 | JWT (golang-jwt) |
| 图标 | Lucide Icons |

## 功能特性

### 用户与权限
- JWT 登录认证
- RBAC 三级角色：管理员 / 主管 / 工程师
- 用户 CRUD、重置密码（手动或自动生成 8 位强密码）
- 团队管理（主管自动归属、唯一主管约束）
- 菜单按角色动态显示/隐藏

### 工单管理
- 全生命周期：创建 → 派单 → 处理 → 完单 → 验收 → 归档
- 工单类型：故障、实施、巡检
- 优先级：紧急、重大、严重、普通（带颜色标签）
- 派单/转派（下拉选择本团队工程师）
- 挂起/恢复、进度上报、流转日志
- 工单与项目绑定
- 完单报告（解决方案、根因分析、处理结果、影响范围、遗留问题、后续建议、交接备注）
- 文件上传（类型白名单、大小限制、路径穿越防护）
- 管理员可删除工单

### 项目管理
- 项目信息：自动编号、类型、优先级、需求方、负责人、成员、预算、日期
- 项目详情抽屉：展示完整信息和关联工单
- 项目成员管理（多选）
- 工单列表按项目筛选
- 完整状态流转：进行中 → 待验收 → 整改中 → 已完成

### 知识库
- TipTap 富文本编辑器（粗体/斜体/下划线/标题/列表/引用/代码块/图片/对齐）
- 上传文档自动解析（Word → HTML、Excel → 表格、文本 → 原文）
- 粘贴/拖拽图片（自动 base64 内联，支持缩放手柄）
- 文件预览（Word/Excel/图片/文本）
- 分类管理、搜索筛选

### 资产管理
- IT 资产登记（服务器/交换机/路由器/防火墙/存储/工作站）
- 资产字段：名称、类型、IP、状态、品牌、型号、序列号、位置、负责人、采购日期、保修到期
- 资产详情抽屉
- 删除资产自动解除工单关联

### 工作台
- 统计卡片（待处理/处理中/本月完单/总数）关联真实工单数据
- 工单趋势折线图（ECharts，近 14 天）
- 工单类型分布饼图（ECharts）
- 最近工单列表
- 响应式布局（自适应移动端）

### UI 设计
- 现代化白色侧边栏 + Lucide 图标
- 渐变统计卡片
- 圆角卡片、柔和阴影
- 深蓝渐变登录页
- 面包屑导航、用户头像+角色标签
- 移动端自动收起侧边栏

### 安全措施
- JWT 认证 + RBAC 权限控制
- 文件上传类型白名单（50MB 限制）
- 路径穿越防护
- 文件删除权限校验
- 密码 bcrypt 加密存储

## 项目结构

```
admin/
├── cmd/server/              # 后端入口
│   └── main.go
├── config/                  # 配置加载
├── internal/
│   ├── handler/             # HTTP 处理器
│   ├── service/             # 业务逻辑
│   ├── repository/          # 数据访问层
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
├── README.md
└── update.md                # 更新日志
```

## 快速开始

### 环境要求

- Go 1.22+
- Node.js 18+
- PostgreSQL 17+

### 1. 克隆项目

```bash
git clone https://github.com/Mcloud136/admin.git
cd admin
```

### 2. 配置数据库

```bash
psql -U postgres -c "CREATE DATABASE ops_platform;"
psql -U postgres -d ops_platform -f migrations/001_init.sql
psql -U postgres -d ops_platform -f migrations/002_project_enhance.sql
psql -U postgres -d ops_platform -f migrations/003_knowledge_base.sql
```

### 3. 启动后端

```bash
cp .env.example .env
# 编辑 .env 修改数据库连接信息
go mod tidy
go run cmd/server/main.go
```

### 4. 启动前端

```bash
cd web
npm install
npm run dev
```

### 5. 登录

| 用户名 | 密码 | 角色 |
|--------|------|------|
| admin | admin123 | 管理员 |
| supervisor1 | admin123 | 主管 |
| engineer1 | admin123 | 工程师 |

## 部署

生产环境部署请参考 [DEPLOY.md](DEPLOY.md)（宝塔面板部署指南）

## 更新日志

详见 [update.md](update.md)

## License

MIT
