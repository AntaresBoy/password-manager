# 代码提交规范 (Git Commit Spec)

> **版本**: v1.0  
> **规范标准**: Conventional Commits 1.0.0 + OpenSpec 扩展  
> **适用范围**: 赛力斯 AI DevOps 平台所有工程项目  
> **最后更新**: 2026-05-05

---

## 1. 分支管理规范

### 1.1 分支模型（Git Flow 简化版）

```
main (protected) ──────── 生产环境，仅接受 PR/MR
    │
    ├── develop ───────── 日常开发基线，合并特性分支
    │       │
    │       ├── feature/login-form ──── 特性分支（OpenSpec 变更）
    │       ├── feature/ai-review
    │       └── hotfix/auth-bug ─────── 热修复分支
    │
    └── release/v1.2.0 ── 预发布分支（可选）
```

### 1.2 分支命名规范

| 分支类型 | 命名格式 | 示例 |
|----------|----------|------|
| 特性开发 | `feature/{openspec-change-name}` | `feature/login-form` |
| 缺陷修复 | `fix/{issue-id}-{description}` | `fix/1042-memory-leak` |
| 热修复 | `hotfix/{description}` | `hotfix/auth-bypass` |
| 重构 | `refactor/{scope}-{description}` | `refactor/api-error-handling` |
| 文档 | `docs/{description}` | `docs/deploy-guide` |
| 测试 | `test/{scope}-{description}` | `test/e2e-checkout-flow` |

**禁止**:
- ❌ 使用个人姓名作为分支名：`zhangsan-fix`
- ❌ 使用无意义数字：`branch-1`, `fix-2`
- ❌ 直接在 `main` 或 `develop` 上提交代码

### 1.3 分支生命周期
1. 从 `develop` 切出特性分支
2. 开发完成后，本地执行完整测试与 Lint
3. 提交 PR/MR 到 `develop`，触发 CI 门禁
4. Code Review 通过且 CI 全绿后合并
5. 合并后删除远程特性分支

---

## 2. Commit Message 格式（Conventional Commits）

### 2.1 标准格式

```
<type>(<scope>): <subject>

<body>

<footer>
```

- **type**: 提交类型（见 2.2）
- **scope**: 影响范围（见 2.3）
- **subject**: 简短描述（不超过 50 字符）
- **body**: 详细说明（可选，每行不超过 72 字符）
- **footer**: 关联 Issue、Breaking Change、OpenSpec 元数据

### 2.2 提交类型（Type）

| Type | 含义 | 使用场景 | CI 版本号影响 |
|------|------|----------|--------------|
| `feat` | 新功能 | 新增业务功能、页面、API | MINOR ↑ |
| `fix` | 缺陷修复 | 修复 Bug、线上问题 | PATCH ↑ |
| `docs` | 文档 | 修改 README、注释、API 文档 | 无 |
| `style` | 代码风格 | 格式化、分号、空格、变量名（不影响逻辑） | 无 |
| `refactor` | 重构 | 代码重构，既不修复 Bug 也不新增功能 | 无 |
| `perf` | 性能优化 | 提升性能、减少资源占用 | PATCH ↑ |
| `test` | 测试 | 新增或修改测试代码 | 无 |
| `chore` | 构建/工具 | 依赖升级、CI 配置、脚本调整 | 无 |
| `ci` | CI 相关 | 修改 CI/CD 流水线配置 | 无 |
| `revert` | 回滚 | 撤销某次提交 | PATCH ↑ |
| `build` | 构建系统 | 影响构建或外部依赖（如 webpack、vite） | 无 |

**禁止**:
- ❌ 使用 `update`, `modify`, `change` 等模糊类型
- ❌ 一个 Commit 包含多种类型变更（应拆分）

### 2.3 影响范围（Scope）

| Scope | 说明 | 示例 |
|-------|------|------|
| `ui` | UI 组件、样式、交互 | `feat(ui): 新增 DatePicker 组件` |
| `api` | 后端接口、路由、Handler | `fix(api): 修复订单查询空指针` |
| `db` | 数据库、模型、迁移脚本 | `feat(db): 用户表增加手机号字段` |
| `auth` | 认证、授权、JWT | `fix(auth): 刷新 Token 并发安全问题` |
| `test` | 测试代码、Mock 数据 | `test(unit): 补充用户服务单元测试` |
| `ci` | CI/CD、构建脚本 | `chore(ci): 增加覆盖率上报步骤` |
| `deps` | 依赖升级 | `chore(deps): 升级 Vue 3.4 → 3.5` |
| `docs` | 文档、README | `docs(api): 补充登录接口 Swagger 注释` |
| `config` | 配置文件 | `chore(config): 调整 ESLint 规则` |
| `ops` | 运维、部署、Docker | `build(ops): Dockerfile 多阶段构建优化` |

