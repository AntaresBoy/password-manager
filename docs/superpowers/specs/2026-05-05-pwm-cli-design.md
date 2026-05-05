# PWM - 本地密码管理器 CLI 设计文档

> **日期**: 2026-05-05
> **技术栈**: Go 1.22+ + Cobra CLI
> **架构**: 单文件保险库 + Argon2id/AES-256-GCM 加密
> **状态**: 已批准

---

## 1. 项目概览

**项目名称**: `pwm` (Password Manager)

**一句话定义**: 本地加密文件存储的密码管理器 CLI，用主密码 + AES-256-GCM 保护所有凭据。

**目标用户**: 需要安全管理多个账号密码的个人开发者，偏好终端操作。

**产品形态**: 纯命令式 CLI 工具，所有操作通过命令行参数完成，可脚本化。

---

## 2. 命令设计

### 2.1 命令列表

```
pwm init                    # 初始化保险库（设置主密码，创建 vault 文件）
pwm add -n <name> -u <user> -s <site> [-t <tags>] [-c <comment>]  # 添加条目
pwm list [-t <tag>]         # 列出所有条目（可选按标签过滤）
pwm get <name>              # 获取指定条目详情（显示密码）
pwm search <keyword>        # 搜索条目（按名称/站点/用户名）
pwm delete <name>           # 删除条目
pwm generate [-l <length>] [--no-symbols]  # 生成随机密码
pwm export [-f csv|json]    # 导出条目到 CSV/JSON
pwm import <file> [-f csv|json]  # 从 CSV/JSON 导入条目
pwm strength <password>     # 检查密码强度
pwm rename <old> <new>      # 重命名条目
```

### 2.2 主密码输入

所有需要解密 vault 的操作会提示输入主密码（从终端 stdin 读取，不回显）。

脚本化场景可通过环境变量 `PWM_MASTER_PASSWORD` 提供，避免交互式输入。

### 2.3 输出格式

- `pwm list`: 表格格式，显示 name / username / site / tags，密码默认隐藏
- `pwm get <name>`: 显示完整条目包括密码
- `pwm search`: 与 list 格式一致，只显示匹配结果

---

## 3. 数据模型

### 3.1 Vault 文件结构

文件路径: `~/.pwm/vault.pwm`

文件格式为两段式：

```
Header (明文 JSON)
---
Body (AES-256-GCM 加密后的 Base64)
```

### 3.2 Header 结构

```json
{
  "version": 1,
  "salt": "<argon2id-salt-base64>",
  "params": {
    "time": 3,
    "memory": 65536,
    "threads": 4
  }
}
```

- `version`: vault 格式版本号，便于未来迁移
- `salt`: 16 bytes 随机盐，init 时生成，固定不变
- `params`: Argon2id 参数（OWASP 推荐值）

### 3.3 Body 结构（解密后的 JSON）

```json
{
  "entries": [
    {
      "name": "github",
      "username": "antares",
      "site": "github.com",
      "password": "s3cur3P@ss!",
      "tags": ["dev", "social"],
      "comment": "工作账号",
      "created_at": "2026-05-05T10:00:00Z",
      "updated_at": "2026-05-05T10:00:00Z"
    }
  ]
}
```

- `name`: 条目唯一标识，用于 `get/delete/rename` 命令定位
- `password`: 明文存储在加密体内，整体文件加密保护
- `tags`: 字符串数组，支持 `list -t <tag>` 过滤
- 时间字段: RFC 3339 格式

---

## 4. 加密流程

### 4.1 密钥派生

```
master_password → Argon2id(salt, time=3, memory=64KB, threads=4) → 32-byte key
```

Argon2id 是 OWASP 推荐的密钥派生函数，抗 GPU/ASIC 破解。

### 4.2 加密写入

```
vault_body_json → AES-256-GCM(key, random_nonce) → ciphertext + auth_tag
写入: base64(nonce + auth_tag + ciphertext)
```

每次 Save vault 时生成新的 12-byte nonce，防止 nonce reuse attack。

### 4.3 解密读取

```
body_base64 → decode → nonce + auth_tag + ciphertext
AES-256-GCM-Decrypt(key, nonce, auth_tag, ciphertext) → vault_body_json
```

如果主密码错误，GCM 认证标签校验会失败，返回明确错误。

### 4.4 安全要点

- 主密码绝不存储在文件中
- 每次写入生成新 nonce
- GCM 提供认证加密（防篡改 + 防泄露）
- Argon2id 参数遵循 OWASP 推荐

---

## 5. 项目结构

```
cmd/
  pwm/                     # CLI 入口
    main.go                # cobra 命令注册
    add.go                 # add 子命令
    list.go                # list 子命令
    get.go                 # get 子命令
    search.go              # search 子命令
    delete.go              # delete 子命令
    generate.go            # generate 子命令
    export.go              # export 子命令
    import.go              # import 子命令
    strength.go            # strength 子命令
    init.go                # init 子命令
    rename.go              # rename 子命令

internal/
  crypto/                  # 加密层（纯算法，无业务逻辑）
    key.go                 # Argon2id 密钥派生
    aes.go                 # AES-256-GCM 加密/解密
  vault/                   # 保险库操作层（文件读写 + 加解密编排）
    vault.go               # Vault 结构体 + Load/Save 方法
    entry.go               # Entry 结构体定义
  service/                 # 业务逻辑层（不含文件 I/O）
    manager.go             # EntryManager: Add/Get/List/Search/Delete/Rename
    generator.go           # PasswordGenerator: 生成随机密码
    validator.go           # StrengthValidator: 密码强度检查
    exporter.go            # Exporter: CSV/JSON 导出
    importer.go            # Importer: CSV/JSON 导入
  config/                  # 配置
    config.go              # vault 文件路径等配置

go.mod
Makefile
```

