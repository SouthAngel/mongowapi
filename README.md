# MongoWAPI

基于 **Go + Gin + MongoDB 官方驱动** 的 MongoDB RESTful Web API，通过 HTTP 接口操作 MongoDB，支持文档 CRUD、索引管理，可注册为 Windows / Linux 系统服务。

## 功能特性

- 文档增删改查（单条 / 批量）
- 查询支持 filter、sort、projection、分页（skip/limit）
- 更新支持 upsert
- 索引的创建、列出、删除
- YAML 配置文件，支持连接超时与请求超时
- 统一的 JSON 响应格式
- 支持 Windows 服务（SCM）与 Linux 服务（systemd）注册/移除
- 一键构建 Windows + Linux 双平台 release 二进制

## 项目结构

```
mongowapi/
├── main.go                 # 入口：命令行参数、服务管理、路由注册
├── build.bat               # 一键构建双平台 release 二进制
├── config.yaml             # 配置文件
├── config/config.go        # 配置加载与默认值
├── models/models.go        # 请求/响应结构体
├── database/mongo.go       # MongoDB 连接管理
├── handlers/
│   ├── response.go          # 统一响应格式
│   ├── document.go          # 文档 CRUD 处理
│   └── index.go             # 索引管理处理
├── service_windows.go       # Windows 服务实现（SCM）
├── service_linux.go        # Linux 服务实现（systemd）
└── service_other.go        # 其他平台回退实现
```

## 快速开始

### 前置要求

- Go 1.21+
- MongoDB 3.6+

### 安装与运行

```bash
# 安装依赖
go mod tidy

# 运行（默认读取 config.yaml）
go run .

# 指定配置文件
go run . -config /path/to/config.yaml
```

### 构建 Release 二进制

```bash
# Windows 下执行 build.bat，一键构建双平台
.\build.bat
```

构建产物输出到 `release/` 目录：

| 文件 | 平台 | 大小 |
|------|------|------|
| `mongowapi-windows-amd64.exe` | Windows amd64 | ~24 MB |
| `mongowapi-linux-amd64` | Linux amd64 | ~24 MB |

## 命令行参数

| 参数 | 说明 |
|------|------|
| `-config` | 配置文件路径（默认 `config.yaml`） |
| `-install` | 注册为系统服务 |
| `-uninstall` | 移除系统服务 |
| `-start` | 启动系统服务 |
| `-stop` | 停止系统服务 |

不加服务参数时，以前台模式直接运行。

## 系统服务

### Windows（SCM 服务）

需以管理员权限运行：

```powershell
# 注册服务
.\mongowapi-windows-amd64.exe -install

# 启动服务
.\mongowapi-windows-amd64.exe -start

# 停止服务
.\mongowapi-windows-amd64.exe -stop

# 移除服务
.\mongowapi-windows-amd64.exe -uninstall
```

注册的服务名 `MongoWAPI`，启动类型为自动启动。也可通过 `services.msc` 管理。

### Linux（systemd 服务）

需以 root 权限运行：

```bash
# 注册服务（生成 /etc/systemd/system/mongowapi.service）
sudo ./mongowapi-linux-amd64 -install

# 启动服务
sudo ./mongowapi-linux-amd64 -start

# 停止服务
sudo ./mongowapi-linux-amd64 -stop

# 移除服务
sudo ./mongowapi-linux-amd64 -uninstall
```

生成的 systemd unit 配置为开机自启、异常自动重启。也可通过 `systemctl` 管理：

```bash
sudo systemctl start mongowapi
sudo systemctl stop mongowapi
sudo systemctl status mongowapi
```

## 配置说明

编辑 [config.yaml](config.yaml)：

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  mode: "debug"          # debug / release

mongo:
  uri: "mongodb://localhost:27017"
  connect_timeout: 10s
  request_timeout: 30s
  white_list_dbs: [myapp, myapp2]   # 数据库访问白名单
```

| 字段 | 说明 | 默认值 |
|------|------|--------|
| `server.host` | 监听地址 | `0.0.0.0` |
| `server.port` | 监听端口 | `8080` |
| `server.mode` | Gin 运行模式 | `debug` |
| `mongo.uri` | MongoDB 连接串 | `mongodb://localhost:27017` |
| `mongo.connect_timeout` | 连接超时 | `10s` |
| `mongo.request_timeout` | 请求超时 | `30s` |
| `mongo.white_list_dbs` | 数据库访问白名单，为空表示不限制；非空时仅允许操作列表内的数据库 | 不限制 |

## API 接口

