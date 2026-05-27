#!/bin/bash
set -e

# ============================================
# 运维管理平台 - 一键安装脚本
# ============================================

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
INSTALL_DIR="/opt/ops-platform"
SOURCE="github"

echo "=========================================="
echo "  运维管理平台 - 一键安装脚本"
echo "=========================================="
echo ""

if [ "$(id -u)" -ne 0 ]; then
    echo "[ERROR] 请使用 root 权限运行: sudo bash install.sh"
    exit 1
fi
echo "[OK] root 权限确认"

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
    apt-get update -qq
    apt-get install -y nginx postgresql postgresql-client
    sudo -u postgres psql -c "ALTER USER postgres PASSWORD 'postgres';"
elif [ "$OS" = "centos" ] || [ "$OS" = "rocky" ] || [ "$OS" = "almalinux" ]; then
    yum install -y nginx postgresql-server postgresql
    postgresql-setup --initdb || true
    systemctl enable postgresql && systemctl start postgresql
    sudo -u postgres psql -c "ALTER USER postgres PASSWORD 'postgres';"
fi
echo "[OK] 系统依赖安装完成"

# ============================================
echo ""
echo "[2/7] 部署项目文件..."
echo "-------------------------------------------"

mkdir -p "$INSTALL_DIR"
cp "$SCRIPT_DIR/ops-server" "$INSTALL_DIR/"
cp "$SCRIPT_DIR/ops-supervisor" "$INSTALL_DIR/"
cp -r "$SCRIPT_DIR/assets" "$INSTALL_DIR/"
cp "$SCRIPT_DIR/index.html" "$INSTALL_DIR/"
cp "$SCRIPT_DIR/.env.example" "$INSTALL_DIR/.env.example"
chmod +x "$INSTALL_DIR/ops-server" "$INSTALL_DIR/ops-supervisor"
mkdir -p "$INSTALL_DIR/uploads/branding" "$INSTALL_DIR/uploads/kb"

# 写入来源标记
echo "$SOURCE" > "$INSTALL_DIR/.source"
echo "[OK] 项目文件部署完成（来源: $SOURCE）"

# ============================================
echo ""
echo "[3/7] 创建数据库..."
echo "-------------------------------------------"

systemctl start postgresql || true
sleep 2
sudo -u postgres psql -c "CREATE DATABASE ops_platform ENCODING 'UTF8';" 2>/dev/null || echo "   数据库已存在，跳过"
echo "[OK] 数据库准备完成"

# ============================================
echo ""
echo "[4/7] 配置 Nginx..."
echo "-------------------------------------------"

SERVER_IP=$(hostname -I | awk '{print $1}')

if [ "$OS" = "ubuntu" ] || [ "$OS" = "debian" ]; then
    NGINX_CONF="/etc/nginx/sites-available/ops-platform"
else
    NGINX_CONF="/etc/nginx/conf.d/ops-platform.conf"
fi

cat > "$NGINX_CONF" << 'NGINXEOF'
server {
    listen 80;
    server_name _;
    root INSTALL_DIR_PLACEHOLDER;
    index index.html;
    client_max_body_size 50m;

    location /api/ {
        proxy_pass http://127.0.0.1:8080/api/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_read_timeout 300s;
    }
    location /uploads/ {
        proxy_pass http://127.0.0.1:8080/uploads/;
    }
    location / {
        try_files $uri $uri/ /index.html;
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
sed -i "s|INSTALL_DIR_PLACEHOLDER|$INSTALL_DIR|g" "$NGINX_CONF"

if [ "$OS" = "ubuntu" ] || [ "$OS" = "debian" ]; then
    ln -sf "$NGINX_CONF" /etc/nginx/sites-enabled/ops-platform
    rm -f /etc/nginx/sites-enabled/default
fi
nginx -t && (systemctl reload nginx || systemctl restart nginx)
echo "[OK] Nginx 配置完成"

# ============================================
echo ""
echo "[5/7] 配置系统服务..."
echo "-------------------------------------------"

cat > /etc/systemd/system/ops-platform.service << SVCEOF
[Unit]
Description=Ops Platform Supervisor
After=network.target postgresql.service

[Service]
Type=simple
User=root
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/ops-supervisor
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
SVCEOF

systemctl daemon-reload
systemctl enable ops-platform
systemctl start ops-platform
sleep 3
systemctl status ops-platform --no-pager || true
echo "[OK] 系统服务配置完成"

# ============================================
echo ""
echo "[6/7] 复制更新脚本..."
echo "-------------------------------------------"
cp "$SCRIPT_DIR/update.sh" "$INSTALL_DIR/" 2>/dev/null || true
cp "$SCRIPT_DIR/update-cn.sh" "$INSTALL_DIR/" 2>/dev/null || true
chmod +x "$INSTALL_DIR/update.sh" "$INSTALL_DIR/update-cn.sh" 2>/dev/null || true
echo "[OK] 更新脚本已部署"

# ============================================
echo ""
echo "[7/7] 安装完成"
echo "=========================================="
echo ""
echo "  访问地址: http://${SERVER_IP}"
echo ""
echo "  首次访问将进入安装向导:"
echo "    1. 数据库信息（默认 postgres 用户）"
echo "    2. 管理员账号密码"
echo "    3. 平台名称和公司名称"
echo ""
echo "  更新命令: cd $INSTALL_DIR && sudo bash update.sh"
echo ""
echo "  服务管理:"
echo "    systemctl start ops-platform    # 启动"
echo "    systemctl stop ops-platform     # 停止"
echo "    systemctl restart ops-platform  # 重启"
echo "    systemctl status ops-platform   # 状态"
echo "    journalctl -u ops-platform -f   # 日志"
echo ""
