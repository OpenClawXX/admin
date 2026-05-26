# 宝塔面板部署指南

## 一、服务器要求

| 项目 | 最低配置 | 推荐配置 |
|------|---------|---------|
| 操作系统 | Ubuntu 20.04+ / Debian 11+ | Ubuntu 22.04 LTS |
| CPU | 2 核 | 4 核 |
| 内存 | 2 GB | 4 GB |
| 硬盘 | 40 GB SSD | 80 GB SSD |

---

## 二、安装宝塔面板

```bash
wget -O install.sh https://download.bt.cn/install/install-ubuntu_6.0.sh && sudo bash install.sh ed8484bec
```

安装完成后记录面板地址、用户名、密码，登录面板。

---

## 三、宝塔应用商店安装软件

登录宝塔面板 → **软件商店** → 搜索安装：

| 软件 | 用途 |
|------|------|
| **Nginx** | 反向代理 + 前端静态文件 |
| **PostgreSQL** | 数据库 |

---

## 四、部署后端

### 4.1 创建目录

```bash
mkdir -p /opt/ops-platform/uploads/branding
mkdir -p /opt/ops-platform/uploads/kb
```

### 4.2 上传文件

将 GitHub 仓库根目录的 `ops-server` 上传到 `/opt/ops-platform/`

```bash
chmod +x /opt/ops-platform/ops-server
```

### 4.3 配置环境变量

```bash
cat > /opt/ops-platform/.env << 'EOF'
SERVER_PORT=8080
GIN_MODE=release

DB_HOST=127.0.0.1
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=你的PostgreSQL密码
DB_NAME=ops_platform
DB_SSLMODE=disable

JWT_SECRET=替换为随机字符串
JWT_EXPIRE_HOUR=24
EOF
```

生成 JWT Secret：
```bash
openssl rand -base64 32
```

### 4.4 创建数据库

宝塔 PostgreSQL 管理 → SQL执行器，依次执行：

```
migrations/001_init.sql
migrations/002_project_enhance.sql
migrations/003_knowledge_base.sql
```

### 4.5 创建 systemd 服务

```bash
cat > /etc/systemd/system/ops-platform.service << 'EOF'
[Unit]
Description=Ops Platform Backend
After=network.target postgresql.service

[Service]
Type=simple
User=root
WorkingDirectory=/opt/ops-platform
ExecStart=/opt/ops-platform/ops-server
Restart=always
RestartSec=5
Environment=GIN_MODE=release

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable ops-platform
systemctl start ops-platform
systemctl status ops-platform
```

---

## 五、部署前端

### 5.1 上传前端文件

将 GitHub 仓库根目录的 `dist/` 文件夹内容上传到：

```
/www/wwwroot/ops-platform/
├── index.html
└── assets/
```

### 5.2 添加站点

宝塔面板 → **网站** → **添加站点**：
- 域名：你的域名或IP
- 根目录：`/www/wwwroot/ops-platform`
- PHP版本：纯静态

### 5.3 配置 Nginx

点击站点名 → **配置文件**，替换为：

```nginx
server {
    listen 80;
    server_name 你的域名或IP;

    root /www/wwwroot/ops-platform;
    index index.html;

    client_max_body_size 50m;

    # API 反向代理
    location /api/ {
        proxy_pass http://127.0.0.1:8080/api/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_read_timeout 300s;
    }

    # 上传文件访问
    location /uploads/ {
        proxy_pass http://127.0.0.1:8080/uploads/;
    }

    # Swagger 文档
    location /swagger/ {
        proxy_pass http://127.0.0.1:8080/swagger/;
    }

    # Vue Router
    location / {
        try_files $uri $uri/ /index.html;
    }

    # 静态资源缓存
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2)$ {
        expires 30d;
        add_header Cache-Control "public, immutable";
    }

    gzip on;
    gzip_types text/plain text/css application/json application/javascript text/xml;
    gzip_min_length 1024;
}
```

重载配置：`nginx -t && systemctl reload nginx`

---

## 六、防火墙

宝塔 → 安全 → 防火墙 → 放行 80 和 443

---

## 七、SSL（推荐）

站点 → SSL → Let's Encrypt → 申请 → 开启强制 HTTPS

---

## 八、验证

1. 访问 `http://你的域名` → 看到登录页
2. 用 `admin / admin123` 登录
3. 访问 `http://你的域名/swagger/index.html` → API 文档

---

## 九、更新部署

```bash
# 上传新的 ops-server 到 /opt/ops-platform/
chmod +x /opt/ops-platform/ops-server
systemctl restart ops-platform

# 上传新的 dist/ 到 /www/wwwroot/ops-platform/
# 无需重启，刷新浏览器即可
```

---

## 十、数据库备份

```bash
# 宝塔 → 计划任务 → 添加
# 类型：备份数据库
# 周期：每天 03:00
# 数据库：ops_platform
```

---

## 十一、项目文件说明

```
admin/                      # GitHub 仓库根目录
├── ops-server              # Go 后端二进制（Linux AMD64，42MB）
├── dist/                   # 前端编译产物（直接部署到 Nginx）
│   ├── index.html
│   └── assets/
├── migrations/             # 数据库迁移脚本（3个）
├── web/                    # 前端源码
├── cmd/                    # 后端源码
├── internal/               # 后端业务代码
├── config/                 # 配置
├── docs/                   # Swagger 文档
├── .env.example            # 环境变量模板
├── DEPLOY.md               # 本文档
├── README.md               # 项目说明
└── update.md               # 更新日志
```

| 文件 | 用途 | 部署位置 |
|------|------|---------|
| `ops-server` | 后端服务 | `/opt/ops-platform/` |
| `dist/*` | 前端页面 | `/www/wwwroot/ops-platform/` |
| `migrations/*.sql` | 建表脚本 | 执行一次即可 |
| `.env.example` | 配置模板 | 复制为 `.env` 使用 |
