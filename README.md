# 运维管理平台

成都商惠计算机系统有限公司 - 运维人员管理平台

## 一键安装

**海外服务器：**
```bash
curl -fsSL https://raw.githubusercontent.com/Mcloud136/admin/main/install.sh | sudo bash
```

**国内服务器（阿里云镜像 + Gitee 下载）：**
```bash
curl -fsSL https://gitee.com/wxbns/Team-Management/raw/main/install-cn.sh | sudo bash
```

安装完成后访问 `http://服务器IP` 进入安装向导，按提示完成数据库、管理员、品牌信息配置即可。

## 功能模块

| 模块 | 功能 |
|------|------|
| 工单管理 | 创建/派单/处理/完单报告/验收/归档，支持文件上传 |
| 项目管理 | 创建/成员管理/整改流程/验收，关联工单 |
| 工程师管理 | 用户/团队/权限/重置密码 |
| 知识库 | 富文本编辑/文档上传解析/图片粘贴缩放/文件预览下载 |
| 资产管理 | IT资产登记/关联工单/生命周期管理 |
| 自动评分 | 多维度绩效评估（响应/效率/SLA/质量/知识贡献） |
| 系统设置 | SLA配置/评分权重/通知规则/品牌定制（Logo+背景） |

## 技术栈

| 组件 | 技术 |
|------|------|
| 前端 | Vue 3 + Arco Design + TipTap + ECharts + Lucide Icons |
| 后端 | Go + Gin + sqlx |
| 数据库 | PostgreSQL |
| 部署 | Nginx + systemd（守护进程自动重启） |

## 文件说明

| 文件 | 说明 |
|------|------|
| `index.html` | 前端入口 |
| `assets/` | 前端 JS/CSS 资源 |
| `ops-server` | 后端服务（Linux AMD64） |
| `ops-supervisor` | 守护进程（监控并自动重启后端） |
| `install.sh` | 一键安装脚本（海外源） |
| `install-cn.sh` | 一键安装脚本（国内镜像） |
| `.env.example` | 环境配置模板 |

## 安装脚本说明

两个脚本功能相同，区别在于下载源：

| 脚本 | apt/yum 源 | 项目文件源 | 适用场景 |
|------|-----------|-----------|---------|
| `install.sh` | 系统默认 | GitHub | 海外服务器 |
| `install-cn.sh` | 阿里云镜像 | Gitee | 国内服务器 |

脚本自动完成：
1. 安装 Nginx + PostgreSQL
2. 下载项目文件到当前目录
3. 创建数据库
4. 配置 systemd 服务（守护进程模式）
5. 配置 Nginx 反向代理
6. 启动服务

## 手动部署

如果不使用安装脚本：

1. 安装 Nginx 和 PostgreSQL
2. 将 `index.html` 和 `assets/` 放到 Nginx 网站目录
3. 运行 `./ops-supervisor`（自动启动 ops-server）
4. 配置 Nginx 反向代理 `/api/` → `http://127.0.0.1:1365`
5. 访问网站，安装向导自动引导完成配置

## License

MIT
