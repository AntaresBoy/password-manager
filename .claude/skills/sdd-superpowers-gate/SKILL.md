---
name: sdd-superpowers-gate
description: OpenSpec + Superpowers 整合门禁：TDD + Code Review + 覆盖率 + 飞书通知
tools: Read, Edit, Bash
---

当你收到任何开发请求时，按以下**强制流程**执行，**禁止跳过任何步骤**：

## Phase 1: OpenSpec 需求治理
1. 读取 `@openspec/config.yaml` 的 `context`
2. 读取相关规范文件（根据变更范围选择 fe.md / be.md / ui.md / fe-testing.md）
3. 使用 `/opsx:propose` 或手动创建 `openspec/changes/{name}/proposal.md`
4. 在 proposal 中明确：
   - 需要修改的文件清单
   - 测试策略（单元测试文件列表 + E2E 场景列表）
   - 预期覆盖率目标（≥ 80%）

## Phase 2: Superpowers 纪律执行
5. **Brainstorming**: 激活 `brainstorming` skill，澄清需求边界和边界情况
6. **Planning**: 激活 `writing-plans` skill，将 proposal 拆分为 2-5 分钟的细粒度任务
7. **Git Worktree**: 激活 `using-git-worktrees` skill，创建独立分支隔离开发

## Phase 3: 子代理开发 + TDD + 审查（循环执行）
对每个子任务，派遣独立子代理执行：

### 3a. TDD 铁律（test-driven-development skill）
- 子代理必须先写失败测试，再写最小实现
- 实现代码必须通过测试后才允许继续
- 禁止先写业务代码再补测试[^81^]

### 3b. 两阶段代码审查（requesting-code-review skill）
每个子任务完成后，**必须**触发审查：
- **阶段 1 - 规格合规审查**: 检查实现是否符合 `openspec/changes/{name}/proposal.md` 中的计划
- **阶段 2 - 代码质量审查**: 检查代码风格、安全漏洞、性能问题、规范符合性
- **审查结果分级**:
  - 🔴 Critical: 阻塞，必须修复后才能继续
  - 🟠 Major: 建议修复，不阻塞但需记录
  - 🟢 Minor: 提示性[^70^]

### 3c. 覆盖率门禁
- 运行 `npm run test:unit -- --coverage`
- 检查 `coverage-summary.json`:
  - lines.pct &gt;= 80
  - branches.pct &gt;= 70
  - functions.pct &gt;= 80
- **未达标**: 禁止进入下一个子任务，补充测试用例直至达标

## Phase 4: 验证与报告（OpenSpec verify）
8. 所有子任务完成后，进入 `verify` 阶段：
   - 运行全量测试：`npm run test`（单元 + E2E）
   - 生成 HTML 报告到 `openspec/reports/test-report-{timestamp}.html`
   - 生成 JSON 摘要到 `openspec/reports/coverage-summary.json`
   - 在 `openspec/changes/{name}/` 下创建 `verify.md` 记录验证结果[^68^]

## Phase 5: 收尾与通知
9. **OpenSpec Archive**: 执行 `/opsx:archive` 归档变更
10. **自动提交**:
    ```bash
    git add .
    git commit -m "feat(scope): 描述 [coverage: XX%] [review: passed]"
    git push
    ```
11. **飞书通知**（无论成功/失败）:
    - 成功: 发送「变更完成」卡片，包含覆盖率数据、报告链接、审查摘要
    - 失败: 发送「变更失败」卡片，包含失败原因、日志片段、需人工介入标记
    - 调用项目根目录的 `openspec-notify.sh` 或配置好的 webhook

## 禁止事项
- ❌ 禁止跳过 brainstorming 直接写代码
- ❌ 禁止跳过 requesting-code-review 直接合并
- ❌ 禁止修改 `vitest.config.ts` 降低覆盖率阈值
- ❌ 禁止删除已有测试用例「美化」覆盖率
- ❌ 禁止在 Critical 审查未修复时继续执行