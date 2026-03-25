# 中国象棋在线对弈

基于 Vue 3 + Gin + WebSocket 的实时双人中国象棋对弈网站。

## 技术栈

| 层级 | 技术 |
|---|---|
| 前端 | Vue 3 · TypeScript · UnoCSS · Pinia · xiangqiboardjs |
| 后端 | Go · Gin · gorilla/websocket |
| 数据库 | MySQL 8 · Redis 6 |
| 部署 | Docker · Docker Compose · GitHub Actions |

## 功能特性

- 创建 / 加入房间，4 位随机房间号一键分享
- WebSocket 实时同步对局，毫秒级落子响应
- 完整象棋规则校验（将军、困毙、白脸将、蹩马腿、拦象眼）
- 支持认输，胜负自动判定
- 走棋记录面板，棋盘可翻转（红方 / 黑方视角）
- 前端通过 GitHub Actions 自动部署到 GitHub Pages

---

## 快速开始

### 方式一：本地开发

**环境要求**

- Go 1.21+
- Node.js 20+ / pnpm 10+
- MySQL 8
- Redis 6

**1. 克隆仓库**

```bash
git clone <你的仓库地址>
cd jd
```

**2. 启动后端**

```bash
cd jd_backend

# 安装 Go 依赖
go mod tidy

# 新建 .env 文件，填入数据库和 Redis 配置
cp .env.example .env   # 或手动创建，见下方说明

# 启动
go run .
```

`.env` 文件内容：

```env
DB_USERNAME=root
DB_PASSWORD=你的数据库密码
DB_HOST=localhost
DB_PORT=3306
DB_NAME=jd
DB_LOG_LEVEL=silent
REDIS_HOST=localhost
REDIS_PORT=6379
FRONTEND_URL=http://localhost:5173
GIN_MODE=debug
```

**3. 启动前端**

```bash
cd jd_frontend

# 安装依赖
pnpm install

# 启动开发服务器（默认 http://localhost:5173）
pnpm dev
```

**4. 访问象棋页面**

打开浏览器访问 `http://localhost:5173/chess`

---

### 方式二：Docker Compose 一键启动（推荐）

**环境要求**

- Docker 24+
- Docker Compose v2

**1. 在项目根目录新建 `.env` 文件**

```env
# 前端访问地址（用于后端跨域白名单）
FRONTEND_URL=http://localhost:5173
```

**2. 构建并启动所有服务**

```bash
docker compose up --build
```

服务启动后：

| 服务 | 地址 |
|---|---|
| 前端 | http://localhost:5173 |
| 后端 API | http://localhost:8080 |
| MySQL | localhost:3307 |
| Redis | localhost:6379 |

**3. 访问象棋**

打开 `http://localhost:5173/chess`

**停止服务**

```bash
docker compose down
```

**清理数据库数据（谨慎）**

```bash
docker compose down -v
```

---

### 方式三：GitHub Pages 部署前端

前端支持通过 GitHub Actions 自动部署到 GitHub Pages，后端需自行部署到服务器（VPS / 云函数等）。

**步骤**

1. Fork / Push 本仓库到 GitHub

2. 在仓库 **Settings → Pages** 中，将 Source 设为 `GitHub Actions`

3. 在仓库 **Settings → Secrets and variables → Actions** 中添加：

   | 名称 | 说明 |
   |---|---|
   | `VITE_WS_URL` | 后端 WebSocket 地址，例如 `wss://your-server.com/ws/chess` |

4. 如果仓库不是 `用户名.github.io` 这种根路径仓库，还需在 **Variables** 中添加：

   | 名称 | 值 |
   |---|---|
   | `VITE_BASE_URL` | `/仓库名/`，例如 `/jd/` |

5. Push 代码到 `main` 分支，Actions 自动构建并部署

6. 访问 `https://<你的用户名>.github.io/<仓库名>/chess`

---

## 游戏使用教程

### 开始一局游戏

1. 打开象棋页面（`/chess`），进入大厅

2. **创建方**：点击「创建房间」，获得 4 位房间号（如 `A3K7`）

3. **加入方**：将房间号分享给对手，对手在「加入对局」输入框中输入房间号后点击加入

4. 双方都进入房间后，游戏自动开始，**红方先走**

### 走棋规则

- 拖拽棋子到目标位置即可落子
- 非法走法会被服务端拒绝，棋子自动弹回原位
- 页面顶部实时显示当前回合方
- 右侧面板记录每一步走棋历史

### 结束游戏

- **将死 / 困毙**：一方无合法走法时，游戏自动判负，弹出结果弹层
- **认输**：点击右上角「认输」按钮，对手自动获胜
- 结果弹层出现后，点击「返回大厅」可重新开局

---

## 项目结构

```
jd/
├── docker-compose.yml          # 一键启动所有服务
├── nginx.conf                  # Nginx 反向代理（API + WebSocket）
├── .github/
│   └── workflows/
│       └── deploy.yml          # GitHub Actions 自动部署
│
├── jd_backend/                 # Go 后端
│   ├── main.go
│   ├── config/                 # 配置加载
│   ├── controller/
│   │   └── chess.go            # 象棋 WebSocket 控制器
│   ├── service/
│   │   └── chess/
│   │       ├── engine.go       # 象棋规则引擎
│   │       └── room.go         # 房间管理器
│   ├── route/route.go          # 路由注册
│   └── Dockerfile
│
└── jd_frontend/                # Vue 3 前端
    ├── public/
    │   └── xiangqi/            # xiangqiboardjs 本地资源
    │       ├── xiangqiboard.min.js
    │       ├── xiangqiboard.min.css
    │       └── img/xiangqipieces/wikimedia/  # 14 个棋子 SVG
    ├── src/
    │   ├── stores/chessStore.ts # 游戏状态 + WebSocket + 坐标转换
    │   └── views/chess/
    │       ├── LobbyView.vue   # 大厅页
    │       └── GameView.vue    # 对弈页
    └── Dockerfile
```

---

## 常见问题

**Q：棋子拖动后弹回，提示「非法走法」？**

服务端进行了完整的象棋规则校验（包括走后是否被将军）。请确认走法符合规则。

**Q：Docker 启动时后端报「数据库连接失败」？**

MySQL 容器启动需要数秒初始化，docker-compose 已配置健康检查等待，若仍失败请稍等片刻后重试：

```bash
docker compose restart backend
```

**Q：GitHub Pages 部署后打开是空白页？**

检查 `VITE_BASE_URL` 是否正确设置为 `/仓库名/`。

**Q：本地开发时 WebSocket 连接失败？**

确认后端已启动（`go run .`），且 `jd_frontend/.env.development` 中的 `VITE_WS_URL=ws://localhost:8080/ws/chess` 配置正确。
