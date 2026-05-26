# 宝塔面板部署指南

## 一、服务器准备

### 1.1 推荐配置

| 项目 | 最低 | 推荐 |
|------|------|------|
| 系统 | Ubuntu 20.04 | Ubuntu 22.04 LTS |
| CPU | 2 核 | 4 核 |
| 内存 | 2 GB | 4 GB |
| 硬盘 | 40 GB SSD | 80 GB SSD |

### 1.2 安装宝塔面板

```bash
wget -O install.sh https://download.bt.cn/install/install-ubuntu_6.0.sh && sudo bash install.sh ed8484bec
```

安装完成后记录面板地址、用户名、密码。

### 1.3 安装基础软件

登录宝塔面板 → **软件商店** → 安装：

| 软件 | 说明 |
|------|------|
| **Nginx** | 反向代理 + 前端托管 |
| **PostgreSQL** | 数据库 |

---

## 二、项目文件说明

```
admin/                      # GitHub 仓库根目录
├── ops-server              # 后端二进制 (42MB, Linux AMD64)
├── dist/                   # 前端编译产物 (35个文件)
│   ├── index.html
│   └── assets/
├── migrations/             # 数据库迁移脚本
│   ├── 001_init.sql
│   ├── 002_project_enhance.sql
│   └── 003_knowledge_base.sql
├── .env.example            # 后端配置模板
└── web/                    # 前端源码（如需重新编译）
```

---

## 三、部署后端

### 3.1 创建目录

```bash
mkdir -p /opt/ops-platform/uploads/branding
mkdir -p /opt/ops-platform/uploads/kb
```

### 3.2 上传 ops-server

将 `ops-server` 上传到 `/opt/ops-platform/`

```bash
chmod +x /opt/ops-platform/ops-server
```

### 3.3 创建配置文件

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

JWT_SECRET=运行 openssl rand -base64 32 生成
JWT_EXPIRE_HOUR=24
EOF
```

### 3.4 创建数据库

宝塔 → PostgreSQL → 管理 → SQL执行器，依次执行三个迁移文件：

1. `migrations/001_init.sql`
2. `migrations/002_project_enhance.sql`
3. `migrations/003_knowledge_base.sql`

### 3.5 创建 systemd 服务

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
```

验证：`systemctl status ops-platform` 应显示 `active (running)`

---

## 四、部署前端

### 4.1 上传 dist 文件

将 `dist/` 内所有文件上传到 `/www/wwwroot/ops-platform/`

```
/www/wwwroot/ops-platform/
├── index.html
└── assets/
    ├── *.js
    └── *.css
```

### 4.2 添加站点

宝塔 → 网站 → 添加站点：
- 域名：你的域名或服务器IP
- 根目录：`/www/wwwroot/ops-platform`
- PHP版本：**纯静态**

### 4.3 配置 Nginx

点击站点名 → **配置文件**，全部替换为：

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

    # 上传文件
    location /uploads/ {
        proxy_pass http://127.0.0.1:8080/uploads/;
    }

    # Swagger 文档
    location /swagger/ {
        proxy_pass http://127.0.0.1:8080/swagger/;
    }

    # Vue Router history 模式
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

保存后重载：`nginx -t && systemctl reload nginx`

---

## 五、防火墙

宝塔 → 安全 → 防火墙 → 放行 80 和 443

---

## 六、SSL 证书（推荐）

站点 → SSL → Let's Encrypt → 申请 → 开启强制 HTTPS

---

## 七、验证部署

| 检查项 | 地址 | 预期结果 |
|--------|------|---------|
| 登录页 | `http://你的域名` | 显示登录页面 |
| 登录 | admin / admin123 | 进入工作台 |
| API文档 | `http://你的域名/swagger/index.html` | 显示Swagger |

---

## 八、日常运维

### 更新后端

```bash
# 上传新的 ops-server 覆盖旧文件
chmod +x /opt/ops-platform/ops-server
systemctl restart ops-platform
```

### 更新前端

```bash
# 上传新的 dist/ 覆盖旧文件到 /www/wwwroot/ops-platform/
# 无需重启，刷新浏览器即可
```

### 更新数据库

```bash
# 执行新的迁移脚本
psql -U postgres -d ops_platform -f /path/to/new_migration.sql
```

### 查看日志

```bash
journalctl -u ops-platform -f
```

### 重启服务

```bash
systemctl restart ops-platform
```

---

## 九、数据备份

宝塔 → 计划任务 → 添加：

| 类型 | 周期 | 说明 |
|------|------|------|
| 备份数据库 | 每天 03:00 | 选择 PostgreSQL → ops_platform |
| 备份网站 | 每天 03:30 | 选择站点 |

---

## 十、常见问题

| 问题 | 排查 |
|------|------|
| 白屏 | 检查 `index.html` 是否在 `/www/wwwroot/ops-platform/` |
| API 404 | `systemctl status ops-platform` 检查后端是否运行 |
| 登录失败 | 检查 `.env` 中数据库密码是否正确 |
| 上传失败 | 检查 `uploads/` 目录权限：`chmod 755 /opt/ops-platform/uploads` |
| 忘记密码 | 数据库执行 `UPDATE users SET password='...' WHERE username='admin'` |
