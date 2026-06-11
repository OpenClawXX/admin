#!/bin/bash
set -e

cleanup() {
    exit_code=$?
    if [ $exit_code -ne 0 ]; then
        echo ""
        echo "[ERROR] 安装失败 (exit code: $exit_code)"
        echo "[ERROR] 请检查错误信息后重试"
        # Clean up temp password file on failure
        rm -f .db_password
    fi
}
trap cleanup EXIT

WORK_DIR=$(pwd)

# Safety check: prevent accidental damage to system directories
case "$WORK_DIR" in
    /|/root|/home|/etc|/usr|/var|/tmp|/bin|/sbin|/opt)
        echo "[ERROR] 请不要在系统目录中运行此脚本"
        echo "[ERROR] 请创建一个专用目录，例如: mkdir -p /opt/ops-platform && cd /opt/ops-platform"
        exit 1
        ;;
esac

echo "=========================================="
echo "  运维管理平台 - 一键安装脚本"
echo "=========================================="
echo ""

# Check root
if [ "$(id -u)" -ne 0 ]; then
    echo "[ERROR] 请使用 root 权限运行: sudo bash install.sh"
    exit 1
fi
echo "[OK] root 权限确认"

# Detect OS
if [ -f /etc/os-release ]; then
    . /etc/os-release
    OS=$ID
    echo "[OK] 检测到系统: $OS"
else
    echo "[ERROR] 不支持的操作系统"
    exit 1
fi

# ============================================
echo ""
echo "[1/7] 安装系统依赖..."
echo "-------------------------------------------"

if [ "$OS" = "ubuntu" ] || [ "$OS" = "debian" ]; then
    echo ">> apt-get update"
    apt-get update
    echo ""
    echo ">> apt-get install nginx postgresql postgresql-client"
    apt-get install -y nginx postgresql postgresql-client
elif [ "$OS" = "centos" ] || [ "$OS" = "rhel" ] || [ "$OS" = "rocky" ]; then
    echo ">> yum install nginx postgresql-server postgresql"
    yum install -y nginx postgresql-server postgresql
    echo ""
    echo ">> 初始化 PostgreSQL"
    postgresql-setup --initdb || true
    echo ">> 启动 PostgreSQL"
    systemctl enable postgresql
    systemctl start postgresql
else
    echo "[WARN] 未知系统，请手动安装 Nginx 和 PostgreSQL"
fi
echo "[OK] 系统依赖安装完成"

# ============================================
echo ""
echo "[2/7] 下载项目文件..."
echo "-------------------------------------------"

echo ">> git clone https://github.com/Mcloud136/admin.git (temp)"
TEMP_DIR=$(mktemp -d)
git clone --depth 1 https://github.com/Mcloud136/admin.git "$TEMP_DIR"
cp -a "$TEMP_DIR"/. "$WORK_DIR/"
rm -rf "$TEMP_DIR"
echo "[OK] 项目文件下载完成"

# ============================================
echo ""
echo "[3/7] 配置数据库密码..."
echo "-------------------------------------------"

# 生成随机数据库密码（在克隆成功后生成，避免克隆失败时留下密码文件）
DB_PASSWORD=$(openssl rand -base64 24 | tr -d '/+=' | head -c 32)
echo "[OK] 已生成随机数据库密码"

# 保存密码到临时文件供安装向导读取
echo "$DB_PASSWORD" > .db_password
chmod 600 .db_password
echo "[OK] 密码已保存到 .db_password"

echo ">> 确保 PostgreSQL 运行"
systemctl start postgresql || true
sleep 2

echo ">> 设置 postgres 用户密码"
sudo -u postgres psql -c "ALTER USER postgres PASSWORD '${DB_PASSWORD}';"
echo "[OK] 数据库密码配置完成"

# ============================================
echo ""
echo "[4/7] 设置文件权限..."
echo "-------------------------------------------"

echo ">> chmod +x ops-server"
chmod +x "$WORK_DIR/ops-server"
echo ">> chmod +x ops-supervisor"
chmod +x "$WORK_DIR/ops-supervisor"
echo ">> 创建 uploads/ 目录"
mkdir -p "$WORK_DIR/uploads/branding"
mkdir -p "$WORK_DIR/uploads/kb"
echo "[OK] 权限设置完成"

# ============================================
echo ""
echo "[5/7] 创建数据库..."
echo "-------------------------------------------"

echo ">> CREATE DATABASE ops_platform"
sudo -u postgres psql -c "CREATE DATABASE ops_platform;" || echo "   数据库已存在，跳过"
echo "[OK] 数据库准备完成"

# ============================================
echo ""
echo "[6/7] 配置 Nginx..."
echo "-------------------------------------------"

SERVER_IP=$(hostname -I | awk '{print $1}')
echo ">> 服务器 IP: $SERVER_IP"