**自定义 Scope**: 允许使用模块名作为 scope，如 `feat(order): 创建订单接口`。

---

## 3. OpenSpec 扩展格式（强制）

与 OpenSpec + Superpowers 流程结合，Footer 必须包含以下元数据：

### 3.1 标准 Footer 模板

```
feat(auth): 实现用户登录与 JWT 鉴权

- 新增 /api/v1/login 接口，支持手机号+密码登录
- 新增 JWT 签发与刷新机制，Access Token 15min + Refresh Token 7d
- 新增登录失败计数，5 次错误后锁定 30 分钟

Coverage: Lines 82.3%, Branches 74.1%, Functions 85.0%
Review: 两阶段审查通过，1 Major 问题已修复
OpenSpec: feature/login-form-20260505
Breaking: 旧版 /api/v1/auth 接口已废弃，请迁移至 /api/v1/login
Refs: #1042, #1056
```

### 3.2 Footer 字段规范

| 字段 | 格式 | 必填 | 说明 |
|------|------|------|------|
| `Coverage` | `Lines X%, Branches X%, Functions X%` | ✅ 是 | 前端/后端覆盖率数据 |
| `Review` | `两阶段审查通过 / 未通过` | ✅ 是 | Code Review 结论 |
| `OpenSpec` | `{change-name}-{YYYYMMDD}` | ✅ 是 | 关联的 OpenSpec 变更目录名 |
| `Breaking` | `描述` | 条件 | 不兼容变更必须声明 |
| `Refs` | `#issue-id, #issue-id` | 条件 | 关联的 Issue/需求单 |
| `Closes` | `#issue-id` | 条件 | 修复的 Bug，合并后自动关闭 |

### 3.3 覆盖率数据格式

```
# 前端项目
Coverage: Lines 82.3%, Branches 74.1%, Functions 85.0%

# 后端项目（如使用 go test -cover）
Coverage: pkg/service 82%, pkg/repository 65%, overall 78%

# 混合项目
Coverage: FE(Lines 82%, Branches 74%) / BE(Service 85%, Overall 78%)
```

**未达标时的 Commit 规范**:
```
feat(ui): 实现仪表盘数据卡片

- 新增 DashboardCard 组件，支持 4 种指标类型
- 新增 useDashboard 数据获取 Composable

Coverage: Lines 76.5% ❌ (Threshold: 80%)
Review: 两阶段审查通过
OpenSpec: feature/dashboard-card-20260505
Note: 覆盖率未达标，已记录技术债务 #TD-003，计划 Sprint 3 补充 E2E 测试
```

---

## 4. Commit 内容规范

### 4.1 Subject 行规范

```
# ✅ 正确
feat(auth): 新增手机号验证码登录
fix(api): 修复订单分页查询总数错误
refactor(db): 将用户密码加密逻辑提取为独立服务
perf(ui): 优化表格大数据渲染性能

# ❌ 错误
feat: update                          # 无 scope，无意义描述
feat: 修改了一些代码                  # 过于笼统
fix: bug fix                          # 无信息量
feat(auth): add login function        # 中英混杂，且 function 冗余
```

**Subject 红线**:
- 不超过 50 个字符（含 type/scope）
- 使用祈使句、现在时（"新增" 而非 "新增了"）
- 首字母无需大写（除非是专有名词）
- 末尾不加句号

### 4.2 Body 行规范

```
feat(api): 实现文件上传与分片存储

- 新增 /api/v1/upload 接口，支持 multipart/form-data
- 实现分片上传逻辑，单文件最大 2GB，分片大小 5MB
- 使用 OSS 预签名 URL 直传，降低服务器带宽压力
- 上传完成后触发异步任务合并分片并校验 MD5

- 新增 UploadService 处理上传业务逻辑
- 新增 ChunkRepository 管理分片元数据
- 新增 UploadTask 异步任务处理器

Coverage: Lines 88.2%, Branches 81.0%, Functions 90.0%
Review: 两阶段审查通过
OpenSpec: feature/file-upload-20260505
Refs: #1102
```

**Body 规范**:
- 每行不超过 72 字符
- 使用项目符号列表（`-`）说明变更点
- 说明「为什么」而非仅「做了什么」
- 破坏性变更必须在 Body 顶部用 `BREAKING CHANGE:` 标注

