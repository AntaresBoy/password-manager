# 后端开发规范 (BE Spec)

> **版本**: v1.0  
> **适用技术栈**: Go 1.22+ + Gin + GORM + MySQL/Redis  
> **配套工具**: swag + golangci-lint + gofmt  
> **最后更新**: 2026-05-05

---

## 1. 技术栈与版本锁定

| 依赖 | 版本要求 | 说明 |
|------|---------|------|
| Go | 1.22+ | 使用 `slog` 标准日志、`context` 超时控制 |
| Gin | v1.9+ | Web 框架 |
| GORM | v2 | ORM（禁止原生 SQL 拼接，防注入） |
| MySQL | 8.0+ | 主数据库 |
| Redis | 7.0+ | 缓存 / 分布式锁 |
| JWT | v5 | 身份认证 |
| Swagger | v1.16+ | API 文档（swag） |
| Viper | v1.18+ | 配置管理 |

---

## 2. 项目目录结构（Standard Go Project Layout）

```
cmd/
├── server/                 # 主应用入口
│   └── main.go
internal/
├── api/                    # HTTP 层（Handler / Router）
│   ├── handler/
│   │   ├── user.go
│   │   └── auth.go
│   ├── middleware/
│   │   ├── jwt.go
│   │   ├── cors.go
│   │   └── logger.go
│   └── router/
│       └── router.go
├── service/                # 业务逻辑层（Biz Logic）
│   ├── user.go
│   └── auth.go
├── repository/             # 数据访问层（DAO）
│   ├── user.go
│   └── db.go
├── model/                  # 领域模型 / 实体
│   ├── user.go
│   └── dto/                # 请求/响应结构体
│       ├── user_req.go
│       └── user_resp.go
├── pkg/                    # 内部公共库
│   ├── errno/              # 错误码定义
│   ├── response/           # 统一响应封装
│   ├── logger/             # 日志封装
│   └── validator/          # 校验工具
├── config/                 # 配置文件
│   └── config.yaml
└── bootstrap/              # 启动初始化
    ├── db.go
    ├── redis.go
    └── router.go
pkg/                        # 可被外部引用的公共库（可选）
scripts/                    # 脚本
Makefile
go.mod
go.sum
```

**禁止**: 在 `internal/` 外直接写业务代码；`main.go` 必须极薄（仅启动）。

---

## 3. 代码风格

### 3.1 格式化（强制）
```bash
# 提交前必须执行
gofmt -w .
golangci-lint run
```

### 3.2 命名规范
| 类型 | 规范 | 示例 |
|------|------|------|
| 包名 | 小写单数 | `service`, `repository` |
| 文件 | snake_case | `user_handler.go` |
| 接口 | 动词 + er | `UserServicer`, `UserRepository` |
| 结构体 | 名词 | `User`, `OrderDetail` |
| 方法 | 动词/动宾 | `CreateUser`, `GetByID` |
| 常量 | 大写下划线 | `MaxRetryCount = 3` |
| 私有 | 小写开头 | `userCache` |
| 错误变量 | `Err` 前缀 | `ErrUserNotFound` |

### 3.3 接口隔离
```go
// repository/user.go
// 仓储接口定义在调用方（service），实现方（repository）依赖接口

type UserRepository interface {
    Create(ctx context.Context, user *model.User) error
    GetByID(ctx context.Context, id int64) (*model.User, error)
    Update(ctx context.Context, user *model.User) error
    Delete(ctx context.Context, id int64) error
    List(ctx context.Context, page, size int) ([]*model.User, int64, error)
}
```

---

## 4. API 设计规范（RESTful + Swagger）

### 4.1 URL 设计
```
GET    /api/v1/users          # 列表（分页）
GET    /api/v1/users/:id      # 详情
POST   /api/v1/users          # 创建
PUT    /api/v1/users/:id      # 全量更新
PATCH  /api/v1/users/:id      # 局部更新
DELETE /api/v1/users/:id      # 删除
```