所有接口统一使用 `POST` 方法，通过不同动作端点区分操作，请求体（JSON）中需包含操作的数据库名 `database` 与集合名 `collection`。

### 统一响应格式

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

### 公共参数

每个操作的请求体都需包含：

| 字段 | 类型 | 说明 |
|------|------|------|
| `database` | string | 数据库名（必填） |
| `collection` | string | 集合名（必填） |

### 文档 CRUD

| 方法 | 端点 | 说明 |
|------|------|------|
| `POST` | `/api/insert` | 插入单条文档 |
| `POST` | `/api/insertmany` | 批量插入文档 |
| `POST` | `/api/find` | 查询文档列表 |
| `POST` | `/api/findone` | 查询单条文档 |
| `POST` | `/api/update` | 更新单条文档 |
| `POST` | `/api/updatemany` | 批量更新文档 |
| `POST` | `/api/delete` | 删除单条文档 |
| `POST` | `/api/deletemany` | 批量删除文档 |

### 索引管理

| 方法 | 端点 | 说明 |
|------|------|------|
| `POST` | `/api/indexes` | 列出所有索引 |
| `POST` | `/api/createindex` | 创建索引 |
| `POST` | `/api/dropindex` | 删除指定索引 |

### 健康检查

| 方法 | 端点 | 说明 |
|------|------|------|
| `GET` | `/health` | 健康检查 |

## 使用示例

### 插入文档

```javascript
const res = await fetch("http://localhost:8080/api/insert", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({
    database: "mydb",
    collection: "users",
    document: { name: "张三", age: 30 },
  }),
});
const data = await res.json();
```

### 批量插入

```javascript
const res = await fetch("http://localhost:8080/api/insertmany", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({
    database: "mydb",
    collection: "users",
    documents: [
      { name: "李四", age: 25 },
      { name: "王五", age: 28 },
    ],
  }),
});
const data = await res.json();
```

### 查询文档列表

```javascript
// 基础查询（默认 limit=10）
const res = await fetch("http://localhost:8080/api/find", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({ database: "mydb", collection: "users" }),
});
const data = await res.json();

// 带过滤、排序、分页
const res2 = await fetch("http://localhost:8080/api/find", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({
    database: "mydb",
    collection: "users",
    filter: { age: { $gte: 25 } },
    sort: [{ key: "age", value: -1 }],
    skip: 0,
    limit: 20,
  }),
});
const data2 = await res2.json();

// 仅返回指定字段（projection 投影）
const res3 = await fetch("http://localhost:8080/api/find", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({
    database: "mydb",
    collection: "users",
    filter: { age: { $gte: 25 } },
    projection: { name: 1, _id: 0 }, // 1=包含，0=排除
  }),
});
const data3 = await res3.json();
```

### 查询单条

```javascript
const res = await fetch("http://localhost:8080/api/findone", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({
    database: "mydb",
    collection: "users",
    filter: { name: "张三" },
  }),
});
const data = await res.json();
```

### 更新文档

```javascript
const res = await fetch("http://localhost:8080/api/update", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({
    database: "mydb",
    collection: "users",
    filter: { name: "张三" },
    update: { $set: { age: 31 } },
    upsert: false,
  }),
});
const data = await res.json();
```

### 删除文档

```javascript
const res = await fetch("http://localhost:8080/api/delete", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({
    database: "mydb",
    collection: "users",
    filter: { name: "张三" },
  }),
});
const data = await res.json();
```

### 创建索引

```javascript
const res = await fetch("http://localhost:8080/api/createindex", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({
    database: "mydb",
    collection: "users",
    keys: [{ key: "name", value: 1 }],
  }),
});
const data = await res.json();
```

### 列出索引

```javascript
const res = await fetch("http://localhost:8080/api/indexes", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({ database: "mydb", collection: "users" }),
});
const data = await res.json();
```

### 删除索引

```javascript
const res = await fetch("http://localhost:8080/api/dropindex", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({
    database: "mydb",
    collection: "users",
    name: "name_1",
  }),
});
const data = await res.json();
```

## 技术栈

| 组件 | 版本 | 说明 |
|------|------|------|
| [Go](https://go.dev/) | 1.21+ | 编程语言 |
| [Gin](https://gin-gonic.com/) | v1.12.0 | HTTP 框架 |
| [mongo-driver](https://go.mongodb.org/mongo-driver/) | v1.17.9 | MongoDB 官方驱动 |
| [yaml.v3](https://github.com/go-yaml/yaml) | v3.0.1 | YAML 配置解析 |
| [golang.org/x/sys](https://pkg.go.dev/golang.org/x/sys) | v0.47.0 | Windows SCM 服务管理 |