### 4.3 提交粒度

| 场景 | 建议 | 示例 |
|------|------|------|
| 新增功能 | 一个 Commit 对应一个独立功能点 | `feat(auth): 新增 JWT 签发` |
| 修复 Bug | 一个 Commit 对应一个 Bug | `fix(api): 修复空指针 #1042` |
| 代码审查反馈 | 单独 Commit，便于 Reviewer 追踪 | `fix(ui): 调整按钮间距（Code Review 反馈）` |
| 大型重构 | 拆分为多个 Commit，每个一个子步骤 | `refactor(db): 提取用户 Repository` → `refactor(db): 替换旧 DAO 调用` |

**禁止**:
- ❌ 一个 Commit 包含 3 个以上无关文件变更
- ❌ 提交未编译/未测试通过的代码
- ❌ 提交包含敏感信息（密码、Token、密钥）

---

## 5. 提交前检查清单（Pre-Commit）

### 5.1 本地检查命令

```bash
# 1. 代码格式
npm run lint          # 前端 ESLint + Prettier
gofmt -w .            # 后端格式化
golangci-lint run     # 后端 Lint

# 2. 类型检查
npm run type-check    # 前端 TypeScript

# 3. 测试
npm run test:unit     # 前端单元测试
go test ./... -cover  # 后端单元测试

# 4. 覆盖率检查
jq '.total.lines.pct' coverage/coverage-summary.json
# 必须 >= 80%

# 5. OpenSpec 变更检查
openspec validate     # 验证配置与规范文件
ls openspec/changes/  # 确认当前变更目录存在
```

### 5.2 Git Hook 脚本（.git/hooks/pre-commit）

```bash
#!/bin/bash
# .git/hooks/pre-commit
# 安装方式: cp scripts/pre-commit.sh .git/hooks/pre-commit && chmod +x .git/hooks/pre-commit

set -e

echo "🔍 Pre-commit checks starting..."

# 1. 检查 Commit Message 格式（使用 commitlint）
# 需安装: npm install -g @commitlint/config-conventional @commitlint/cli
if command -v commitlint &> /dev/null; then
    echo "Checking commit message format..."
    cat $1 | commitlint
fi

# 2. 前端 Lint
if [ -f "package.json" ]; then
    echo "Running ESLint..."
    npm run lint

    echo "Running TypeScript check..."
    npm run type-check

    echo "Running unit tests..."
    npm run test:unit -- --run
fi

# 3. 后端 Lint
if [ -f "go.mod" ]; then
    echo "Running golangci-lint..."
    golangci-lint run

    echo "Running Go tests..."
    go test ./... -coverprofile=coverage.out

    # 检查覆盖率（示例阈值 70%）
    COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
    if (( $(echo "$COVERAGE < 70" | bc -l) )); then
        echo "❌ Go test coverage $COVERAGE% is below 70%"
        exit 1
    fi
fi

# 4. 检查敏感信息泄露
echo "Checking for secrets..."
if git diff --cached --name-only | xargs grep -lE "(password|secret|token|key)\s*[:=]\s*["'][^"']{8,}["']" 2>/dev/null; then
    echo "⚠️  Potential secret detected in staged files!"
    exit 1
fi

echo "✅ All pre-commit checks passed!"
```

### 5.3 检查清单（Checklist）

每次 `git commit` 前必须确认：

- [ ] **变更范围**: 本次 Commit 只包含一个功能点或修复点
- [ ] **Commit Message**: 符合 Conventional Commits 格式，包含 OpenSpec Footer
- [ ] **Lint 通过**: ESLint / golangci-lint 无 Error
- [ ] **类型检查**: TypeScript / Go build 通过
- [ ] **单元测试**: 全部通过，无 `.skip` 或 `.only`
- [ ] **覆盖率**: 前端 Lines ≥ 80%，后端 Service ≥ 80%
- [ ] **E2E 测试**: 关键路径已覆盖（如适用）
- [ ] **OpenSpec 关联**: Footer 包含 `OpenSpec: {change-name}`
- [ ] **敏感信息**: 无密码、Token、密钥、内网 IP 硬编码
- [ ] **调试代码**: 无 `console.log`, `debugger`, `fmt.Println` 调试残留
- [ ] **文档更新**: 如变更接口或组件，README / Storybook / Swagger 已同步

---

## 6. PR / MR 规范

### 6.1 标题格式
```
[Type/Scope] 简短描述 | OpenSpec: {change-name}

# 示例
[feat/auth] 实现 JWT 登录与鉴权 | OpenSpec: feature/login-form-20260505
[fix/api] 修复订单查询并发问题 | OpenSpec: fix/order-race-20260505
```

