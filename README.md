This repository was transferred from `xiaoxuan010/HFLive-BMT` to `HFLive/live-titler` on June 2, 2025.

---

# Live Titler

![](https://i.bmp.ovh/imgs/2022/02/3984a1bfd6d18100.png)

## 项目介绍

这是一套可用于OBS的标题Web页面，可以实现基本的文字显示和动画。

本项目已重构为现代化浏览器/服务器架构：
- **后端**: Go语言实现的HTTP服务器，使用SQLite3数据库
- **前端**: Vue 3 + Vite + TailwindCSS 现代化前端框架
- **通信**: REST API + WebSocket实时推送
- **部署**: 单一可执行文件，包含所有静态资源

## 快速开始

### Windows用户

1. 下载或构建 `live-titler.exe`
2. 双击运行 `live-titler.exe`
3. 浏览器访问：
   - 控制面板: http://localhost:3001/control-panel
   - 显示面板: http://localhost:3001/show-source

### 在OBS中使用

1. **添加自定义浏览器停靠窗口** (控制面板)
   - 打开OBS，左上角 视图 > 停靠部件 > 自定义浏览器停靠窗口
   - Dock名填写"Live Titler - 控制面板"
   - URL填写: `http://localhost:3001/control-panel`

2. **添加浏览器源** (显示面板)
   - 在场景中添加"浏览器"源
   - URL填写: `http://localhost:3001/show-source`
   - 宽度: 1920, 高度: 1080
   - 取消勾选"本地文件"

### 局域网访问

如果需要在局域网内其他设备访问：
- 查看服务器IP地址（例如：192.168.1.100）
- 在其他设备访问: `http://192.168.1.100:3001/control-panel`

## 功能特性

### 核心功能
- ✅ 动态Key管理（创建、删除、重命名、排序、切换类型）
- ✅ 4个独立的Key控制通道（可扩展）
- ✅ 预设管理（KEY0/1: 节目信息，KEY2/3: 歌词）
- ✅ 歌词实时显示和逐字动画
- ✅ 转场动画控制
- ✅ JSON导入/导出配置

### 新架构特性
- ✅ 浏览器/服务器架构
- ✅ 规范化数据库设计（多表：keys, programs, songs, lyrics）
- ✅ 并发控制（乐观锁机制，版本号检查）
- ✅ 多控制端同步操作
- ✅ WebSocket实时推送
- ✅ 自动重连机制（1秒间隔）
- ✅ 局域网多设备访问

## 界面导航

- **控制面板**: `http://localhost:3001/control-panel` - 主控制界面
- **Key管理**: `http://localhost:3001/control-panel/keys` - 动态管理Keys
- **显示面板**: `http://localhost:3001/show-source` - OBS浏览器源

## 数据管理

### 数据存储
所有数据存储在 `live-titler.db` SQLite数据库文件中，包括：
- 4个Key的配置信息
- 所有预设数据
- 歌词内容
- 播放状态

### 备份和恢复
1. **数据库文件备份**：复制 `live-titler.db` 文件
2. **JSON导入导出**：在控制面板点击"导入/导出设置"

## 开发指南

### 前置要求
- Go 1.20 或更高版本
- Node.js 18+ 和 pnpm
- Git

### 从源代码构建

1. **克隆仓库**
```bash
git clone https://github.com/xiaoxuan010/live-titler.git
cd live-titler
```

2. **构建前端**
```bash
cd frontend
pnpm install
pnpm build
cd ..
```

3. **下载Go依赖**
```bash
go mod tidy
```

4. **构建可执行文件**

Windows:
```bash
go build -o live-titler.exe main.go
```

或使用批处理文件:
```bash
build-windows.bat
```

Linux/Mac:
```bash
./build.sh
```

### 前端开发

在开发模式下运行前端（带热重载）：

```bash
cd frontend
pnpm dev
```

前端开发服务器会运行在 http://localhost:5173，并自动代理API请求到后端（需要后端同时运行）。

### 后端开发

运行开发服务器：
```bash
go run main.go
```

### 技术栈

**后端:**
- Go + gorilla/mux + gorilla/websocket + GORM + SQLite3

**前端:**
- Vue 3 + Vite
- TailwindCSS
- TypeScript
- vue-router
- lyrics.js

### 目录结构
```
live-titler/
├── main.go              # Go服务器主程序
├── go.mod               # Go模块定义
├── live-titler.db       # SQLite数据库（运行时生成）
├── frontend/            # 前端项目目录
│   ├── src/
│   │   ├── views/       # 页面组件
│   │   ├── components/  # UI组件
│   │   ├── api/         # API调用
│   │   └── types.ts     # TypeScript类型定义
│   ├── dist/            # 构建输出（嵌入到Go二进制）
│   └── package.json
└── build-windows.bat    # Windows构建脚本
```

## API文档

### REST API
- `GET /api/presets` - 获取所有预设
- `PUT /api/presets/{keyNum}` - 更新指定Key的预设
- `POST /api/presets/import` - 导入JSON配置
- `GET /api/presets/export` - 导出JSON配置

### WebSocket
- `WS /ws` - 实时状态推送

## 错误处理

### 控制面板
当服务器连接断开时：
- 页面顶部显示红色提示条
- 所有控制按钮和输入框被禁用
- 自动尝试重新连接（每1秒重试一次）

### 显示面板
当服务器连接断开时：
- 保持当前显示内容不变
- 不显示任何错误提示（避免干扰直播）
- 自动尝试重新连接（每1秒重试一次）

## 常见问题

**Q: 端口3001被占用怎么办？**
A: 目前端口是固定的，需要停止占用3001端口的其他程序。

**Q: 如何重置所有配置？**
A: 删除 `live-titler.db` 文件，重启服务器即可恢复默认配置。

**Q: 可以同时有多个控制端吗？**
A: 可以！所有控制端的操作会通过WebSocket实时同步到所有显示端。

## 安装字体

作者测试时使用的字体是：方正灵飞经小楷 简、方正拉勾标题体 简（均可在方正官网免费下载）。安装字体后即可默认使用上述两款字体显示，否则会用浏览器默认字体。有需要可以到`bmt-css/source-style.css`里面自己设置。

## 开源相关

本项目完全开源，开源地址：https://github.com/HFLive/live-titler ，开源协议见LICENSE文件

本项目使用的开源代码有：
1. MDUI( https://github.com/zdhxiong/mdui )，其基于MIT协议
