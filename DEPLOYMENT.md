# Live Titler 部署和使用指南

## 快速开始

### Windows用户（推荐）

1. **下载发布版本**
   - 从Releases页面下载 `live-titler.exe`
   
2. **启动服务器**
   - 双击 `live-titler.exe` 启动服务器
   - 服务器将在端口 8080 上启动
   - 控制台会显示访问地址

3. **访问界面**
   - 控制面板: http://localhost:8080/control-pannel.html
   - 显示面板: http://localhost:8080/show-source.html

4. **OBS集成**
   - 在OBS中添加"浏览器"源
   - URL设置为显示面板地址
   - 宽度: 1920, 高度: 1080
   - 在OBS中添加"自定义浏览器停靠窗口"
   - URL设置为控制面板地址

### 局域网访问

如果需要在局域网内其他设备访问：

1. 查看服务器IP地址（例如：192.168.1.100）
2. 在其他设备上访问：
   - 控制面板: http://192.168.1.100:8080/control-pannel.html
   - 显示面板: http://192.168.1.100:8080/show-source.html

## 系统架构

```
┌─────────────┐         REST API          ┌──────────────┐
│  控制面板    │◄─────────────────────────►│              │
│ (Browser)   │                            │   Go Server  │
└─────────────┘         WebSocket          │   + SQLite   │
                 ┌──────────────────────────┤              │
                 │                          └──────────────┘
                 ▼
         ┌─────────────┐
         │  显示面板    │
         │ (Browser)   │
         └─────────────┘
```

## 功能特性

### 核心功能
- ✅ 4个独立的Key控制通道
- ✅ 预设管理（节目名、表演者）
- ✅ 歌词管理和实时显示
- ✅ 逐字动画效果
- ✅ JSON导入/导出配置

### 新增功能
- ✅ 浏览器/服务器架构
- ✅ SQLite数据持久化
- ✅ 多控制端同步操作
- ✅ WebSocket实时推送
- ✅ 自动重连机制
- ✅ 错误处理和状态提示

### 网络功能
- ✅ 局域网多设备访问
- ✅ 实时状态同步
- ✅ 断线自动重连
- ✅ 无需身份认证（局域网环境）

## 错误处理

### 控制面板
当服务器连接断开时：
- 页面顶部显示红色提示条："服务器连接已断开，正在重试..."
- 所有控制按钮和输入框被禁用
- 自动尝试重新连接（每1秒重试一次）
- 错误信息记录在浏览器控制台

### 显示面板
当服务器连接断开时：
- 保持当前显示内容不变
- 不显示任何错误提示（避免干扰直播）
- 自动尝试重新连接（每1秒重试一次）
- 错误信息记录在浏览器控制台

## 数据管理

### 数据存储
所有数据存储在 `live-titler.db` 文件中，包括：
- 4个Key的配置
- 所有预设信息
- 歌词内容
- 播放状态

### 备份和恢复

#### 方法1：数据库文件备份
```bash
# 备份
copy live-titler.db live-titler.db.backup

# 恢复
copy live-titler.db.backup live-titler.db
```

#### 方法2：JSON导入导出
1. 在控制面板点击"导入/导出设置"
2. 复制JSON配置到剪贴板
3. 保存到文件作为备份
4. 需要恢复时，粘贴JSON并点击保存

## 构建说明

### 从源代码构建

#### 前置要求
- Go 1.20 或更高版本
- Git

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
chmod +x build.sh
./build.sh
```

### 依赖安装
```bash
go get github.com/gorilla/mux
go get github.com/gorilla/websocket
go get gorm.io/driver/sqlite
go get gorm.io/gorm
```

## API文档

### REST API端点

#### 获取所有预设
```
GET /api/presets
Response: JSON数组，包含所有4个Key的配置
```

#### 更新单个预设
```
PUT /api/presets/{keyNum}
Content-Type: application/json
Body: Preset对象
Response: {"status": "success"}
```

#### 导入配置
```
POST /api/presets/import
Content-Type: application/json
Body: Preset数组
Response: {"status": "success"}
```

#### 导出配置
```
GET /api/presets/export
Response: JSON数组，包含所有配置
```

### WebSocket端点

```
WS /ws
接收: 完整的preset数组（每次状态更新）
```

## 常见问题

### Q: 端口8080被占用怎么办？
A: 目前端口是固定的，需要停止占用8080端口的其他程序。

### Q: 如何查看服务器日志？
A: 所有日志输出到控制台，建议通过命令行启动以查看日志。

### Q: 数据会丢失吗？
A: 所有数据实时保存到SQLite数据库，不会丢失。建议定期备份 `live-titler.db` 文件。

### Q: 可以同时有多个控制端吗？
A: 可以！所有控制端的操作会通过WebSocket实时同步到所有显示端。

### Q: 如何重置所有配置？
A: 删除 `live-titler.db` 文件，重启服务器即可恢复默认配置。

### Q: 显示面板出现错误提示怎么办？
A: 显示面板不应该显示任何错误提示。如果出现，请检查浏览器控制台日志。

## 技术栈

### 后端
- **Go 1.20+** - 高性能服务器
- **gorilla/mux** - HTTP路由
- **gorilla/websocket** - WebSocket支持
- **GORM** - ORM框架
- **SQLite3** - 嵌入式数据库

### 前端
- **HTML5/CSS3/JavaScript** - 原有前端代码
- **MDUI** - Material Design UI框架
- **Fetch API** - REST调用
- **WebSocket API** - 实时通信

## 开发

### 运行开发服务器
```bash
go run main.go
```

### 目录结构
```
live-titler/
├── main.go              # Go服务器主程序
├── go.mod               # Go模块定义
├── live-titler.db       # SQLite数据库（运行时生成）
├── control-pannel.html  # 控制面板
├── show-source.html     # 显示面板
├── bmt-js/             # JavaScript脚本
│   ├── control-script.js
│   ├── source-script.js
│   └── default-preset.js
├── bmt-css/            # 样式文件
├── mdui/               # MDUI框架
└── js-lyrics/          # 歌词处理库（子模块）
```

## 贡献

欢迎提交Issue和Pull Request！

## 许可证

GPL License - 详见 LICENSE 文件

## 原项目信息

- 原作者: HFLive13.0 xiaoxuan010
- 原项目: https://github.com/xiaoxuan010/HFLive-BMT
- 本项目: https://github.com/xiaoxuan010/live-titler

## 版本历史

### v2.0.0 (当前版本)
- ✅ 重构为浏览器/服务器架构
- ✅ 添加Go后端和SQLite数据库
- ✅ 实现WebSocket实时同步
- ✅ 添加错误处理和自动重连
- ✅ 支持多控制端同步操作

### v1.0.0 (原版本)
- 纯前端HTML/CSS/JavaScript实现
- 使用localStorage存储数据
- 通过localStorage事件同步