### 6.2 PR 描述模板

```markdown
## 变更摘要
- 新增 / 修复 / 重构了 XXX 功能
- 影响范围：前端 UI / 后端 API / 数据库 / 部署配置

## 关联规范
- OpenSpec 变更: `openspec/changes/feature-xxx/`
- 前端规范: `fe.md`
- 测试规范: `fe-testing.md`

## 测试验证
- [ ] 单元测试通过（覆盖率: Lines __%, Branches __%）
- [ ] E2E 测试通过（关键路径: __）
- [ ] 手动验证通过（验证项: __）

## Code Review
- [ ] 两阶段审查通过
- [ ] 无 Critical 问题遗留
- [ ] Major 问题已记录或修复

## 破坏性变更
- [ ] 无破坏性变更
- [ ] 有破坏性变更，已声明: __

## 部署注意
- [ ] 需执行数据库迁移
- [ ] 需更新环境变量
- [ ] 需通知下游服务方
```

### 6.3 合并策略
- **Squash Merge**: 特性分支合并到 `develop` 时使用，保持主线整洁
- **Merge Commit**: `develop` 合并到 `main` 时使用，保留完整历史
- **禁止**: Fast-forward 合并（丢失分支上下文）

---

## 7. 版本发布规范

### 7.1 版本号规则（Semantic Versioning）
```
主版本号.次版本号.修订号
X.Y.Z
```

| 版本号变化 | 触发条件 | Commit Type |
|-----------|---------|-------------|
| X（主版本） | 不兼容 API 变更 | `feat` + `BREAKING CHANGE` |
| Y（次版本） | 向下兼容的功能新增 | `feat` |
| Z（修订号） | 向下兼容的问题修复 | `fix` / `perf` |

### 7.2 版本标签
```bash
# 发布时打标签
git tag -a v1.2.0 -m "Release v1.2.0 - 新增 AI Review 模块"
git push origin v1.2.0

# 标签格式
v{主版本}.{次版本}.{修订号}[-{预发布标识}]
# 示例: v1.2.0, v1.2.0-beta.1, v1.2.0-hotfix.1
```

### 7.3 CHANGELOG 生成
使用 `standard-version` 或 `semantic-release` 自动生成：
```bash
npm run release
# 自动生成 CHANGELOG.md + 版本号提升 + Git Tag
```

---

## 8. 回滚规范

### 8.1 回滚场景
| 场景 | 操作 | Commit Message |
|------|------|----------------|
| 刚提交到本地，未 Push | `git reset --soft HEAD~1` | 无需 Commit |
| 已 Push 到远程，未合并 | `git revert <commit-hash>` | `revert(scope): 撤销 XXX` |
| 已合并到 develop | `git revert -m 1 <merge-commit>` | `revert: 撤销 Merge PR #123` |
| 已发布到生产 | 热修复分支 + 紧急发布 | `hotfix(scope): 修复 XXX 回滚问题` |

### 8.2 Revert Commit 格式
```
revert(auth): 撤销 "feat(auth): 新增 JWT 签发"

This reverts commit abc1234.

原因: 生产环境 JWT 验证导致旧版 App 无法登录，
      需等待 App 强制升级后再发布。

Refs: #1102
```

---

## 9. 工具链配置

### 9.1 commitlint 配置（commitlint.config.js）
```js
module.exports = {
  extends: ['@commitlint/config-conventional'],
  rules: {
    'type-enum': [2, 'always', [
      'feat', 'fix', 'docs', 'style', 'refactor', 'perf', 'test', 'chore', 'ci', 'revert', 'build'
    ]],
    'scope-enum': [1, 'always', [
      'ui', 'api', 'db', 'auth', 'test', 'ci', 'deps', 'docs', 'config', 'ops'
    ]],
    'subject-max-length': [2, 'always', 50],
    'body-max-line-length': [2, 'always', 72],
    'footer-max-line-length': [2, 'always', 100],
    // 自定义规则：必须包含 OpenSpec Footer
    'footer-leading-blank': [2, 'always'],
  },
  parserPreset: {
    parserOpts: {
      issuePrefixes: ['#', 'OPENSPEC-', 'TD-']
    }
  }
}
```

