#!/bin/bash

# ============================================
# 运维管理平台 - 彻底清理脚本
# 清除所有安装残留，为全新安装做准备
# ============================================

echo "=========================================="
echo "  运维管理平台 - 彻底清理脚本"
echo "=========================================="
echo ""

if [ "$(id -u)" -ne 0 ]; then
    echo "[ERROR] 请使用 root 权限运行: sudo bash clean.sh"
    exit 1
fi

echo "  ⚠ 此脚本将清除所有运维管理平台的安装残留"
echo ""

# Detect install directory
INSTALL_DIR=""
if [ -f "/opt/ops-platform/ops-server" ]; then
    INSTALL_DIR="/opt/ops-platform"
elif [ -f "$(pwd)/ops-server" ] && [ -f "$(pwd)/index.html" ]; then
    INSTALL_DIR="$(pwd)"
fi

echo "  检测到的安装目录: ${INSTALL_DIR:-未找到}"
echo ""
read -p "  确认清理？(y/N): " CONFIRM < /dev/tty
if [ "$CONFIRM" != "y" ] && [ "$CONFIRM" != "Y" ]; then
    echo "已取消"
    exit 0
fi

# ============================================
echo ""
echo "[1/7] 停止并删除 systemd 服务..."
echo "-------------------------------------------"

if systemctl is-active --quiet ops-platform 2>/dev/null; then
    systemctl stop ops-platform
    echo "[OK] 服务已停止"
fi

if [ -f /etc/systemd/system/ops-platform.service ]; then
    systemctl disable ops-platform 2>/dev/null || true
    rm -f /etc/systemd/system/ops-platform.service
    systemctl daemon-reload
    echo "[OK] 服务文件已删除"
else
    echo "[INFO] 服务文件不存在"
fi

# Kill any remaining processes
pkill -f "ops-server" 2>/dev/null && echo "[OK] 已杀死残留进程" || true
pkill -f "ops-supervisor" 2>/dev/null && echo "[OK] 已杀死残留守护进程" || true

# ============================================
echo ""
echo "[2/7] 清理 Nginx 配置..."
echo "-------------------------------------------"

# Detect OS
OS="unknown"
if [ -f /etc/os-release ]; then
    . /etc/os-release
    OS=$ID
fi

NGINX_CONF=""
if [ "$OS" = "ubuntu" ] || [ "$OS" = "debian" ]; then
    NGINX_CONF="/etc/nginx/sites-available/ops-platform"
    rm -f /etc/nginx/sites-enabled/ops-platform
else
    NGINX_CONF="/etc/nginx/conf.d/ops-platform.conf"
fi

if [ -f "$NGINX_CONF" ]; then
    rm -f "$NGINX_CONF"
    nginx -t 2>/dev/null && systemctl reload nginx 2>/dev/null || true
    echo "[OK] Nginx 配置已删除"
else
    echo "[INFO] Nginx 配置不存在"
fi

# ============================================
echo ""
echo "[3/7] 清理安装目录..."
echo "-------------------------------------------"

if [ -n "$INSTALL_DIR" ] && [ -d "$INSTALL_DIR" ]; then
    echo "  目录: $INSTALL_DIR"
    echo "  内容:"
    ls -la "$INSTALL_DIR" | head -15
    echo ""

    # Backup .env just in case
    if [ -f "$INSTALL_DIR/.env" ]; then
        cp "$INSTALL_DIR/.env" /tmp/ops-env-backup-$(date +%s) 2>/dev/null
        echo "[INFO] .env 已备份到 /tmp/"
    fi

    # Remove everything except uploads (ask separately)
    find "$INSTALL_DIR" -maxdepth 1 ! -name "uploads" ! -name "." -exec rm -rf {} \; 2>/dev/null
    echo "[OK] 安装目录已清理（保留 uploads/）"

    # Ask about uploads
    if [ -d "$INSTALL_DIR/uploads" ]; then
        UPLOAD_COUNT=$(find "$INSTALL_DIR/uploads" -type f | wc -l)
        echo ""
        read -p "  是否删除 uploads/ 目录（${UPLOAD_COUNT} 个文件）？(y/N): " UP_CONFIRM < /dev/tty
        if [ "$UP_CONFIRM" = "y" ] || [ "$UP_CONFIRM" = "Y" ]; then
            rm -rf "$INSTALL_DIR/uploads"
            echo "[OK] uploads/ 已删除"
        else
            echo "[INFO] 保留 uploads/"
        fi
    fi

    # Remove the directory itself if empty
    rmdir "$INSTALL_DIR" 2>/dev/null && echo "[OK] 安装目录已删除" || echo "[INFO] 安装目录非空，保留"
