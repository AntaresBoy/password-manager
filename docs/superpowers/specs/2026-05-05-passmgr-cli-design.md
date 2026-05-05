# passmgr CLI 设计文档

> **版本**: v1.0
> **日期**: 2026-05-05
> **技术栈**: Go 1.22+
> **范围**: MVP（初始化 vault、条目 CRUD、生成密码、剪贴板复制+定时清除）

---

## 1. 概述

`passmgr` 是一个本地命令行密码管理器。所有数据以 AES-256-GCM 加密存储在本地文件系统中，密钥由用户主密码通过 Argon2id 派生。MVP 版本仅支持单 vault 单文件，无网络依赖，无云端同步。

---

## 2. 架构与模块边界

### 2.1 目录结构

```
cmd/passmgr/
└── main.go              // 极薄入口：创建根 cobra.Command 并 Execute()

internal/
├── vault/
│   ├── vault.go         // Vault 结构体：管理生命周期（Init/Open/Save）
│   ├── crypto.go        // Argon2id + AES-256-GCM 封装（内部使用）
│   └── vault_test.go
├── store/
│   ├── store.go         // Store 接口定义
│   ├── file_store.go    // 文件系统实现
│   └── store_test.go
├── clipboard/
│   ├── clipboard.go     // Clipboard 接口定义
│   ├── system_clip.go   // 系统剪贴板实现（atotto/clipboard）
│   └── clipboard_test.go
├── passgen/
│   ├── passgen.go       // 密码生成器
│   └── passgen_test.go
├── errno/
│   └── errno.go         // 错误码枚举
└── config/
    └── config.go        // 数据文件路径解析（XDG Base Directory）

pkg/crypto/
└── crypto.go            // 可复用的 Argon2id/AES-256-GCM 小库
```

### 2.2 模块依赖关系（单向依赖）

- `cmd` → `internal/{vault,clipboard,passgen,errno}` → `pkg/crypto`
- `vault` → `store` + `pkg/crypto`
- `clipboard` 无内部依赖
- `passgen` 无内部依赖
- **禁止循环依赖**：任何模块不得依赖 `cmd`

### 2.3 关键接口

```go
// store/store.go
type Store interface {
    Read() ([]byte, error)
    Write(data []byte) error
    Exists() bool
    Path() string
}

// clipboard/clipboard.go
type Clipboard interface {
    Copy(text string) error
    Clear() error
}
```

### 2.4 外部依赖

| 依赖 | 用途 |
|------|------|
| `github.com/spf13/cobra` | CLI 命令树 |
| `golang.org/x/crypto/argon2` | Argon2id KDF |
| `github.com/atotto/clipboard` | 跨平台剪贴板 |
| `github.com/stretchr/testify` | 测试断言与 mock |
| `crypto/aes`, `crypto/cipher`, `crypto/rand` | 标准库对称加密 |

---

## 3. 数据结构与加密

### 3.1 明文数据结构

```go
type VaultData struct {
    Version    int       `json:"version"`      // 当前 = 1
    Entries    []Entry   `json:"entries"`
    ModifiedAt time.Time `json:"modified_at"`
}

type Entry struct {
    ID        string    `json:"id"`          // UUID v4
    Name      string    `json:"name"`        // 展示名，如 "GitHub"
    Username  string    `json:"username"`
    Password  string    `json:"password"`
    URL       string    `json:"url,omitempty"`
    Notes     string    `json:"notes,omitempty"`
    Tags      []string  `json:"tags,omitempty"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

### 3.2 加密文件格式（磁盘存储）

```
[4 bytes]   Magic Header: "PMV1"
[32 bytes]  Salt（Argon2id 随机盐）
[12 bytes]  AES-GCM Nonce（每次保存重新生成）
[N bytes]   Ciphertext（JSON payload 密文）
[16 bytes]  AES-GCM Tag（GCM 认证标签，自动附加）
```

### 3.3 密钥派生参数

| 参数 | 值 |
|------|-----|
| 算法 | Argon2id |
| 时间因子 (t) | 3 |
| 内存 (m) | 64 MB |
| 并行度 (p) | 4 |
| 输出密钥长度 | 32 bytes（AES-256） |
| 盐长度 | 32 bytes |

### 3.4 加密/解密流程

**加密：**
1. 生成随机 Salt（32B）和 Nonce（12B）
2. `Argon2id(password, salt)` → 32B key
3. `AES-256-GCM(key, nonce)` → encrypt(JSON)
4. 写入文件：`Magic + Salt + Nonce + Ciphertext`

**解密：**
1. 读取 Magic，校验是否为 `"PMV1"`
2. 读取 Salt（32B）、Nonce（12B）、Ciphertext
3. `Argon2id(password, salt)` → key
4. `AES-256-GCM(key, nonce)` → decrypt → JSON

### 3.5 安全红线

- 主密码仅在内存中参与 Argon2id 计算，**永不落盘**
- 密码字段在 JSON 序列化后参与加密，**不以明文形式存储**
- **每次 `Save()` 重新生成 Salt + Nonce**，防止重放分析和文件差异推断
- Salt 与密文一同存储在文件中，解密时从文件读取
- 剪贴板复制后启动后台 goroutine，**10 秒后自动调用 `Clear()`**