### 9.2 package.json 脚本
```json
{
  "scripts": {
    "commit": "git-cz",
    "lint:commit": "commitlint --from=HEAD~1",
    "release": "standard-version",
    "postinstall": "husky install"
  },
  "devDependencies": {
    "@commitlint/cli": "^19.0.0",
    "@commitlint/config-conventional": "^19.0.0",
    "commitizen": "^4.3.0",
    "cz-conventional-changelog": "^3.3.0",
    "husky": "^9.0.0",
    "standard-version": "^9.5.0"
  },
  "config": {
    "commitizen": {
      "path": "cz-conventional-changelog"
    }
  }
}
```

### 9.3 Husky 配置（v9+）
```bash
# .husky/pre-commit
npm run lint
npm run test:unit -- --run

# .husky/commit-msg
npx --no-install commitlint --edit ${1}

# .husky/pre-push
npm run test:e2e:ci
```

---

## 10. 示例集

### 10.1 完整示例：特性开发
```
feat(auth): 实现基于手机号的验证码登录

- 新增 /api/v1/login/sms 接口，支持手机号 + 验证码登录
- 新增 SMS Service，集成阿里云短信服务
- 新增验证码缓存逻辑，5 分钟有效，最多重发 3 次
- 前端新增 LoginBySms.vue 组件，支持验证码倒计时

BREAKING CHANGE: 旧版 /api/v1/login/password 接口返回结构变更，
                 新增 `loginType` 字段区分登录方式。

Coverage: FE(Lines 85%, Branches 78%) / BE(Service 88%, Overall 82%)
Review: 两阶段审查通过，无遗留问题
OpenSpec: feature/sms-login-20260505
Refs: #1102, #1103
Closes: #1042
```

### 10.2 完整示例：缺陷修复
```
fix(api): 修复订单查询并发下数据不一致问题

- 在 OrderRepository.List 中增加 FOR UPDATE 行锁
- 修复库存扣减与订单创建非原子操作导致的超卖
- 新增分布式锁（Redis）防止并发创建重复订单

Coverage: Lines 81.2%, Branches 75.0%, Functions 83.5%
Review: 两阶段审查通过，1 Major 问题已修复（锁超时时间调整）
OpenSpec: fix/order-race-20260505
Refs: #1156
```

### 10.3 完整示例：测试补充
```
test(auth): 补充登录模块单元测试与 E2E 场景

- 新增 useAuth Composable 测试（覆盖登录、登出、Token 刷新）
- 新增 LoginForm 组件测试（Props、Events、校验逻辑）
- 新增 E2E 场景：验证码错误 3 次后锁定、Token 过期自动跳转

Coverage: Lines 从 65% 提升至 82%
Review: 两阶段审查通过
OpenSpec: feature/sms-login-20260505
Note: 本次提交仅补充测试，无业务逻辑变更
```

### 10.4 完整示例：热修复
```
hotfix(auth): 修复 JWT 验证绕过漏洞

- 修复中间件中 JWT 过期时间校验逻辑错误
- 新增 Token 黑名单校验，防止已登出 Token 复用

Coverage: Lines 79.0%（热修复豁免，Sprint 3 补充测试）
Review: 紧急审查通过（1 位架构师 + 1 位安全工程师）
OpenSpec: hotfix/jwt-bypass-20260505
Refs: #P0-20260505
```

---

## 11. 禁止清单

| 禁止项 | 正确做法 |
|--------|---------|
| `git commit -m "update"` | 使用完整格式：`feat(ui): 新增 Dashboard 数据卡片` |
| 一个 Commit 改 20 个文件 | 拆分为多个原子 Commit |
| Commit 中硬编码密码 | 使用环境变量 + `.env.example` |
| 提交未测试代码 | 本地运行完整测试套件 |
| 提交包含 `.only` 的测试 | 删除 `.only` 后提交 |
| 提交后手动修改 CI 配置绕过门禁 | 修复代码直至 CI 自然通过 |
| 使用 `merge` 提交到 main | 使用 PR/MR + Squash Merge |
| 删除他人 Commit 历史 | 使用 `git revert` 保留记录 |
| 提交二进制文件（图片、视频）到 Git | 使用 LFS 或 CDN |

---

**附录: 快速查询卡**

```bash
# 提交命令速查
git add -p                              # 交互式选择变更块
git commit -m "feat(ui): 新增 XXX"      # 快速提交（仅适用于简单变更）
git commit                              # 使用编辑器编写完整 Message

# 查看提交历史
git log --oneline --graph               # 图形化历史
git log --grep="feat(auth)"             # 按类型筛选

# 修改最后一次提交
git commit --amend                      # 修改 Message 或追加文件

# 交互式 Rebase（整理提交）
git rebase -i HEAD~3                    # 合并、修改、删除最近 3 个提交
```