else
    echo "[INFO] 未找到安装目录"
fi

# ============================================
echo ""
echo "[4/7] 清理数据库（可选）..."
echo "-------------------------------------------"

if command -v psql &> /dev/null; then
    # List all user databases
    DB_LIST=$(sudo -u postgres psql -tAc "SELECT datname FROM pg_database WHERE datistemplate = false AND datname != 'postgres'" 2>/dev/null)
    if [ -n "$DB_LIST" ]; then
        echo "  发现以下数据库："
        echo "$DB_LIST" | while read db; do echo "    - $db"; done
        echo ""
        read -p "  ⚠ 是否卸载 PostgreSQL 并删除所有数据库？(y/N): " DB_CONFIRM < /dev/tty
        if [ "$DB_CONFIRM" = "y" ] || [ "$DB_CONFIRM" = "Y" ]; then
            systemctl stop postgresql 2>/dev/null || true

            if [ "$OS" = "ubuntu" ] || [ "$OS" = "debian" ]; then
                apt-get remove -y --purge postgresql postgresql-client postgresql-common 2>/dev/null
                apt-get autoremove -y 2>/dev/null
                rm -rf /var/lib/postgresql 2>/dev/null
                rm -rf /etc/postgresql 2>/dev/null
            elif [ "$OS" = "centos" ] || [ "$OS" = "rocky" ] || [ "$OS" = "almalinux" ]; then
                yum remove -y postgresql-server postgresql 2>/dev/null
                rm -rf /var/lib/pgsql 2>/dev/null
            fi
            echo "[OK] PostgreSQL 已卸载，所有数据库已删除"
        else
            echo "[INFO] 保留 PostgreSQL 和数据库"
        fi
    else
        echo "[INFO] 没有用户数据库"
    fi
else
    echo "[INFO] psql 未找到，跳过数据库清理"
fi

# ============================================
echo ""
echo "[5/7] 清理临时文件和备份..."
echo "-------------------------------------------"

# Remove temp backups
rm -rf /tmp/ops-uploads-backup 2>/dev/null
rm -rf /tmp/ops-clone 2>/dev/null
rm -rf /tmp/ops-env-backup-* 2>/dev/null
rm -rf /tmp/ops-platform-backup-* 2>/dev/null
echo "[OK] /tmp/ 临时文件已清理"

# Remove .env backups in common locations
rm -f /root/.env 2>/dev/null
rm -f /tmp/.env 2>/dev/null
echo "[OK] 残留 .env 文件已清理"

# ============================================
echo ""
echo "[6/7] 重置 PostgreSQL 认证（可选）..."
echo "-------------------------------------------"

PG_CONF=$(find /etc/postgresql -maxdepth 2 -name "pg_hba.conf" 2>/dev/null | head -1)
if [ -n "$PG_CONF" ] && [ -f "$PG_CONF.bak" ]; then
    read -p "  是否恢复 pg_hba.conf 原始配置？(y/N): " PG_CONFIRM < /dev/tty
    if [ "$PG_CONFIRM" = "y" ] || [ "$PG_CONFIRM" = "Y" ]; then
        cp "$PG_CONF.bak" "$PG_CONF"
        systemctl restart postgresql 2>/dev/null || true
        echo "[OK] pg_hba.conf 已恢复原始配置"
    else
        echo "[INFO] 保留当前 pg_hba.conf"
    fi
else
    echo "[INFO] 无 pg_hba.conf 备份可恢复"
fi

# ============================================
echo ""
echo "[7/7] 清理完成"
echo "=========================================="
echo ""
echo "  已清理："
echo "    ✓ systemd 服务 (ops-platform.service)"
echo "    ✓ Nginx 站点配置"
echo "    ✓ 安装目录文件"
echo "    ✓ 残留进程"
echo "    ✓ 临时文件和备份"
echo ""
echo "  如需重新安装："
echo "    curl -fsSL https://gitee.com/wxbns/Team-Management/raw/main/install-cn.sh -o install-cn.sh"
echo "    sudo bash install-cn.sh"
echo ""
