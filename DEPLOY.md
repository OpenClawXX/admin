# 宝塔面板部署指南

## 一、服务器要求

| 项目 | 最低配置 | 推荐配置 |
|------|---------|---------|
| 操作系统 | CentOS 7+ / Ubuntu 20.04+ | Ubuntu 22.04 LTS |
| CPU | 2 核 | 4 核 |
| 内存 | 2 GB | 4 GB |
| 硬盘 | 40 GB SSD | 80 GB SSD |
| 带宽 | 3 Mbps | 5 Mbps+ |

---

## 二、安装宝塔面板

```bash
# CentOS
yum install -y wget && wget -O install.sh https://download.bt.cn/install/install_6.0.sh && sh install.sh ed8484bec

# Ubuntu
wget -O install.sh https://download.bt.cn/install/install-ubuntu_6.0.sh && sudo bash install.sh ed8484bec
```

安装完成后记录面板地址、用户名、密码。

---

## 三、安装基础软件

登录宝塔面板 → **软件商店** → 安装以下软件：

| 软件 | 版本 | 说明 |
|------|------|------|
| Nginx | 1.24+ | 反向代理 + 前端静态文件 |
| PostgreSQL | 15+ | 数据库 |
| PM2管理器 | 最新 | Node.js 进程管理（可选） |

---

## 四、安装 Go 环境

```bash
# 下载最新 Go
cd /tmp
wget https://go.dev/dl/go1.24.4.linux-amd64.tar.gz

# 解压到 /usr/local
rm -rf /usr/local/go && tar -C /usr/local -xzf go1.24.4.linux-amd64.tar.gz

# 配置环境变量
echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
echo 'export GOPATH=/opt/go' >> /etc/profile
echo 'export PATH=$PATH:$GOPATH/bin' >> /etc/profile
source /etc/profile

# 验证
go version
```

---

## 五、安装 Node.js 环境

```bash
# 使用 nvm 安装 Node.js 22
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.0/install.sh | bash
source ~/.bashrc
nvm install 22
nvm use 22

# 或者使用宝塔面板 Node.js 版本管理器直接安装

# 验证
node -v
npm -v
```

---

## 六、部署项目

### 6.1 克隆代码

```bash
cd /opt
git clone https://github.com/Mcloud136/admin.git
cd admin
```

### 6.2 配置数据库

```bash
# 创建数据库
sudo -u postgres psql -c "CREATE DATABASE ops_platform;"

# 导入表结构
sudo -u postgres psql -d ops_platform -f migrations/001_init.sql
sudo -u postgres psql -d ops_platform -f migrations/002_project_enhance.sql
sudo -u postgres psql -d ops_platform -f migrations/003_knowledge_base.sql
```

### 6.3 配置后端

```bash
cp .env.example .env
```

编辑 `.env` 文件：

```env
SERVER_PORT=8080
GIN_MODE=release

DB_HOST=127.0.0.1
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=你的数据库密码
DB_NAME=ops_platform
DB_SSLMODE=disable

REDIS_ADDR=127.0.0.1:6379
REDIS_PASSWORD=
REDIS_DB=0

JWT_SECRET=替换为一个随机长字符串
JWT_EXPIRE_HOUR=24
```

生成 JWT Secret：

```bash
openssl rand -base64 32
```

### 6.4 编译后端

```bash
cd /opt/admin
go mod tidy
CGO_ENABLED=0 GOOS=linux go build -o ops-server ./cmd/server/
```

### 6.5 创建 systemd 服务

```bash
cat > /etc/systemd/system/ops-platform.service << 'EOF'
[Unit]
Description=Ops Platform Backend
After=network.target postgresql.service

[Service]
Type=simple
User=root
WorkingDirectory=/opt/admin
ExecStart=/opt/admin/ops-server
Restart=always
RestartSec=5
Environment=GIN_MODE=release

[Install]
WantedBy=multi-user.target
EOF
```

```bash
# 启动服务
systemctl daemon-reload
systemctl enable ops-platform
systemctl start ops-platform

# 查看状态
systemctl status ops-platform

# 查看日志
journalctl -u ops-platform -f
```

### 6.6 构建前端

```bash
cd /opt/admin/web
npm install
npm run build
```

构建产物在 `web/dist/` 目录。

### 6.7 部署前端

```bash
# 复制构建产物到网站目录
mkdir -p /www/wwwroot/ops-platform
cp -r /opt/admin/web/dist/* /www/wwwroot/ops-platform/
```

---

## 七、Nginx 配置

### 7.1 宝塔面板配置

1. **网站** → **添加站点**
2. 域名填写你的域名（如 `ops.example.com`）
3. 根目录填写 `/www/wwwroot/ops-platform`
4. 数据库选择 **不创建**

### 7.2 配置反向代理

点击站点 → **反向代理** → **添加反向代理**：

| 配置项 | 值 |
|--------|-----|
| 代理名称 | api |
| 目标URL | http://127.0.0.1:8080 |
| 发送域名 | $host |

### 7.3 或手动编辑 Nginx 配置

点击站点 → **配置文件**，替换为：

