# passmgr

`passmgr` 是一个本地账号密码管理 CLI，用主密码保护本地加密 vault 文件，适合个人开发者在终端里管理网站、服务和工具账号。

## 功能

- 初始化本地加密 vault。
- 添加、列出、查看、删除账号条目。
- 默认隐藏密码，只有 `get --show-password` 才输出明文。
- 生成随机密码。
- 复制密码到系统剪贴板，并安排 10 秒后清空。
- 支持通过 `--vault-path` 指定 vault 文件路径。
- 支持 `PASSMGR_MASTER_PASSWORD` 做非交互式脚本调用。

## 安全模型

- Vault 使用 `PMV1 + salt + nonce + ciphertext` 文件格式。
- 密钥由主密码通过 Argon2id 派生。
- Vault 内容使用 AES-256-GCM 加密。
- 每次保存会重新生成 salt 和 nonce。
- POSIX 平台下 vault 文件权限为 `0600`，父目录权限为 `0700`。
- 主密码不写入磁盘。

当前限制：交互式主密码输入使用普通行读取，输入时会回显。自动化场景建议使用 `PASSMGR_MASTER_PASSWORD`。

## 快速开始

构建当前平台二进制：

```bash
go build -o dist/passmgr ./cmd/passmgr
```

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

删除条目：

```bash
PASSMGR_MASTER_PASSWORD='change-me' \
./dist/passmgr --vault-path ./vault.dat rm github --yes
```

生成随机密码：

```bash
./dist/passmgr gen --length 20
```

## 分发包

当前工作区已生成 macOS Apple Silicon 分发包：

```bash
dist/passmgr-darwin-arm64.tar.gz
dist/passmgr-darwin-arm64.tar.gz.sha256
```

校验：

```bash
shasum -a 256 -c dist/passmgr-darwin-arm64.tar.gz.sha256
```

## 开发

要求：

- Go 1.23+

运行测试：

```bash
go test ./...
```

在当前受限 worktree 中推荐使用本地缓存：

```bash
GOCACHE=$PWD/.gocache GOMODCACHE=$PWD/.gomodcache GOPROXY=off go test ./...
```

构建分发二进制：

```bash
GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags='-s -w' -o dist/passmgr ./cmd/passmgr
tar -C dist -czf dist/passmgr-darwin-arm64.tar.gz passmgr README-passmgr.txt
shasum -a 256 dist/passmgr-darwin-arm64.tar.gz > dist/passmgr-darwin-arm64.tar.gz.sha256
```

## 目录结构

```text
cmd/passmgr/          CLI 入口
internal/clipboard/  剪贴板接口和系统实现
internal/config/     vault 路径解析
internal/errno/      应用错误码和退出码
internal/passgen/    随机密码生成
internal/store/      文件存储接口和实现
internal/vault/      vault 数据模型、加密文件格式和生命周期
pkg/crypto/          Argon2id 与 AES-GCM 原语
```

## 详细文档

完整使用、开发、冒烟测试和发布说明见：

- [docs/passmgr-usage-development.md](docs/passmgr-usage-development.md)

## 当前发布状态

- Feature 分支：`feature/implement-passmgr-mvp`
- 最近文档提交：`5608a29 docs: add passmgr usage and development guide`
- OpenSpec 通知报告提交：`4431ced docs(openspec): add notification report summary`
- 飞书通知脚本已调用成功，接口返回 `StatusCode:0` / `msg:success`
- 覆盖率门禁已按用户要求跳过，报告中记录的实测 statement coverage 为 `21.8%`