### 4.2 Swagger 注释（必须）
每个 Handler 必须写 Swagger 注释，否则 CI 阻断：
```go
// GetUser godoc
// @Summary      获取用户详情
// @Description  根据用户 ID 获取详细信息
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "用户ID"
// @Success      200  {object}  response.Response{data=model.User}
// @Failure      400  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/v1/users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        response.Error(c, errno.ErrInvalidParam)
        return
    }
    // ...
}
```

### 4.3 统一响应格式
```go
// pkg/response/response.go
type Response struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

// 成功响应
func Success(c *gin.Context, data interface{}) {
    c.JSON(http.StatusOK, Response{
        Code:    0,
        Message: "success",
        Data:    data,
    })
}

// 错误响应（使用 errno 错误码）
func Error(c *gin.Context, err *errno.Error) {
    c.JSON(err.HTTPCode(), Response{
        Code:    err.Code(),
        Message: err.Message(),
    })
}
```

---

## 5. 错误处理规范

### 5.1 错误码定义（errno 包）
```go
// pkg/errno/code.go
package errno

var (
    OK                  = NewError(0, "success")
    ErrInternal         = NewError(10001, "服务器内部错误")
    ErrInvalidParam     = NewError(10002, "请求参数错误")
    ErrUnauthorized     = NewError(10003, "未授权")
    ErrForbidden        = NewError(10004, "禁止访问")
    ErrNotFound         = NewError(10005, "资源不存在")
    ErrTooManyRequests  = NewError(10006, "请求过于频繁")
)

// 业务错误码 20001-29999
var (
    ErrUserNotFound     = NewError(20001, "用户不存在")
    ErrPasswordWrong    = NewError(20002, "密码错误")
    ErrUserExists       = NewError(20003, "用户已存在")
)
```

### 5.2 错误处理原则
- **禁止裸返回错误**: `return err` → 必须包装为 `errno` 错误码
- **日志记录**: 内部错误（500）必须记录 `slog.Error()` 含堆栈
- **用户提示**: 客户端只展示 `Message`，不暴露内部细节

```go
// ✅ 正确
user, err := h.svc.GetUser(ctx, id)
if err != nil {
    if errors.Is(err, gorm.ErrRecordNotFound) {
        response.Error(c, errno.ErrUserNotFound)
        return
    }
    slog.ErrorContext(ctx, "get user failed", "error", err, "user_id", id)
    response.Error(c, errno.ErrInternal)
    return
}

// ❌ 错误：直接暴露内部错误
if err != nil {
    c.JSON(500, gin.H{"error": err.Error()})  // 泄露堆栈！
}
```

---

## 6. 数据库与 GORM 规范

### 6.1 Model 定义
```go
// model/user.go
package model

import "time"

type User struct {
    ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
    Username  string    `gorm:"size:64;not null;uniqueIndex" json:"username"`
    Password  string    `gorm:"size:128;not null" json:"-"`  // 禁止序列化密码
    Email     string    `gorm:"size:128" json:"email"`
    Phone     string    `gorm:"size:32;index" json:"phone"`
    Status    int8      `gorm:"default:1;comment:1正常 2禁用" json:"status"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

func (User) TableName() string {
    return "users"
}
```

### 6.2 查询规范
```go
// ✅ 正确：使用 GORM 链式调用，禁止拼接 SQL
func (r *userRepo) List(ctx context.Context, page, size int) ([]*model.User, int64, error) {
    var users []*model.User
    var total int64

    db := r.db.WithContext(ctx).Model(&model.User{})

    if err := db.Count(&total).Error; err != nil {
        return nil, 0, err
    }

    if err := db.Offset((page - 1) * size).Limit(size).Find(&users).Error; err != nil {
        return nil, 0, err
    }

    return users, total, nil
}