```nginx
server {
    listen 80;
    server_name ops.example.com;  # 替换为你的域名

    # 前端静态文件
    root /www/wwwroot/ops-platform;
    index index.html;

    # API 反向代理
    location /api/ {
        proxy_pass http://127.0.0.1:8080/api/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_read_timeout 300s;
        proxy_send_timeout 300s;
        client_max_body_size 50m;
    }

    # 文件上传大小限制
    client_max_body_size 50m;

    # Vue Router history 模式
    location / {
        try_files $uri $uri/ /index.html;
    }

    # 静态资源缓存
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
        expires 30d;
        add_header Cache-Control "public, immutable";
    }
}
```

### 7.4 重启 Nginx

```bash
nginx -t && systemctl reload nginx
```

---

## 八、SSL/HTTPS 配置（推荐）

1. 宝塔面板 → 站点 → **SSL**
2. 选择 **Let's Encrypt** 免费证书
3. 勾选域名，点击申请
4. 开启 **强制 HTTPS**

---

## 九、初始化数据

访问系统后使用以下账号登录：

```bash
# 通过 API 创建管理员
curl -X POST http://你的域名/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# 如果没有管理员账号，直接操作数据库插入：
sudo -u postgres psql -d ops_platform -c "
INSERT INTO users (username, password, real_name, email, role, status)
VALUES ('admin', '\\\$2a\\\$10\\\$这里替换为bcrypt哈希值', 'Admin', 'admin@example.com', 'admin', 1);
"
```

生成密码哈希：

```bash
cd /opt/admin
cat > /tmp/hashpw.go << 'EOF'
package main
import (
    "fmt"
    "golang.org/x/crypto/bcrypt"
)
func main() {
    hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
    fmt.Print(string(hash))
}
EOF
go run /tmp/hashpw.go
```

---

## 十、目录结构（部署后）

```
/opt/admin/                     # 项目源码
├── ops-server                  # 编译后的二进制
├── .env                        # 环境配置
├── uploads/                    # 上传文件存储
├── web/                        # 前端源码
└── migrations/                 # 数据库迁移

/www/wwwroot/ops-platform/      # 前端静态文件（Nginx 目录）
├── index.html
├── assets/
│   ├── js/
│   └── css/
└── ...

/etc/systemd/system/
└── ops-platform.service        # systemd 服务配置
```

---

## 十一、常用运维命令

```bash
# 服务管理
systemctl start ops-platform    # 启动
systemctl stop ops-platform     # 停止
systemctl restart ops-platform  # 重启
systemctl status ops-platform   # 状态
journalctl -u ops-platform -f   # 实时日志

# 更新部署
cd /opt/admin
git pull
go build -o ops-server ./cmd/server/
systemctl restart ops-platform

# 更新前端
cd /opt/admin/web
npm install
npm run build
cp -r dist/* /www/wwwroot/ops-platform/
```

---

## 十二、备份策略

### 数据库备份脚本

```bash
cat > /opt/admin/backup.sh << 'EOF'
#!/bin/bash
BACKUP_DIR="/opt/backups/ops-platform"
DATE=$(date +%Y%m%d_%H%M%S)
mkdir -p $BACKUP_DIR

# 备份数据库
sudo -u postgres pg_dump ops_platform | gzip > $BACKUP_DIR/db_$DATE.sql.gz

# 备份上传文件
tar -czf $BACKUP_DIR/uploads_$DATE.tar.gz /opt/admin/uploads/

# 保留最近 30 天
find $BACKUP_DIR -name "*.gz" -mtime +30 -delete

echo "Backup completed: $DATE"
EOF

chmod +x /opt/admin/backup.sh

# 添加定时任务（每天凌晨3点）
echo "0 3 * * * /opt/admin/backup.sh >> /var/log/ops-backup.log 2>&1" | crontab -
```

---

## 十三、防火墙配置

```bash
# 开放端口
firewall-cmd --permanent --add-port=80/tcp
firewall-cmd --permanent --add-port=443/tcp
firewall-cmd --reload

# 或在宝塔面板 → 安全 → 防火墙 中放行 80 和 443
```

---

## 十四、常见问题

### 1. 后端启动失败
```bash
# 检查日志
journalctl -u ops-platform --no-pager -n 50

# 常见原因：数据库连接失败
# 检查 PostgreSQL 是否运行
systemctl status postgresql

# 检查连接
psql -U postgres -h 127.0.0.1 -d ops_platform
```

### 2. 前端白屏
```bash
# 检查 Nginx 配置
nginx -t

# 检查前端文件是否存在
ls /www/wwwroot/ops-platform/index.html

# 检查 Nginx 日志
tail -f /www/wwwlogs/ops-platform.error.log
```

### 3. API 404
```bash
# 检查后端是否运行
systemctl status ops-platform

# 检查端口是否监听
netstat -tlnp | grep 8080

# 检查 Nginx 反向代理配置
cat /www/server/panel/vhost/nginx/ops-platform.conf
```

### 4. 文件上传失败
```bash
# 检查上传目录权限
chmod 755 /opt/admin/uploads
chown -R www:www /opt/admin/uploads

# 检查 Nginx client_max_body_size
# 确保配置了 client_max_body_size 50m;
```

---

## 十五、监控建议

| 工具 | 用途 |
|------|------|
| 宝塔面板监控 | CPU/内存/磁盘/带宽 |
| PostgreSQL pg_stat | 数据库性能 |
| Nginx 日志分析 | 访问量/错误率 |
| systemd journal | 后端服务日志 |
