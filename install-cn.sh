#!/bin/bash
set -e

# ============================================
# 运维管理平台 - 一键安装脚本（国内版）
# ============================================

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
echo "  运维管理平台 - 一键安装脚本（国内版）"
echo "=========================================="
echo ""

# Check root
if [ "$(id -u)" -ne 0 ]; then
    echo "[ERROR] 请使用 root 权限运行: sudo bash install-cn.sh"
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
echo "[1/9] 配置国内镜像源..."
echo "-------------------------------------------"

if [ "$OS" = "ubuntu" ]; then
    echo ">> 备份原源: /etc/apt/sources.list"
    cp /etc/apt/sources.list /etc/apt/sources.list.bak 2>/dev/null || true
    CODENAME=$(lsb_release -cs 2>/dev/null || echo "jammy")
    echo ">> 系统代号: $CODENAME"
    echo ">> 写入阿里云镜像源"
    cat > /etc/apt/sources.list << APTLIST
deb http://mirrors.aliyun.com/ubuntu/ ${CODENAME} main restricted universe multiverse
deb http://mirrors.aliyun.com/ubuntu/ ${CODENAME}-updates main restricted universe multiverse
deb http://mirrors.aliyun.com/ubuntu/ ${CODENAME}-security main restricted universe multiverse
deb http://mirrors.aliyun.com/ubuntu/ ${CODENAME}-backports main restricted universe multiverse
APTLIST
    echo "[OK] Ubuntu 镜像源已切换为阿里云"

elif [ "$OS" = "debian" ]; then
    echo ">> 备份原源"
    cp /etc/apt/sources.list /etc/apt/sources.list.bak 2>/dev/null || true
    CODENAME=$(lsb_release -cs 2>/dev/null || echo "bookworm")
    echo ">> 系统代号: $CODENAME"
    echo ">> 写入阿里云镜像源"
    cat > /etc/apt/sources.list << APTLIST
deb http://mirrors.aliyun.com/debian/ ${CODENAME} main contrib non-free non-free-firmware
deb http://mirrors.aliyun.com/debian/ ${CODENAME}-updates main contrib non-free non-free-firmware
deb http://mirrors.aliyun.com/debian/ ${CODENAME}-security main contrib non-free non-free-firmware
APTLIST
    echo "[OK] Debian 镜像源已切换为阿里云"

