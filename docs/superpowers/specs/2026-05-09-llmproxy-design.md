# LLM Scout — 设计文档

项目名称：**LLM Scout**
GitHub 仓库: `llmscout/llmscout`（待创建）

## 概述

在 Wails v2 桌面应用基础上，开发一个 LLM 代理调试分析工具，取名 **LLM Scout**（LLM 侦察兵/侦探）。通过路径映射将请求转发到不同 LLM 提供商，不修改请求内容，完整记录请求/响应报文供后续调试分析。

### 开源规范

本项目将作为开源项目发布，所有提交必须遵循以下规范：
- **Conventional Commits** 格式提交信息（`feat:`、`fix:`、`chore:`、`docs:`、`refactor:` 等）
- 所有代码须附带 LICENSE（Apache 2.0 已有）
- 合理的 PR 拆分，每个提交具有完整语义
- 禁止在代码中硬编码任何 API Key 或敏感信息

## 架构

```
┌─────────────────────────────────────────────────┐
│                    Vue 前端                       │
│          (侧边栏: 代理/路由/日志/设置)             │
├─────────────────────────────────────────────────┤
│           Wails Bind (JS ↔ Go 桥接)              │
├─────────────────────────────────────────────────┤
│  App (门面层) — 绑定所有方法，生命周期管理          │
├─────────────────────────────────────────────────┤
│  ┌──────────┐ ┌─────────┐ ┌───────────────┐     │
│  │ proxy    │ │ route   │ │ log           │     │
│  │ ProxyEngine│ │Mgr    │ │ LogService    │     │
│  │ (HTTP转发,  │(路由匹配,│ │ (记录,查询,    │     │
│  │ SSE流式)  │ CRUD)   │ │  搜索/分页)     │     │
│  └─────┬─────┘ └────┬────┘ └───────┬───────┘     │
│        │            │              │              │
│  ┌─────┴────────────┴──────────────┴───────┐     │
│  │           SQLite (routes + logs)          │     │
│  └──────────────────────────────────────────┘     │
└─────────────────────────────────────────────────┘
```

## 后端组件

### 1. RouteManager — 路由管理

**数据模型：**

