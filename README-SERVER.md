# Live Titler - 浏览器/服务器架构

## 架构说明

本项目已重构为浏览器/服务器架构：
- **后端**: Go语言实现的HTTP服务器，使用SQLite3数据库
- **前端**: 保留原有的HTML/CSS/JavaScript前端页面
- **通信**: REST API + WebSocket实时推送

## 快速开始

### Windows用户

1. 双击 `live-titler.exe` 启动服务器
2. 浏览器访问:
   - 控制面板: http://localhost:8080/control-pannel.html
   - 显示面板: http://localhost:8080/show-source.html
3. 在OBS中使用浏览器源，URL设置为上述地址

### 构建

#### Windows
```bash
go build -o live-titler.exe main.go
```

或使用批处理文件:
```bash
build-windows.bat
```

#### Linux/Mac
```bash
go build -o live-titler main.go
```

或使用脚本:
```bash
./build.sh
```

## 功能特性

- ✅ 保留所有原有功能
- ✅ 4个Key的预设管理
- ✅ 歌词管理和实时显示
- ✅ JSON导入/导出
- ✅ 实时同步（控制面板 → 服务器 → 显示面板）
- ✅ SQLite3数据持久化
- ✅ 支持多个控制端同时操作
- ✅ 局域网访问（无需认证）

## API端点

- `GET /api/presets` - 获取所有预设
- `PUT /api/presets/{keyNum}` - 更新指定Key的预设
- `POST /api/presets/import` - 导入JSON配置
- `GET /api/presets/export` - 导出JSON配置
- `WS /ws` - WebSocket连接，用于实时推送

## 数据库

数据存储在 `live-titler.db` SQLite数据库文件中，包含：
- 4个Key的配置信息
- 所有预设数据
- 歌词内容
- 播放状态

## 技术栈

- **后端**: Go 1.24+
  - gorilla/mux - HTTP路由
  - gorilla/websocket - WebSocket支持
  - gorm - ORM框架
  - sqlite - 数据库驱动
- **前端**: 原有的HTML/CSS/JavaScript
  - MDUI - UI框架
  - Fetch API - REST调用
  - WebSocket - 实时通信

## 部署说明

1. 确保服务器可执行文件与静态文件在同一目录
2. 运行 `live-titler.exe`（Windows）或 `./live-titler`（Linux/Mac）
3. 服务器默认监听 8080 端口
4. 局域网内其他设备可通过服务器IP访问

## 开发

### 依赖安装
```bash
go get github.com/gorilla/mux
go get github.com/gorilla/websocket
go get gorm.io/driver/sqlite
go get gorm.io/gorm
```

### 运行开发服务器
```bash
go run main.go
```

## 注意事项

- 本系统设计用于局域网环境，无身份认证
- 数据库文件 `live-titler.db` 包含所有配置，请定期备份
- 支持JSON导入导出功能，可用于数据迁移和备份
- 多个控制端可同时操作，通过WebSocket实现实时同步

## 原项目信息

原作者: HFLive13.0 xiaoxuan010
开源地址: https://github.com/xiaoxuan010/HFLive-BMT
开源协议: GPL
