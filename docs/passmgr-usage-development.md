# passmgr 使用和开发说明

## 1. 项目概览

`passmgr` 是本地账号密码管理 CLI。MVP 支持初始化加密 vault、添加/查看/列出/删除条目、生成随机密码，以及复制密码到系统剪贴板。

核心特性：

- 本地单文件 vault，不做云端同步。
- 主密码不落盘。
- Vault 内容使用 Argon2id 派生密钥和 AES-256-GCM 加密。
- 默认查看条目时隐藏密码，只有显式传入 `--show-password` 才输出明文。
- `cp` 和 `gen --copy` 会复制到系统剪贴板，并安排 10 秒后清空。

## 2. 安装与分发包

当前工作区已生成 macOS Apple Silicon 分发包：

```bash
dist/passmgr-darwin-arm64.tar.gz
dist/passmgr-darwin-arm64.tar.gz.sha256
```

校验分发包：

```bash
shasum -a 256 -c dist/passmgr-darwin-arm64.tar.gz.sha256
```

解压：

```bash
mkdir -p /tmp/passmgr-release
tar -xzf dist/passmgr-darwin-arm64.tar.gz -C /tmp/passmgr-release
/tmp/passmgr-release/passmgr
```

本机直接使用已构建二进制：

```bash
./dist/passmgr
```

## 3. Vault 路径

默认 vault 路径由 `internal/config` 根据平台解析：

- macOS: `~/Library/Application Support/passmgr/vault.dat`
- Linux/其他: `$XDG_DATA_HOME/passmgr/vault.dat`，未设置时为 `~/.local/share/passmgr/vault.dat`
- Windows: `%APPDATA%/passmgr/vault.dat`，未设置时为 `%USERPROFILE%/AppData/Roaming/passmgr/vault.dat`

也可以在命令前传入自定义路径：

```bash
./dist/passmgr --vault-path ./vault.dat init
```

## 4. 主密码输入

交互式使用时，命令会提示输入主密码。

自动化或测试时可使用环境变量：

```bash
export PASSMGR_MASTER_PASSWORD='change-me'
export PASSMGR_MASTER_PASSWORD_CONFIRM='change-me'
```

`PASSMGR_MASTER_PASSWORD_CONFIRM` 只用于 `init` 命令确认主密码。

## 5. 常用命令

初始化 vault：

```bash
PASSMGR_MASTER_PASSWORD='change-me' \
PASSMGR_MASTER_PASSWORD_CONFIRM='change-me' \
./dist/passmgr --vault-path ./vault.dat init
```

添加条目：

```bash
PASSMGR_MASTER_PASSWORD='change-me' \
./dist/passmgr --vault-path ./vault.dat add github -u antares -p 'secret123' --url github.com --notes dev
```

如果不传 `-p`，会自动生成一个默认长度为 16 的随机密码：

```bash
PASSMGR_MASTER_PASSWORD='change-me' \
./dist/passmgr --vault-path ./vault.dat add gitlab -u antares --url gitlab.com
```

列出条目：

```bash
PASSMGR_MASTER_PASSWORD='change-me' \
./dist/passmgr --vault-path ./vault.dat list
```

查看条目，默认隐藏密码：

```bash
PASSMGR_MASTER_PASSWORD='change-me' \
./dist/passmgr --vault-path ./vault.dat get github
```

查看明文密码：

```bash
PASSMGR_MASTER_PASSWORD='change-me' \
./dist/passmgr --vault-path ./vault.dat get github --show-password
```

复制密码到剪贴板：

```bash
PASSMGR_MASTER_PASSWORD='change-me' \
./dist/passmgr --vault-path ./vault.dat cp github
```

删除条目：

```bash
PASSMGR_MASTER_PASSWORD='change-me' \
./dist/passmgr --vault-path ./vault.dat rm github --yes
```

生成随机密码：

```bash
./dist/passmgr gen --length 20
```

生成随机密码并复制到剪贴板：

```bash
./dist/passmgr gen --length 20 --copy
```

生成不含符号的随机密码：

```bash
./dist/passmgr gen --length 20 --no-symbols
```

## 6. 错误码

CLI 会将内部错误映射为退出码：

- `2`: vault 不存在或文件损坏。
- `3`: 主密码错误。
- `4`: 条目不存在。
- `5`: 输入无效、vault 已存在、条目重复或密码确认不一致。
- `10`: 内部错误。

## 7. 安全说明

当前实现的安全边界：

