# passmgr

`passmgr` 是一个本地账号密码管理 CLI。它使用主密码派生加密密钥，将账号条目写入本地 AES-GCM 加密 vault 文件，不依赖云端服务。

完整使用和开发说明见 [docs/passmgr-usage-development.md](docs/passmgr-usage-development.md)。

当前已生成 macOS Apple Silicon 分发包：

```bash
dist/passmgr-darwin-arm64.tar.gz
dist/passmgr-darwin-arm64.tar.gz.sha256
```