elif [ "$OS" = "centos" ] || [ "$OS" = "rocky" ] || [ "$OS" = "almalinux" ]; then
    echo ">> 备份原源"
    cp /etc/yum.repos.d/*.repo /etc/yum.repos.d/*.repo.bak 2>/dev/null || true
    if [ "$OS" = "centos" ]; then
        echo ">> 替换为阿里云镜像"
        sed -i 's|mirror.centos.org|mirrors.aliyun.com|g' /etc/yum.repos.d/*.repo
    fi
    echo "[OK] YUM 镜像源已切换为阿里云"
fi

# ============================================
echo ""
echo "[2/9] 安装系统依赖..."
echo "-------------------------------------------"

if [ "$OS" = "ubuntu" ] || [ "$OS" = "debian" ]; then
    # 安装最新 PostgreSQL 18
    echo ">> 添加 PostgreSQL 官方源"
    sh -c 'echo "deb http://apt.postgresql.org/pub/repos/apt $(lsb_release -cs)-pgdg main" > /etc/apt/sources.list.d/pgdg.list'
    curl -fsSL https://www.postgresql.org/media/keys/ACCC4CF8.asc | gpg --dearmor -o /etc/apt/trusted.gpg.d/postgresql.gpg 2>/dev/null || \
    wget -qO- https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add - 2>/dev/null

    echo ">> apt-get update"
    apt-get update

    echo ">> 安装 PostgreSQL 18"
    apt-get install -y postgresql-18 postgresql-client-18 2>/dev/null || apt-get install -y postgresql postgresql-client

    # 安装 Nginx 1.31（下载官方二进制）
    echo ">> 安装 Nginx 1.31（含 HTTP/3 + Brotli）"
    NGINX_VER="1.31.1"
    apt-get install -y libpcre3 libpcre3-dev zlib1g zlib1g-dev libssl-dev build-essential cmake 2>/dev/null || true

    cd /tmp

    # 克隆 Brotli 模块
    echo ">> 下载 ngx_brotli 模块"
    git clone --depth 1 https://github.com/google/ngx_brotli.git 2>/dev/null
    cd ngx_brotli && git submodule update --init && cd /tmp

    # 下载 Nginx 源码
    curl -fsSL "http://nginx.org/download/nginx-${NGINX_VER}.tar.gz" -o nginx.tar.gz
    tar xzf nginx.tar.gz
    cd nginx-${NGINX_VER}

    # 编译（含 HTTP/3 + Brotli + 常用模块）
    ./configure \
        --prefix=/etc/nginx \
        --sbin-path=/usr/sbin/nginx \
        --modules-path=/usr/lib64/nginx/modules \
        --conf-path=/etc/nginx/nginx.conf \
        --error-log-path=/var/log/nginx/error.log \
        --http-log-path=/var/log/nginx/access.log \
        --pid-path=/var/run/nginx.pid \
        --with-http_ssl_module \
        --with-http_v2_module \
        --with-http_v3_module \
        --with-http_gzip_static_module \
        --with-http_realip_module \
        --with-http_stub_status_module \
        --with-http_secure_link_module \
        --with-pcre \
        --add-module=/tmp/ngx_brotli

    make -j$(nproc)
    make install
    cd "$WORK_DIR"
    rm -rf /tmp/nginx-* /tmp/nginx.tar.gz /tmp/ngx_brotli

    # 创建 systemd 服务文件
    cat > /etc/systemd/system/nginx.service << 'NGINXSVC'
[Unit]
Description=nginx - high performance web server
Documentation=http://nginx.org/en/docs/
After=network.target remote-fs.target nss-lookup.target

[Service]
Type=forking
PIDFile=/var/run/nginx.pid
ExecStartPre=/usr/sbin/nginx -t
ExecStart=/usr/sbin/nginx
ExecReload=/bin/kill -s HUP $MAINPID
ExecStop=/bin/kill -s QUIT $MAINPID
PrivateTmp=true

[Install]
WantedBy=multi-user.target
NGINXSVC
    systemctl daemon-reload
    systemctl enable nginx

    # 创建必要目录
    mkdir -p /var/log/nginx /var/cache/nginx /etc/nginx/conf.d /etc/nginx/sites-available /etc/nginx/sites-enabled

elif [ "$OS" = "centos" ] || [ "$OS" = "rocky" ] || [ "$OS" = "almalinux" ]; then
    # 安装最新 PostgreSQL
    echo ">> 添加 PostgreSQL 官方源"
    yum install -y https://download.postgresql.org/pub/repos/yum/reporpms/EL-$(rpm -E %{rhel})-x86_64/pgdg-redhat-repo-latest.noarch.rpm
    yum -y module disable postgresql
    yum install -y postgresql18-server postgresql18

    echo ">> 安装 Nginx"
    yum install -y epel-release
    yum install -y nginx

    echo ">> 初始化 PostgreSQL"
    /usr/pgsql-18/bin/postgresql-18-setup initdb
    systemctl enable postgresql-18
    systemctl start postgresql-18
fi

# 获取 PostgreSQL 版本号
PG_VERSION=$(psql --version 2>/dev/null | grep -oP '\d+' | head -1)
echo "[OK] PostgreSQL 版本: $PG_VERSION"
echo "[OK] Nginx 版本: $(nginx -v 2>&1 | grep -oP '[\d.]+')"
echo "[OK] 系统依赖安装完成"

# ============================================
echo ""
echo "[3/9] 下载项目文件 (Gitee)..."
echo "-------------------------------------------"

echo ">> git clone https://gitee.com/wxbns/Team-Management.git (temp)"
TEMP_DIR=$(mktemp -d)
git clone --depth 1 https://gitee.com/wxbns/Team-Management.git "$TEMP_DIR"
cp -a "$TEMP_DIR"/. "$WORK_DIR/"
rm -rf "$TEMP_DIR"
echo "[OK] 项目文件下载完成"

# ============================================
echo ""
echo "[4/9] 配置数据库密码..."
echo "-------------------------------------------"

# 生成随机数据库密码
DB_PASSWORD=$(openssl rand -base64 24 | tr -d '/+=' | head -c 32)
echo "[OK] 已生成随机数据库密码"

echo ">> 确保 PostgreSQL 运行"
# 创建集群（如果不存在）并启动
pg_createcluster "$PG_VERSION" main --start 2>/dev/null || true
systemctl start postgresql || true
sleep 2

echo ">> 设置 postgres 用户密码"
if sudo -u postgres psql -c "ALTER USER postgres PASSWORD '${DB_PASSWORD}';"; then
    echo "[OK] 数据库密码配置完成"
    # 确认成功后才保存密码文件供安装向导读取
    echo "$DB_PASSWORD" > .db_password
    chmod 600 .db_password
    echo "[OK] 密码已保存到 .db_password"
else
    echo "[ERROR] 数据库密码设置失败"
    exit 1
fi

# ============================================
echo ""
echo "[5/9] 设置文件权限..."
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
echo "[6/9] 创建数据库..."
echo "-------------------------------------------"

echo ">> CREATE DATABASE ops_platform"
sudo -u postgres psql -c "CREATE DATABASE ops_platform ENCODING 'UTF8';" || echo "   数据库已存在，跳过"
echo "[OK] 数据库准备完成"

# ============================================
echo ""
echo "[7/9] 配置 Nginx..."
echo "-------------------------------------------"

SERVER_IP=$(hostname -I | awk '{print $1}')
echo ">> 服务器 IP: $SERVER_IP"

if [ "$OS" = "ubuntu" ] || [ "$OS" = "debian" ]; then
    NGINX_CONF="/etc/nginx/sites-available/ops-platform"
elif [ "$OS" = "centos" ] || [ "$OS" = "rocky" ] || [ "$OS" = "almalinux" ]; then
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
        sendfile off;
        add_header Cache-Control "no-cache, no-store, must-revalidate" always;
        add_header Pragma "no-cache" always;
        add_header Expires "0" always;
        add_header X-Content-Type-Options "nosniff" always;
        add_header X-Frame-Options "SAMEORIGIN" always;
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
echo ">> 重载 Nginx"
echo ">> 重启 Nginx（确保新组权限生效）"
systemctl restart nginx
echo "[OK] Nginx 配置完成"

# ============================================
echo ""
echo "[8/9] 配置系统服务..."
echo "-------------------------------------------"

echo ">> 创建 ops-platform 系统用户"
if ! id "ops-platform" &>/dev/null; then
    useradd --system --no-create-home --shell /usr/sbin/nologin ops-platform
    echo "[OK] 用户 ops-platform 已创建"
else
    echo "[OK] 用户 ops-platform 已存在"
fi

echo ">> 写入 systemd 服务: /etc/systemd/system/ops-platform.service"
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
echo "[9/9] 安装完成"
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
