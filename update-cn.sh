#!/bin/bash
set -e

# ============================================
# 运维管理平台 - 更新脚本（国内版）
# 自动检测安装来源（GitHub / Gitee），从对应仓库拉取最新版本
# GitHub 来源自动使用国内镜像加速
# ============================================

INSTALL_DIR="/opt/ops-platform"
BACKUP_DIR="/opt/ops-platform-backup-$(date +%Y%m%d%H%M%S)"
SERVICE_NAME="ops-platform"

# 下载地址
GITHUB_DOWNLOAD="https://github.com/Mcloud136/admin/archive/refs/heads/main.tar.gz"
GITEE_DOWNLOAD="https://gitee.com/wxbns/Team-Management/repository/archive/main.tar.gz"

# 国内镜像（优先级从高到低）
GITHUB_MIRRORS=(
    "https://ghfast.top/GITHUB_URL"
    "https://mirror.ghproxy.com/GITHUB_URL"
    "https://gh-proxy.com/GITHUB_URL"
)

echo "=========================================="
echo "  运维管理平台 - 更新脚本（国内版）"
echo "=========================================="
echo ""

if [ "$(id -u)" -ne 0 ]; then
    echo "[ERROR] 请使用 root 权限运行: sudo bash update-cn.sh"
    exit 1
fi

if [ ! -d "$INSTALL_DIR" ]; then
    echo "[ERROR] 安装目录不存在: $INSTALL_DIR"
    exit 1
fi

# ============================================
# 检测安装来源
# ============================================
SOURCE_FILE="$INSTALL_DIR/.source"
if [ -f "$SOURCE_FILE" ]; then
    SOURCE=$(cat "$SOURCE_FILE" | tr -d '[:space:]')
else
    # 无标记文件，优先尝试 Gitee（国内更快）
    if curl -fsSL --connect-timeout 5 "$GITEE_DOWNLOAD" -o /dev/null 2>/dev/null; then
        SOURCE="gitee"
    else
        SOURCE="github"
    fi
    echo "[WARN] 未找到来源标记，默认使用: $SOURCE"
fi

case "$SOURCE" in
    gitee|GITEE)
        DOWNLOAD_URL="$GITEE_DOWNLOAD"
        SOURCE_NAME="Gitee"
        USE_MIRROR=false
        ;;
    github|GitHub|*)
        DOWNLOAD_URL="$GITHUB_DOWNLOAD"
        SOURCE_NAME="GitHub"
        USE_MIRROR=true
        ;;
esac

echo "[INFO] 安装来源: $SOURCE_NAME"
echo "[INFO] 安装目录: $INSTALL_DIR"
echo "[INFO] 备份目录: $BACKUP_DIR"
echo ""

# ============================================
echo "[1/6] 备份当前版本..."
echo "-------------------------------------------"

mkdir -p "$BACKUP_DIR"
cp "$INSTALL_DIR/ops-server" "$BACKUP_DIR/" 2>/dev/null || true
cp "$INSTALL_DIR/ops-supervisor" "$BACKUP_DIR/" 2>/dev/null || true
cp -r "$INSTALL_DIR/assets" "$BACKUP_DIR/" 2>/dev/null || true
cp "$INSTALL_DIR/index.html" "$BACKUP_DIR/" 2>/dev/null || true
cp "$INSTALL_DIR/.env" "$BACKUP_DIR/" 2>/dev/null || true
cp "$INSTALL_DIR/.source" "$BACKUP_DIR/" 2>/dev/null || true
echo "[OK] 备份完成"

# ============================================
echo ""
echo "[2/6] 停止服务..."
echo "-------------------------------------------"

if systemctl is-active --quiet "$SERVICE_NAME"; then
    systemctl stop "$SERVICE_NAME"
    echo "[OK] 服务已停止"
else
    echo "[INFO] 服务未运行，跳过"
fi

# ============================================
echo ""
echo "[3/6] 下载最新版本 ($SOURCE_NAME)..."
echo "-------------------------------------------"

TEMP_DIR=$(mktemp -d)
DOWNLOADED=false

download() {
    local url="$1"
    local label="$2"
    echo ">> $label..."
    if curl -fsSL --connect-timeout 15 "$url" -o "$TEMP_DIR/source.tar.gz" 2>/dev/null; then
        echo "[OK] 下载成功"
        DOWNLOADED=true
        return 0
    fi
    return 1
}

if [ "$USE_MIRROR" = true ]; then
    # GitHub：先试镜像，再试直连
    for MIRROR_PATTERN in "${GITHUB_MIRRORS[@]}"; do
        MIRROR_URL="${MIRROR_PATTERN//GITHUB_URL/$GITHUB_DOWNLOAD}"
        MIRROR_HOST=$(echo "$MIRROR_URL" | cut -d'/' -f3)
        download "$MIRROR_URL" "镜像: $MIRROR_HOST" && break
    done
    [ "$DOWNLOADED" = false ] && download "$GITHUB_DOWNLOAD" "GitHub 直连"
else
    # Gitee：直连
    download "$DOWNLOAD_URL" "Gitee 直连"
fi

if [ "$DOWNLOADED" = false ]; then
    echo "[ERROR] 所有下载源均失败，请检查网络连接"
    rm -rf "$TEMP_DIR"
    systemctl start "$SERVICE_NAME" 2>/dev/null || true
    exit 1
fi

echo ">> 解压..."
tar -xzf "$TEMP_DIR/source.tar.gz" -C "$TEMP_DIR" 2>/dev/null || {
    cd "$TEMP_DIR" && unzip -o source.tar.gz 2>/dev/null || true
}