NGINX_CONF="/etc/nginx/sites-available/ops-platform"
if [ "$OS" = "centos" ] || [ "$OS" = "rhel" ] || [ "$OS" = "rocky" ]; then
    NGINX_CONF="/etc/nginx/conf.d/ops-platform.conf"
fi

echo ">> 写入 Nginx 配置: $NGINX_CONF"
cat > "$NGINX_CONF" << NGINXEOF
server {
    listen 80;
    server_name _;

    root $WORK_DIR;
    index index.html;

    client_max_body_size 50m;

    # Security headers
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;

    location /api/ {
        proxy_pass http://127.0.0.1:8080/api/;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_read_timeout 300s;
    }

    location /swagger/ {
        proxy_pass http://127.0.0.1:8080/swagger/;
    }

    location / {
        add_header Cache-Control "no-cache, no-store, must-revalidate";
        add_header Pragma "no-cache";
        add_header Expires "0";
        try_files \$uri \$uri/ /index.html;
    }

    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2)$ {
        expires 30d;
        add_header Cache-Control "public, immutable";
    }

    gzip on;
    gzip_types text/plain text/css application/json application/javascript text/xml;
    gzip_min_length 1024;
}
NGINXEOF

if [ "$OS" = "ubuntu" ] || [ "$OS" = "debian" ]; then
    echo ">> 创建 sites-enabled 软链接"
    ln -sf "$NGINX_CONF" /etc/nginx/sites-enabled/ops-platform
    rm -f /etc/nginx/sites-enabled/default
fi

echo ">> 测试 Nginx 配置"
nginx -t
echo ">> 重启 Nginx（确保新组权限生效）"
systemctl restart nginx
echo "[OK] Nginx 配置完成"

# ============================================
echo ""
echo "[7/7] 配置系统服务..."
echo "-------------------------------------------"

echo ">> 创建 ops-platform 系统用户"
if ! id "ops-platform" &>/dev/null; then
    useradd --system --no-create-home --shell /usr/sbin/nologin ops-platform
    echo "[OK] 用户 ops-platform 已创建"
else
    echo "[OK] 用户 ops-platform 已存在"
fi

echo ">> 写入 systemd 服务文件"
cat > /etc/systemd/system/ops-platform.service << SVCEOF
[Unit]
Description=Ops Platform Supervisor
After=network.target postgresql.service

[Service]
Type=simple
User=ops-platform
WorkingDirectory=$WORK_DIR
ExecStart=$WORK_DIR/ops-supervisor
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
SVCEOF

echo ">> systemctl daemon-reload"
systemctl daemon-reload
echo ">> 设置文件权限"
chown -R ops-platform:ops-platform "$WORK_DIR"
# Allow Nginx (www-data) to read static files
usermod -a -G ops-platform www-data
# Directories need execute permission for traversal
chmod 750 "$WORK_DIR"
chmod 750 "$WORK_DIR/assets" 2>/dev/null || true
chmod 750 "$WORK_DIR/config" 2>/dev/null || true
chmod 750 "$WORK_DIR/internal" 2>/dev/null || true
chmod 750 "$WORK_DIR/uploads" 2>/dev/null || true
chmod 750 "$WORK_DIR/uploads/branding" 2>/dev/null || true
chmod 750 "$WORK_DIR/uploads/kb" 2>/dev/null || true
# Static files readable by group
chmod 640 "$WORK_DIR"/index.html 2>/dev/null || true
chmod 640 "$WORK_DIR"/assets/*.js 2>/dev/null || true
chmod 640 "$WORK_DIR"/assets/*.css 2>/dev/null || true
chmod 750 "$WORK_DIR/ops-server" "$WORK_DIR/ops-supervisor"
# Restart nginx (not reload) so workers pick up new group membership
systemctl restart nginx 2>/dev/null || true
echo ">> systemctl enable ops-platform"
systemctl enable ops-platform
echo ">> systemctl start ops-platform"
systemctl start ops-platform
echo ">> 等待服务启动..."
sleep 3
echo ">> 检查服务状态"
systemctl status ops-platform --no-pager || true
echo "[OK] 系统服务配置完成"

# ============================================
echo ""
echo "  安装完成"
echo "=========================================="
echo ""
echo "  访问地址: http://${SERVER_IP}"
echo ""
echo "  首次访问将进入安装向导，请按提示完成："
echo "    1. 数据库信息（默认 postgres 用户）"
echo "    2. 管理员账号密码"
echo "    3. 平台名称和公司名称"
echo ""
echo "  [NOTE] 数据库密码已保存到 .db_password，安装向导会自动读取"
echo "  [NOTE] 安装完成后 .db_password 将自动删除"
echo ""
echo "  API 文档: http://${SERVER_IP}/swagger/index.html"
echo ""
echo "  服务管理命令:"
echo "    systemctl start ops-platform    # 启动"
echo "    systemctl stop ops-platform     # 停止"
echo "    systemctl restart ops-platform  # 重启"
echo "    systemctl status ops-platform   # 状态"
echo "    journalctl -u ops-platform -f   # 实时日志"
echo ""