```
type RouteRule struct {
    ID        int64
    Name      string           // 显示名称，如 "openai"
    Type      string           // "prefix" | "exact"
    Path      string           // 代理路径，如 /openai
    TargetURL string           // 目标域名或完整URL
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

**匹配逻辑：**
- exact 规则优先于 prefix 规则
- 按添加顺序匹配，返回第一条命中的规则
- 匹配后根据规则类型拼接目标URL：
  - prefix：剥离 Path 前缀，剩余路径拼接到 TargetURL
  - exact：直接完整映射到 TargetURL
- 不命中任何规则时返回 502 Bad Gateway

**对外方法（通过 Wails Bind 暴露）：**
- `ListRoutes() []RouteRule`
- `AddRoute(rule) error`
- `UpdateRoute(id, rule) error`
- `DeleteRoute(id) error`
- `GetRoute(id) RouteRule`

### 2. ProxyEngine — 代理引擎

**核心职责：**
- 在独立 goroutine 中运行 `http.Server`
- 监听用户配置的端口（默认 8899）
- 接收到请求 → 调用 RouteManager.Match(path) → 转发到目标URL
- 不修改 HTTP 方法和请求体
- 保留全部原始请求头，并注入 X-Forwarded-For

**流式响应 (SSE) 处理：**
- 检测响应头 `Content-Type: text/event-stream`
- 逐行读取 `data:` 事件，缓存到临时 buffer
- 边接收边转发给客户端（不等待全部完成）
- 请求完成后将完整的事件序列写入 LogService

**对外方法：**
- `Start() error`
- `Stop() error`
- `Status() ProxyStatus` (端口、运行时长、是否运行中)

### 3. LogService — 日志服务

**数据模型：**

```
type RequestLog struct {
    ID          int64
    RouteName   string       // 路由名称，如 "openai"
    Method      string       // HTTP method
    Path        string       // 完整请求路径
    Protocol    string       // "REST" | "SSE"
    StatusCode  int
    Latency     int64        // 毫秒
    ReqHeaders  string       // JSON
    ReqBody     string
    RespHeaders string       // JSON
    RespBody    string
    CreatedAt   time.Time
}
```

**写入流程：**
- 代理转发完成后通过 channel 异步写入
- channel buffer 100，满了时丢弃日志（不影响代理）
- SQLite INSERT，WAL 模式保证写入性能

**查询方法：**
- `QueryLogs(filter LogFilter) ([]RequestLog, int64)` — 返回数据和总数
- LogFilter: keyword（搜索req/body）、routeName、statusCode、protocol、startTime、endTime、page、pageSize
- 默认按时间倒序

**对外方法：**
- `ListLogs(filter, page, pageSize) → {list, total}`
- `GetLog(id) *RequestLog`
- `ClearLogs()`
- `SetAutoRefresh(interval int)` — 无请求时前端轮询间隔

### 4. SQLite 仓储

- 表 routes：存储路由规则
- 表 logs：存储请求日志
- 使用 `modernc.org/sqlite`（纯 Go 实现，无需 CGO）
- WAL 模式开启
- 数据库文件默认存放在应用数据目录，可在设置中配置

### 5. 设置

```
type Settings struct {
    Port       int    // 代理端口，默认 8899
    DbPath     string // 数据库路径
}
```

- 保存在 SQLite 的 settings 表中（key-value）
- 前端设置页面可修改

## 前端组件

### 布局 — 可收缩侧边栏

- 侧边栏 4 个菜单项：📡 代理 / 🔀 路由 / 📋 日志 / ⚙ 设置
- 底部「« 收缩」按钮，收缩后仅显示图标（52px 宽度）
- 使用 NaiveUI 的 `n-layout` + `n-layout-sider`

### 📡 代理页面

```
┌─────────────────────────────┐
│ 端口: [  8899  ]            │
│ 状态: ● 运行中  [停止代理]    │
├─────────────────────────────┤
│ 已配置路由: 3 条             │
│ 今日请求: 128                │
│ 运行时长: 02:35:42           │
└─────────────────────────────┘
```

### 🔀 路由页面

```
┌────────────────────────────────────┐
│ 路由规则                 [+ 添加]   │
├────────────────────────────────────┤
│ [prefix]  /openai  → api.openai.com  [编辑] [删除] │
│ [prefix]  /anthropic → api.anthropic.com [编辑] [删除] │
│ [exact]   /deepseek-chat → api.deepseek.com/v1/chat/... [编辑] [删除] │
└────────────────────────────────────┘
```

添加/编辑使用弹窗表单（NaiveUI `n-modal`）：
- 名称（显示标签）
- 类型（prefix / exact）
- 代理路径
- 目标域名/URL

### 📋 日志页面

```
┌──────────────────────────────────────────────────────────────┐
│ [🔍 搜索...] [全部服务商▼] [全部状态▼] [全部协议▼] [时间范围▼] │
│                                               🔄 [3 秒▼]    │
├──┬────┬───┬──────┬────────────────────┬─────┬──────────────┤
│协议│方法│状态│服务商│ 路径              │耗时 │ 时间          │
├──┼────┼───┼──────┼────────────────────┼─────┼──────────────┤
│SSE│POST│200│openai│ /v1/chat/completions│4.2s │05-09 14:32   │
│REST│POST│401│anthropic│ /v1/messages    │0.3s │05-09 14:31   │
├──┴────┴───┴──────┴────────────────────┴─────┴──────────────┤
│ 每页 [20▼] 条     « ‹ 1 2 3 ... 7 › »                     │
├──────────────────────────────────────────────────────────────┤
│ ▸ 展开详情：请求头 / 请求体 / 响应体（Tab切换+JSON格式化）     │
│    [📋 复制全部]                                             │
│    {                                                        │
│      "model": "deepseek-chat",                              │
│      "messages": [...],                                     │
│      "stream": true                                         │
│    }                                                        │
└──────────────────────────────────────────────────────────────┘
```

**说明：**
- 支持关键词全局搜索（全文检索请求体和响应体）
- 下拉筛选：服务商 / 状态码 / 协议类型(REST/SSE) / 时间范围
- 表格点击行展开详情，Tab 切换查看请求头/请求体/响应体
- JSON 自动格式化并语法高亮
- 自动刷新开关（开/关 + 间隔配置）

### ⚙ 设置页面

- 代理端口设置（输入框，保存后下次启动生效或重启代理）
- 自动刷新默认间隔
- 数据库路径显示
- 清空日志按钮

## 数据流

### 请求代理流程

```
1. 客户端发送 POST /openai/v1/chat/completions → localhost:8899
2. ProxyEngine 接收请求，解析路径 /openai/v1/chat/completions
3. RouteManager.Match("/openai/v1/chat/completions")
   → 命中 prefix 规则 /openai → api.openai.com
   → 目标URL: https://api.openai.com/v1/chat/completions