### 3.6 默认存储路径

| 平台 | 路径 |
|------|------|
| Linux | `$XDG_DATA_HOME/passmgr/vault.dat`（fallback `~/.local/share/passmgr/vault.dat`） |
| macOS | `~/Library/Application Support/passmgr/vault.dat` |
| Windows | `%APPDATA%\passmgr\vault.dat` |

---

## 4. 数据流与命令生命周期

### 4.1 `passmgr init` — 初始化 vault

```
CLI 解析 → 检查 vault 已存在？ → 提示输入主密码 → 确认主密码
  → 创建空 VaultData{Version:1} → vault.EncryptAndSave(data, password)
  → 提示 "Vault created at <path>"
```

### 4.2 `passmgr add <name>` — 添加条目

```
CLI 解析 → 检查 vault 存在？ → 提示主密码 → vault.Open(password) → VaultData
  → 交互式收集：username, password（为空时提示是否生成随机密码）, url（可选）, notes（可选）, tags（可选）
  → 生成 UUID → 追加到 Entries → vault.Save(data, password)
  → 提示 "Added: <name> (<uuid>)"
```

### 4.3 `passmgr get <name>` — 查看条目

```
CLI 解析 → 检查 vault 存在？ → 提示主密码 → vault.Open(password) → VaultData
  → 按 name（模糊匹配第一条）或 --id 精确查找 Entry
  → 格式化输出（password 默认显示为 "********"，--show-password 才明文）
```

### 4.4 `passmgr list` — 列出全部

```
CLI 解析 → 检查 vault 存在？ → 提示主密码 → vault.Open(password) → VaultData
  → 按 UpdatedAt 倒序输出表格：Name | Username | URL | Tags
```

### 4.5 `passmgr rm <name>` — 删除条目

```
CLI 解析 → 检查 vault 存在？ → 提示主密码 → vault.Open(password) → VaultData
  → 查找 Entry → 确认 "Delete <name>? [y/N]" → 从 Entries 移除
  → vault.Save(data, password) → 提示已删除
```

### 4.6 `passmgr gen` — 生成密码（无需主密码）

```
CLI 解析 --length/--upper/--lower/--digits/--symbols
  → passgen.Generate(opts) → 输出密码到 stdout
  → --copy ? clipboard.Copy(password) + 10s 后 Clear()
```

### 4.7 `passmgr cp <name>` — 复制密码到剪贴板

```
CLI 解析 → 检查 vault 存在？ → 提示主密码 → vault.Open(password) → VaultData
  → 查找 Entry → clipboard.Copy(password)
  → 启动 goroutine: time.Sleep(10s) → clipboard.Clear()
  → 提示 "Copied, will clear in 10s"
```

### 4.8 公共前置检查

每条命令进入后执行：
1. 解析全局 flag `--vault-path`（默认走 XDG 路径）
2. 判断命令是否需要 vault 已存在 / 需要主密码
3. 主密码通过 `golang.org/x/term.ReadPassword` 隐藏输入

---

## 5. 错误处理

### 5.1 错误码定义

```go
package errno

var (
    OK                  = NewError(0, "success", 0)
    ErrInternal         = NewError(10001, "internal error", 10)

    // Vault 层 20001-20099
    ErrVaultNotFound    = NewError(20001, "vault not found", 2)
    ErrVaultExists      = NewError(20002, "vault already exists", 5)
    ErrVaultCorrupted   = NewError(20003, "vault file corrupted", 2)
    ErrWrongPassword    = NewError(20004, "wrong master password", 3)

    // Entry 层 20101-20199
    ErrEntryNotFound    = NewError(20101, "entry not found", 4)
    ErrEntryExists      = NewError(20102, "entry already exists", 5)

    // 校验层 20201-20299
    ErrInvalidInput     = NewError(20201, "invalid input", 5)
    ErrPasswordMismatch = NewError(20202, "passwords do not match", 5)

    // 系统层 20301-20399
    ErrClipboardFail    = NewError(20301, "clipboard unavailable", 1)
)
```

### 5.2 Error 结构体

```go
type Error struct {
    code    int    // 业务码
    message string // 用户可见消息
    exit    int    // 进程退出码
    cause   error  // 内部原因（调试/日志）
}

func (e *Error) Error() string   { return e.message }
func (e *Error) ExitCode() int   { return e.exit }
func (e *Error) Unwrap() error   { return e.cause }
```

### 5.3 错误传播规则

- **Vault 层** 返回 `*errno.Error`，内部错误（如文件 IO 失败）包装为 `ErrInternal` 并附带 `cause`
- **CLI 层**（cobra）统一处理：打印 `Error: <message>` 到 stderr，以 `ExitCode()` 退出
- **禁止裸返回** `return err` — 所有错误必须经过 `errno` 包装

### 5.4 CLI 输出示例