### 5.1 依赖

| 依赖 | 用途 | 来源 |
|------|------|------|
| cobra | CLI 命令框架 | github.com/spf13/cobra |
| argon2 | 密钥派生 | golang.org/x/crypto/argon2 |
| AES/GCM | 加密 | crypto/aes + crypto/cipher (标准库) |
| csv | CSV 读写 | encoding/csv (标准库) |

### 5.2 分层原则

- **cmd 层**: 只做参数解析，调用 service 层方法
- **service 层**: 业务逻辑，不含文件 I/O，通过 vault 层接口操作数据
- **vault 层**: 文件读写 + 加解密编排，内部调用 crypto 层
- **crypto 层**: 纯加密算法函数，无业务逻辑

---

## 6. 业务逻辑细节

### 6.1 EntryManager

- **Add**: 检查 name 是否重复 → 添加新条目 → Save
- **Get**: 按 name 查找 → 返回完整条目
- **List**: 返回所有条目，可选按 tag 过滤（解密后内存过滤）
- **Search**: 按 keyword 在 name/username/site 中搜索（substring match）
- **Delete**: 按 name 查找 → 删除 → Save
- **Rename**: 检查 new name 是否重复 → 更新 name → Save

### 6.2 PasswordGenerator

- 默认长度 16，字符集: a-z, A-Z, 0-9, symbols (!@#$%^&*)
- `--no-symbols`: 只用 a-z, A-Z, 0-9
- `-l <length>`: 自定义长度
- 保证至少包含每种字符集各 1 个字符

### 6.3 StrengthValidator

评分维度:
- 长度 (< 8: weak, 8-12: medium, > 12: strong)
- 字符多样性 (小写/大写/数字/符号 各占权重)
- 评分输出: Very Weak / Weak / Medium / Strong / Very Strong

### 6.4 Exporter/Importer

- CSV 格式: name,username,site,password,tags,comment
- JSON 格式: 与 vault body entries 数组结构一致
- Import 时跳过重复 name 的条目（warn 提示）

---

## 7. 错误处理

| 场景 | 错误信息 | 退出码 |
|------|---------|--------|
| vault 文件不存在 | "Vault not initialized. Run `pwm init` first." | 1 |
| 主密码错误 | "Master password incorrect." | 2 |
| 条目名称重复 | "Entry '{name}' already exists." | 3 |
| 条目不存在 | "Entry '{name}' not found." | 4 |
| vault 文件损坏 | "Vault file corrupted or invalid format." | 5 |
| 导入格式错误 | "Invalid import file format." | 6 |

---

## 8. 测试策略

### 8.1 单元测试

| 模块 | 测试重点 | 目标覆盖率 |
|------|---------|-----------|
| crypto/key.go | Argon2id 派生正确性、参数边界 | ≥ 90% |
| crypto/aes.go | AES-256-GCM 加解密、nonce 不重复 | ≥ 90% |
| vault/vault.go | Load/Save 文件读写、Header 解析 | ≥ 80% |
| vault/entry.go | Entry 字段校验 | ≥ 80% |
| service/manager.go | CRUD、重复名称、空 vault | ≥ 80% |
| service/generator.go | 长度、字符集、不含指定字符集 | ≥ 85% |
| service/validator.go | 强弱密码评分、边界值 | ≥ 85% |
| service/exporter.go | CSV/JSON 格式正确性 | ≥ 80% |
| service/importer.go | CSV/JSON 解析、错误格式处理 | ≥ 80% |

### 8.2 E2E/集成测试

手动验证场景:
- 完整 CRUD 流程: init → add → get → list → search → delete
- 数据往返: export csv → import csv → 验证数据一致
- 主密码错误拒绝解密

### 8.3 Mock 策略

- crypto 层: 可用 mock 替换 Argon2id/AES 加速测试
- service 层: mock vault 层接口
- vault 层: 用临时目录而非真实 ~/.pwm/

### 8.4 覆盖率目标

Lines ≥ 80%, Branches ≥ 70%, Functions ≥ 80%

---

## 9. 开发里程碑

### Phase 1: 基础框架
- 项目初始化 + go.mod + Cobra 命令注册
- crypto 层 (Argon2id + AES-256-GCM)
- vault 层 (Load/Save + Header/Body)

### Phase 2: 核心 CRUD
- init 命令
- add / get / list / search / delete / rename 命令
- service/manager 层

### Phase 3: 扩展功能
- generate 命令 (密码生成器)
- strength 命令 (密码强度检查)
- export / import 命令

### Phase 4: 测试补齐 + 覆盖率达标
- 补充单元测试至覆盖率 ≥ 80%
- 集成测试验证
- E2E 手动验证