// ❌ 错误：SQL 拼接（SQL 注入风险）
query := fmt.Sprintf("SELECT * FROM users WHERE name = '%s'", name)  // 严重安全漏洞！
```

### 6.3 事务处理
```go
func (s *userService) Transfer(ctx context.Context, from, to int64, amount float64) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        if err := s.repo.DecreaseBalance(tx, from, amount); err != nil {
            return err
        }
        if err := s.repo.IncreaseBalance(tx, to, amount); err != nil {
            return err
        }
        return nil
    })
}
```

---

## 7. 缓存与 Redis 规范

### 7.1 缓存策略
- **读多写少**: Cache-Aside（先读缓存，未命中读 DB 再写缓存）
- **写频繁**: Write-Through 或延迟双删
- **缓存时间**: 默认 5 分钟，热点数据 1 小时，配置类数据永久

### 7.2 Key 命名规范
```
{project}:{module}:{business}:{id}
# 示例
ai_nexus:user:profile:10086
ai_nexus:order:list:10086:1:20
```

### 7.3 分布式锁
```go
import "github.com/bsm/redislock"

func (s *service) CreateOrder(ctx context.Context, userID int64) error {
    lockKey := fmt.Sprintf("lock:order:%d", userID)
    lock, err := s.locker.Obtain(ctx, lockKey, 5*time.Second, nil)
    if err != nil {
        return errno.ErrTooManyRequests
    }
    defer lock.Release(ctx)

    // 执行业务逻辑
    return nil
}
```

---

## 8. 日志规范（slog）

### 8.1 日志级别使用
| 级别 | 使用场景 |
|------|---------|
| DEBUG | 开发调试，生产关闭 |
| INFO  | 业务流水（请求进入、离开） |
| WARN  | 可恢复异常（参数校验失败、限流触发） |
| ERROR | 不可恢复错误（DB 连接失败、Panic 恢复） |

### 8.2 结构化日志（必须）
```go
// ✅ 正确：结构化键值对
slog.InfoContext(ctx, "user login",
    "user_id", user.ID,
    "ip", c.ClientIP(),
    "duration_ms", time.Since(start).Milliseconds(),
)

// ❌ 错误：字符串拼接
log.Printf("user %d login from %s", user.ID, c.ClientIP())  // 不利于检索
```

### 8.3 请求日志中间件
每个请求必须记录：
- 请求 ID（`X-Request-ID`）
- 方法、路径、状态码
- 耗时（ms）
- 错误信息（如有）

---

## 9. 并发与安全

### 9.1 Goroutine 规范
- **禁止裸起 goroutine**: 必须使用 `errgroup` 或自定义 Worker Pool
- **必须传递 context**: 所有 goroutine 接收 `ctx`，支持级联取消
- **必须处理 panic**: 使用 `recover()` 防止单个 goroutine 崩溃拖垮服务

```go
import "golang.org/x/sync/errgroup"

g, ctx := errgroup.WithContext(ctx)
for _, id := range ids {
    id := id  // 闭包捕获
    g.Go(func() error {
        return s.processOne(ctx, id)
    })
}
if err := g.Wait(); err != nil {
    return err
}
```

### 9.2 安全红线
- **JWT**: Access Token 15 分钟，Refresh Token 7 天，必须 HTTPS 传输
- **CORS**: 白名单控制，禁止 `*` 通配生产环境
- **参数校验**: 所有入参使用 `go-playground/validator` 标签校验
- **SQL 注入**: 100% 使用 GORM/参数化查询，禁止字符串拼接
- **敏感数据**: 密码 bcrypt 加密（cost ≥ 10），禁止明文存储

---

## 10. 测试规范

### 10.1 单元测试
```go
// service/user_test.go
package service

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

func TestUserService_CreateUser(t *testing.T) {
    mockRepo := new(mockUserRepo)
    svc := NewUserService(mockRepo)

    mockRepo.On("GetByUsername", mock.Anything, "test").
        Return(nil, gorm.ErrRecordNotFound)
    mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*model.User")).
        Return(nil)

    err := svc.CreateUser(context.Background(), &model.User{Username: "test"})
    assert.NoError(t, err)
    mockRepo.AssertExpectations(t)
}
```

### 10.2 测试要求
- 覆盖率：Service 层 ≥ 80%，Repository 层 ≥ 60%
- Mock：外部依赖（DB、Redis、第三方 API）全部 Mock
- 表驱动：复杂场景使用 `tests := []struct{...}{}`
- 并行：无状态测试使用 `t.Parallel()`