4. ProxyEngine 创建反向代理请求，复制原始 headers + body
5. 发送请求到 api.openai.com
6. 检测响应 Content-Type:
   a) 非流式: 等待完整响应 → 转发给客户端 → 写入 LogService
   b) SSE: 逐行读取 → 逐行转发 → 完成后写入 LogService（含完整事件序列）
7. LogService 通过 channel 异步写入 SQLite
8. 前端日志列表自动刷新（如开启）
```

### 前端查询流程

```
1. 用户打开日志页面，输入筛选条件
2. Wails Bridge 调用 LogService.QueryLogs(filter)
3. LogService 构造 SQL 查询 + 分页
4. 返回 []RequestLog + total count
5. Vue 渲染表格，展开详情时格式化 JSON
```

## 技术栈

- Go 标准库 `net/http/httputil` 反向代理
- `modernc.org/sqlite` 纯 Go SQLite 驱动
- Vue 3 `<script setup>` + NaiveUI 组件库
- Wails v2 Bind 机制桥接前后端

## 项目初始化（从模板到真实项目）

开始实施前的准备工作：
1. `go.mod` 模块路径改为 `github.com/llmscout/llmscout`
2. `wails.json` 的 name/outputfilename 改为 `llmscout`
3. `main.go` 窗口标题改为 `"LLM Scout"`
4. 清理模板不需要的文件：
   - 删除 `frontend/src/components/HelloWorld.vue`
   - 删除 `frontend/src/assets/images/` 下的模板 logo（logo-universal.png、naive-logo.svg）
   - 删除 `frontend/src/style.css`（后续按需重写）
   - 删除 `README.md` 模板内容（后续重写）
   - 删除 `wails-naive.png`
5. 简化 `App.vue`：去掉网格布局和 logo 展示，改为侧边栏基本结构
6. 初始化 git 仓库并创建初始 commit

## 项目文件结构（新增/修改）

```
main.go                     — 新增 App 初始化 + Bind 注册
app.go                      — 扩展 App struct，集成所有服务
go.mod                      — 模块改为 github.com/llmscout/llmscout，新增 sqlite 依赖

internal/
  route/
    manager.go              — RouteManager (CRUD + 匹配)
    model.go                — RouteRule struct
  proxy/
    engine.go               — ProxyEngine (HTTP 服务器 + 转发)
  log/
    service.go              — LogService (记录 + 查询)
    model.go                — RequestLog struct
  storage/
    db.go                   — SQLite 初始化 + 迁移
    routes_repo.go          — 路由表 CRUD
    logs_repo.go            — 日志表 CRUD
    settings_repo.go        — 设置存取

frontend/src/
  App.vue                   — 改造为侧边栏布局
  components/
    ProxyPanel.vue           — 代理控制页面
    RoutePanel.vue           — 路由配置页面
    LogViewer.vue            — 日志查看页面
    SettingsPanel.vue        — 设置页面
    JsonViewer.vue           — JSON 格式化展示组件
  wailsjs/go/main/App.js    — 自动生成
```

## 验证方案

1. `wails dev` 启动应用
2. 在路由页面添加一条 openai 路由：prefix /openai → api.openai.com
3. 代理页面启动代理（端口 8899）
4. 用 curl 测试转发：`curl http://localhost:8899/openai/v1/chat/completions -H "Authorization: Bearer sk-xxx" -d '{"model":"gpt-4","messages":[{"role":"user","content":"hi"}]}'`
5. 验证日志页面显示请求记录，展开查看格式化的请求/响应体
6. 测试 SSE 转发：添加 stream:true 参数，验证协议列显示 SSE
7. 测试筛选、搜索、分页、自动刷新功能