- Vault 文件以 `PMV1 + salt + nonce + ciphertext` 格式保存。
- 每次保存会重新生成 salt 和 nonce。
- Vault 明文 JSON 只在内存中出现。
- 文件写入使用临时文件和同目录 rename 替换。
- Vault 文件权限在 POSIX 平台上设置为 `0600`，父目录设置为 `0700`。

当前限制：

- 交互式输入当前使用普通行读取，输入时会回显；自动化可使用 `PASSMGR_MASTER_PASSWORD`。
- 单 vault 单文件，不支持多用户、同步或远程备份。
- `list` 不显示密码；`get --show-password` 会将密码输出到终端。
- 剪贴板清理由后台 goroutine 安排，CLI 进程很快退出时不同平台的实际清理行为可能受运行时影响。

## 8. 开发环境

要求：

- Go 1.23+
- macOS/Linux/Windows 均可构建；当前分发包是 `darwin/arm64`

依赖：

- `github.com/atotto/clipboard`
- `golang.org/x/crypto`
- `golang.org/x/sys`

模块信息见根目录 `go.mod`。

## 9. 代码结构

```text
cmd/passmgr/
  main.go                 CLI 入口和命令解析

internal/clipboard/
  clipboard.go            Clipboard 接口和延时清理 helper
  system_clip.go          系统剪贴板实现

internal/config/
  config.go               默认 vault 路径解析

internal/errno/
  errno.go                应用错误码和退出码

internal/passgen/
  passgen.go              安全随机密码生成

internal/store/
  store.go                Store 接口
  file_store.go           文件读写实现

internal/vault/
  vault.go                Vault 数据模型和生命周期
  crypto.go               Vault 文件格式封装

pkg/crypto/
  crypto.go               Argon2id 和 AES-GCM 原语
```

## 10. 本地开发命令

运行测试：

```bash
go test ./...
```

在当前 sandbox/worktree 中，建议使用本地缓存目录：

```bash
GOCACHE=$PWD/.gocache GOMODCACHE=$PWD/.gomodcache GOPROXY=off go test ./...
```

构建当前平台二进制：

```bash
go build -o dist/passmgr ./cmd/passmgr
```

构建 macOS Apple Silicon 分发二进制：

```bash
GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags='-s -w' -o dist/passmgr ./cmd/passmgr
```

生成压缩包和校验文件：

```bash
tar -C dist -czf dist/passmgr-darwin-arm64.tar.gz passmgr README-passmgr.txt
shasum -a 256 dist/passmgr-darwin-arm64.tar.gz > dist/passmgr-darwin-arm64.tar.gz.sha256
```

## 11. 冒烟测试流程

```bash
rm -f /tmp/passmgr-smoke-vault.dat

PASSMGR_MASTER_PASSWORD=testpass \
PASSMGR_MASTER_PASSWORD_CONFIRM=testpass \
./dist/passmgr --vault-path /tmp/passmgr-smoke-vault.dat init

PASSMGR_MASTER_PASSWORD=testpass \
./dist/passmgr --vault-path /tmp/passmgr-smoke-vault.dat add github -u antares -p secret123 --url github.com --notes dev

PASSMGR_MASTER_PASSWORD=testpass \
./dist/passmgr --vault-path /tmp/passmgr-smoke-vault.dat list

PASSMGR_MASTER_PASSWORD=testpass \
./dist/passmgr --vault-path /tmp/passmgr-smoke-vault.dat get github

PASSMGR_MASTER_PASSWORD=testpass \
./dist/passmgr --vault-path /tmp/passmgr-smoke-vault.dat get github --show-password

PASSMGR_MASTER_PASSWORD=testpass \
./dist/passmgr --vault-path /tmp/passmgr-smoke-vault.dat rm github --yes
```

错误主密码检查：

```bash
PASSMGR_MASTER_PASSWORD=wrong \
./dist/passmgr --vault-path /tmp/passmgr-smoke-vault.dat list
```

预期退出码为 `3`，错误消息为 `wrong master password`。

## 12. 发布注意事项

发布前建议执行：

```bash
go test ./...
go vet ./...
go mod tidy
go build -trimpath -ldflags='-s -w' -o dist/passmgr ./cmd/passmgr
```

当前为了满足“直接打包构建分发”的目标，覆盖率门禁已按用户要求跳过。后续如果恢复完整 SDD gate，需要补齐 `passgen`、`vault` 和 `cmd/passmgr` 的测试覆盖，并重新生成 OpenSpec verify 报告。