```bash
# 正常错误
$ passmgr get github
Enter master password: ****
Error: entry not found
# 退出码 4

# 内部错误（DEBUG=1 时显示 cause）
$ passmgr list
Error: internal error
# DEBUG=1 时附加: cause: open /path: permission denied
```

### 5.5 Panic 防护

- `cmd/passmgr/main.go` 顶层用 `defer recover()` 捕获 panic
- panic → 记录 stderr + 退出码 10
- `vault.Save` 操作前先在临时文件写入，成功后原子重命名，避免写坏原文件

---

## 6. 测试策略

### 6.1 覆盖率红线

```
Lines      ≥ 80%
Branches   ≥ 70%
Functions  ≥ 80%
```

### 6.2 单元测试矩阵

| 包 | 测试重点 | Mock 策略 |
|----|---------|----------|
| `pkg/crypto` | Argon2id 派生一致性、AES-GCM 加解密往返、错误密码解密失败、篡改密文检测 | 无外部依赖，纯算法测试 |
| `internal/vault` | Init/Open/Save 生命周期、错误密码拒绝、损坏文件识别 | Mock `Store` 接口（内存实现） |
| `internal/store` | 文件读写、XDG 路径解析、`Exists()`、权限错误 | `testing.TempDir()` + 真实文件 |
| `internal/clipboard` | Copy/Clear 调用记录、超时清除触发 | Mock `Clipboard` 接口 |
| `internal/passgen` | 长度正确性、字符集合规、熵检测、边界值 | 无 |
| `internal/errno` | 错误码值、ExitCode()、Unwrap()、错误链 | 无 |

### 6.3 集成测试

`tests/integration/cli_test.go` 使用 `os/exec` 构建真实二进制并运行：

```go
func TestCLI_FullLifecycle(t *testing.T) {
    // 1. passmgr init --vault-path /tmp/xxx
    // 2. passmgr add github --vault-path /tmp/xxx
    // 3. passmgr list → 断言包含 github
    // 4. passmgr get github --show-password → 断言密码正确
    // 5. passmgr rm github
    // 6. passmgr list → 断言空
}
```

### 6.4 E2E 测试说明

openspec config 中指定了 Cypress 用于 E2E，但本项目为 CLI 工具而非 Web 应用。

- **当前 MVP**：E2E 由 `tests/integration/cli_test.go` 承担（Go `os/exec` 驱动真实二进制）
- **未来扩展**：如需 TUI（bubbletea），可引入 expect-style 终端交互测试

### 6.5 Mock 实现示例

```go
// 仅测试文件中使用
type MockStore struct {
    Data   []byte
    FileExists bool
}

func (m *MockStore) Read() ([]byte, error) { return m.Data, nil }
func (m *MockStore) Write(d []byte) error  { m.Data = d; return nil }
func (m *MockStore) Exists() bool          { return m.FileExists }
func (m *MockStore) Path() string          { return "/mock/vault.dat" }
```

### 6.6 测试命名规范

```
TestVault_Open_WithWrongPassword
TestVault_Save_AtomicWrite
TestPassGen_Generate_WithAllCharsets
TestCLI_FullLifecycle
```

---

## 7. 安全设计

| 层面 | 措施 |
|------|------|
| 密钥派生 | Argon2id（t=3, m=64MB, p=4），抵御 GPU/ASIC 暴力破解 |
| 对称加密 | AES-256-GCM，提供机密性 + 认证，防止篡改 |
| 盐值管理 | 每次保存生成新 Salt，与密文一同存储 |
| Nonce 管理 | 每次保存生成新随机 Nonce，12 bytes 由 `crypto/rand` 提供 |
| 文件原子写 | 先写临时文件，成功后 `os.Rename` 原子替换 |
| 剪贴板 | 复制后 10 秒自动清除，降低肩窥风险 |
| 主密码输入 | `term.ReadPassword` 隐藏回显，不记录历史 |

---

## 8. 未来扩展

以下功能不在 MVP 范围内，但架构预留了扩展空间：

- **搜索/过滤**：`list` 已输出全部字段，后续可加 `--search` flag
- **导入/导出**：JSON/CSV 转换层可独立于 vault 加密层实现
- **主密码修改**：重新加密全部数据即可，无需改存储格式
- **TOTP**：可新增 `Entry.TOTPSecret` 字段 + `passmgr otp` 子命令
- **多 vault**：`--vault-path` flag 已支持，后续可加 `vault switch`
- **TUI**：`bubbletea` 可替换 cobra 的交互式输入
- **同步**：vault 是单文件，天然可用 git/syncthing 同步

---

## 9. 附录：CLI 命令速查

```bash
passmgr init                          # 初始化 vault
passmgr add <name>                    # 添加条目（交互式）
passmgr get <name>                    # 查看条目（密码隐藏）
passmgr get <name> --show-password    # 查看明文密码
passmgr list                          # 列出全部条目
passmgr rm <name>                     # 删除条目（需确认）
passmgr gen --length 16               # 生成密码
passmgr gen --length 16 --copy        # 生成并复制到剪贴板
passmgr cp <name>                     # 复制已有条目密码到剪贴板
passmgr --vault-path /custom/path     # 指定 vault 路径
```