EXTRACTED_DIR=$(find "$TEMP_DIR" -maxdepth 1 -type d \( -name "*admin*" -o -name "*Team*" \) 2>/dev/null | head -1)
[ -z "$EXTRACTED_DIR" ] || [ ! -d "$EXTRACTED_DIR" ] && EXTRACTED_DIR="$TEMP_DIR"

echo "[OK] 下载解压完成"

# ============================================
echo ""
echo "[4/6] 更新文件..."
echo "-------------------------------------------"

for BIN in ops-server ops-supervisor; do
    if [ -f "$EXTRACTED_DIR/$BIN" ]; then
        cp "$EXTRACTED_DIR/$BIN" "$INSTALL_DIR/$BIN"
        chmod +x "$INSTALL_DIR/$BIN"
        echo "[OK] $BIN 已更新"
    else
        echo "[WARN] $BIN 未找到，跳过"
    fi
done

[ -f "$EXTRACTED_DIR/index.html" ] && { cp "$EXTRACTED_DIR/index.html" "$INSTALL_DIR/index.html"; echo "[OK] index.html 已更新"; }

if [ -d "$EXTRACTED_DIR/assets" ]; then
    rm -rf "$INSTALL_DIR/assets"
    cp -r "$EXTRACTED_DIR/assets" "$INSTALL_DIR/assets"
    echo "[OK] assets/ 已更新"
fi

for SCRIPT in install.sh install-cn.sh update.sh update-cn.sh; do
    [ -f "$EXTRACTED_DIR/$SCRIPT" ] && { cp "$EXTRACTED_DIR/$SCRIPT" "$INSTALL_DIR/$SCRIPT"; chmod +x "$INSTALL_DIR/$SCRIPT"; }
done

[ ! -f "$INSTALL_DIR/.source" ] && echo "$SOURCE" > "$INSTALL_DIR/.source"

rm -rf "$TEMP_DIR"
echo "[OK] 文件更新完成"

# ============================================
echo ""
echo "[5/6] 验证完整性..."
echo "-------------------------------------------"

ERRORS=0
for f in ops-server ops-supervisor index.html; do
    if [ ! -f "$INSTALL_DIR/$f" ]; then
        echo "[ERROR] $f 缺失"
        ERRORS=$((ERRORS + 1))
    else
        echo "[OK] $f"
    fi
done

[ ! -d "$INSTALL_DIR/assets" ] && { echo "[ERROR] assets/ 缺失"; ERRORS=$((ERRORS + 1)); } || echo "[OK] assets/"

echo ""
echo "--- 数据保留 ---"
[ -f "$INSTALL_DIR/.env" ] && echo "[OK] .env 配置保留"
[ -f "$INSTALL_DIR/.initialized" ] && echo "[OK] .initialized 安装标记保留"
[ -d "$INSTALL_DIR/uploads" ] && echo "[OK] uploads/ 保留 ($(find "$INSTALL_DIR/uploads" -type f | wc -l) 个文件)"
[ -f "$INSTALL_DIR/.source" ] && echo "[OK] .source 来源标记: $(cat $INSTALL_DIR/.source)"

if [ $ERRORS -gt 0 ]; then
    echo ""
    echo "[ERROR] 验证失败，回滚中..."
    cp "$BACKUP_DIR/ops-server" "$INSTALL_DIR/" 2>/dev/null || true
    cp "$BACKUP_DIR/ops-supervisor" "$INSTALL_DIR/" 2>/dev/null || true
    cp "$BACKUP_DIR/index.html" "$INSTALL_DIR/" 2>/dev/null || true
    cp -r "$BACKUP_DIR/assets" "$INSTALL_DIR/" 2>/dev/null || true
    systemctl start "$SERVICE_NAME" 2>/dev/null || true
    echo "[OK] 已回滚到备份版本"
    exit 1
fi

# ============================================
echo ""
echo "[6/6] 启动服务..."
echo "-------------------------------------------"

systemctl start "$SERVICE_NAME"
sleep 3

if systemctl is-active --quiet "$SERVICE_NAME"; then
    echo ""
    echo "=========================================="
    echo "  更新完成！"
    echo "  来源: $SOURCE_NAME"
    echo "  备份: $BACKUP_DIR"
    echo "=========================================="
    echo ""
    echo "  回滚: systemctl stop $SERVICE_NAME"
    echo "        cp $BACKUP_DIR/ops-server $INSTALL_DIR/"
    echo "        cp -r $BACKUP_DIR/assets $INSTALL_DIR/"
    echo "        systemctl start $SERVICE_NAME"
    echo ""
    echo "  日志: journalctl -u $SERVICE_NAME -f"
    echo ""
else
    echo "[ERROR] 服务启动失败，回滚中..."
    systemctl stop "$SERVICE_NAME" 2>/dev/null || true
    cp "$BACKUP_DIR/ops-server" "$INSTALL_DIR/" 2>/dev/null || true
    cp "$BACKUP_DIR/ops-supervisor" "$INSTALL_DIR/" 2>/dev/null || true
    cp "$BACKUP_DIR/index.html" "$INSTALL_DIR/" 2>/dev/null || true
    cp -r "$BACKUP_DIR/assets" "$INSTALL_DIR/" 2>/dev/null || true
    systemctl start "$SERVICE_NAME" 2>/dev/null || true
    echo "[OK] 已回滚"
    echo "[INFO] 日志: journalctl -u $SERVICE_NAME -n 50"
    exit 1
fi
