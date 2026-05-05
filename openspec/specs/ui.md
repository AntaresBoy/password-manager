# UI 设计规范 (UI Spec)

> **版本**: v1.0  
> **设计系统**: Seres Design System (SDS)  
> **适用框架**: Vue3 + Tailwind CSS + Figma  
> **最后更新**: 2026-05-05

---

## 1. 设计原则

1. **一致性**: 同一语义在不同场景表现一致（如「删除」永远是红色）
2. **反馈**: 每个用户操作必须有即时反馈（加载、成功、失败、进度）
3. **效率**: 高频操作 ≤ 2 步完成，关键信息 ≤ 3 秒获取
4. **容错**: 危险操作二次确认，可撤销 ≥ 30 秒
5. **无障碍**: 支持键盘全流程操作，WCAG 2.1 AA 级对比度

---

## 2. 设计 Token（Design Tokens）

### 2.1 色彩系统

#### 主色板（Primary）
| Token | 色值 | 用途 |
|-------|------|------|
| `--sd-primary-50`  | `#EFF6FF` | 背景高亮 |
| `--sd-primary-100` | `#DBEAFE` | Hover 背景 |
| `--sd-primary-500` | `#3B82F6` | 主按钮、链接 |
| `--sd-primary-600` | `#2563EB` | 主按钮 Hover |
| `--sd-primary-700` | `#1D4ED8` | 按下态 |

#### 功能色（Semantic）
| Token | 色值 | 用途 |
|-------|------|------|
| `--sd-success-500` | `#22C55E` | 成功、通过、在线 |
| `--sd-warning-500` | `#F59E0B` | 警告、待处理 |
| `--sd-error-500`   | `#EF4444` | 错误、删除、失败 |
| `--sd-info-500`    | `#3B82F6` | 提示、信息 |

#### 中性色（Neutral）
| Token | 色值 | 用途 |
|-------|------|------|
| `--sd-gray-900` | `#111827` | 主标题 |
| `--sd-gray-700` | `#374151` | 正文 |
| `--sd-gray-500` | `#6B7280` | 辅助文字 |
| `--sd-gray-300` | `#D1D5DB` | 边框、分割线 |
| `--sd-gray-100` | `#F3F4F6` | 背景 |
| `--sd-gray-50`  | `#F9FAFB` | 卡片背景 |

#### 暗黑模式映射
```css
/* tailwind.config.js 中配置 */
.dark {
  --sd-gray-900: #F9FAFB;  /* 反色 */
  --sd-gray-700: #E5E7EB;
  --sd-gray-100: #1F2937;
  --sd-gray-50:  #111827;
}
```

### 2.2 字体系统

| Token | 尺寸 | 字重 | 行高 | 用途 |
|-------|------|------|------|------|
| `--sd-text-xs`   | 12px | 400 | 16px  | 标签、时间戳 |
| `--sd-text-sm`   | 14px | 400 | 20px  | 辅助文字、按钮 |
| `--sd-text-base` | 16px | 400 | 24px  | 正文 |
| `--sd-text-lg`   | 18px | 500 | 28px  | 小标题 |
| `--sd-text-xl`   | 20px | 600 | 30px  | 模块标题 |
| `--sd-text-2xl`  | 24px | 700 | 32px  | 页面标题 |
| `--sd-text-3xl`  | 30px | 700 | 36px  | 大标题、数据指标 |

**字体栈**: `"Inter", "PingFang SC", "Microsoft YaHei", sans-serif`

### 2.3 间距系统（4px 网格）
```
--sd-space-1: 4px
--sd-space-2: 8px
--sd-space-3: 12px
--sd-space-4: 16px
--sd-space-5: 20px
--sd-space-6: 24px
--sd-space-8: 32px
--sd-space-10: 40px
--sd-space-12: 48px
```

**使用规范**: 禁止使用奇数像素间距（如 5px、7px），所有间距必须是 4 的倍数。

### 2.4 圆角系统
| Token | 值 | 用途 |
|-------|-----|------|
| `--sd-radius-sm`  | 4px  | 标签、小按钮 |
| `--sd-radius-md`  | 8px  | 输入框、卡片 |
| `--sd-radius-lg`  | 12px | 模态框、面板 |
| `--sd-radius-xl`  | 16px | 大卡片、Banner |
| `--sd-radius-full`| 9999px| 头像、Pill 标签 |

---

## 3. 组件规范

### 3.1 按钮（Button）

| 类型 | 背景 | 文字 | Hover | 禁用 |
|------|------|------|-------|------|
| Primary | `primary-500` | white | `primary-600` | `gray-300` |
| Secondary | `gray-100` | `gray-700` | `gray-200` | `gray-50` |
| Danger | `error-500` | white | `error-600` | `gray-300` |
| Ghost | transparent | `primary-500` | `primary-50` | `gray-300` |

**尺寸规范**:
- Small: 高 28px, 内边距 `px-3 py-1`, 文字 `text-sm`
- Medium: 高 36px, 内边距 `px-4 py-2`, 文字 `text-base`
- Large: 高 44px, 内边距 `px-6 py-3`, 文字 `text-lg`

**状态规范**:
- Loading: 显示 Spinner，文字保留，禁用点击
- Disabled: `opacity-50 cursor-not-allowed`，Tooltip 提示原因

### 3.2 输入框（Input）

```html
<!-- 标准输入框 -->
<div class="flex flex-col gap-1.5">
  <label class="text-sm font-medium text-gray-700">
    用户名 <span class="text-error-500">*</span>
  </label>
  <input 
    class="h-10 px-3 rounded-md border border-gray-300 
           focus:border-primary-500 focus:ring-2 focus:ring-primary-100
           disabled:bg-gray-50 disabled:text-gray-500
           dark:bg-gray-800 dark:border-gray-700"
    placeholder="请输入用户名"
  />
  <span class="text-xs text-gray-500">支持字母、数字，长度 4-20</span>
</div>
```

**状态规范**:
- Default: `border-gray-300`
- Focus: `border-primary-500 ring-2 ring-primary-100`
- Error: `border-error-500 ring-2 ring-error-100`，下方显示错误文案
- Disabled: `bg-gray-50 text-gray-500 cursor-not-allowed`
- ReadOnly: 与 Disabled 区分，使用 `bg-gray-50` 但保留文字色

### 3.3 表格（Table）

```html
<table class="w-full text-left border-collapse">
  <thead>
    <tr class="border-b border-gray-200 bg-gray-50">
      <th class="px-4 py-3 text-sm font-semibold text-gray-700">姓名</th>
      <th class="px-4 py-3 text-sm font-semibold text-gray-700">状态</th>
      <th class="px-4 py-3 text-sm font-semibold text-gray-700 text-right">操作</th>
    </tr>
  </thead>
  <tbody class="divide-y divide-gray-100">
    <tr class="hover:bg-gray-50 transition-colors">
      <td class="px-4 py-3 text-sm text-gray-900">张三</td>
      <td class="px-4 py-3">
        <span class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-success-100 text-success-700">
          正常
        </span>
      </td>
      <td class="px-4 py-3 text-right space-x-2">
        <button class="text-primary-600 hover:text-primary-700 text-sm">编辑</button>
        <button class="text-error-600 hover:text-error-700 text-sm">删除</button>
      </td>
    </tr>
  </tbody>
</table>
```

**规范**:
- 表头: `bg-gray-50`, 文字 `text-sm font-semibold text-gray-700`
- 行 Hover: `hover:bg-gray-50`
- 空状态: 居中显示插图 + "暂无数据" + 新建按钮
- 分页: 固定在表格底部右侧，页码最多显示 7 个

### 3.4 模态框（Modal）

**尺寸规范**:
- Small: 400px（确认类）
- Medium: 560px（表单类）
- Large: 720px（复杂表单/详情）
- Fullscreen: 100vw×100vh（全屏操作）

**交互规范**:
- 打开: 背景 `backdrop-blur-sm bg-black/50`，内容从底部滑入（mobile）或缩放淡入（desktop）
- 关闭: 点击遮罩、按 ESC、点击右上角 ×
- 堆叠: 禁止模态框上再开模态框，使用 Drawer 或页面跳转替代

### 3.5 消息提示（Message / Toast）

| 类型 | 位置 | 时长 | 动画 |
|------|------|------|------|
| Success | 顶部居中 | 3s | 下滑入，淡出 |
| Error | 顶部居中 | 5s | 下滑入，抖动后淡出 |
| Warning | 顶部居中 | 4s | 下滑入，淡出 |
| Info | 顶部居中 | 3s | 下滑入，淡出 |

**规范**:
- 最多同时显示 3 条，新消息顶掉最旧
- Error 类型必须提供「查看详情」按钮，展开错误堆栈
- 移动端宽度为 `calc(100vw - 32px)`，桌面端最大 400px

---

## 4. 布局与响应式

### 4.1 断点系统

| 断点 | 宽度 | 别名 | 典型设备 |
|------|------|------|---------|
| Mobile | < 640px | `sm` | 手机竖屏 |
| Tablet | 640px - 1024px | `md` / `lg` | 平板、手机横屏 |
| Desktop | 1024px - 1440px | `xl` | 笔记本 |
| Wide | ≥ 1440px | `2xl` | 外接显示器 |

### 4.2 布局模式

**管理后台（Dashboard）**:
```
+------------------------------------------+
|  Header (fixed, h-16)                    |
+----------+-------------------------------+
| Sidebar  |  Main Content                 |
| (w-64)   |  (flex-1, overflow-auto)      |
| fixed    |                               |
+----------+-------------------------------+
```
- Sidebar: 深色背景 `bg-gray-900`，菜单文字 `text-gray-300`，Hover `text-white bg-gray-800`
- 移动端: Sidebar 变为 Drawer，汉堡菜单触发

**数据看板（Data Viz）**:
- 卡片网格: `grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-4`
- 图表容器: 最小高度 300px，自适应宽度

### 4.3 安全区域
- 页面左右留白: `px-4 md:px-6 xl:px-8`
- 内容最大宽度: `max-w-7xl mx-auto`（1280px）
- 底部留白: 固定底部操作栏时，内容区底部 `pb-20`

---

## 5. 图标与插图

### 5.1 图标规范
- **图标库**: Heroicons（Outline 用于导航，Solid 用于按钮/状态）
- **尺寸**: 标准 4 档 —— 16px（内联）、20px（按钮）、24px（导航）、32px（空状态）
- **颜色**: 跟随父元素文字色 `currentColor`，禁止硬编码色值
- **描边**: Outline 图标统一 1.5px 描边

### 5.2 插图规范
- **空状态**: 使用品牌色插画，尺寸 120px-160px，下方标题 + 描述 + 操作按钮
- **错误页**: 500/404 使用动态插图，提供「返回首页」和「联系支持」双按钮

---

## 6. 动效与交互

### 6.1 缓动曲线
```css
--sd-ease-default: cubic-bezier(0.4, 0, 0.2, 1);      /* 150ms，常规过渡 */
--sd-ease-in:     cubic-bezier(0.4, 0, 1, 1);        /* 100ms，退出 */
--sd-ease-out:    cubic-bezier(0, 0, 0.2, 1);        /* 200ms，进入 */
--sd-ease-bounce: cubic-bezier(0.34, 1.56, 0.64, 1); /* 300ms，弹性（仅 Toast） */
```

### 6.2 过渡时长
| 场景 | 时长 | 说明 |
|------|------|------|
| 按钮 Hover | 150ms | 背景色、边框色 |
| 卡片 Hover | 200ms | 阴影、位移 `translateY(-2px)` |
| 模态框打开 | 300ms | 遮罩淡入 + 内容缩放 |
| 页面切换 | 200ms | 淡入淡出 |
| 数据加载 | 无限循环 | Skeleton 脉冲动画 1.5s |
| Toast | 300ms | 进入；退出 200ms |

### 6.3 微交互
- **按钮按下**: `scale(0.98)`，时长 100ms
- **输入框聚焦**: 边框色变化 + 外发光 `ring-2`
- **开关（Switch）**: 圆形滑块平移 + 背景色渐变
- **下拉菜单**: 列表项依次滑入，间隔 30ms

---

## 7. 无障碍（a11y）

### 7.1 键盘导航
- 所有可交互元素必须可通过 `Tab` 聚焦
- 聚焦环: `focus-visible:ring-2 focus-visible:ring-primary-500 focus-visible:ring-offset-2`
- 快捷键: `?` 键呼出快捷键面板，快捷键不与系统冲突

### 7.2 屏幕阅读器
- 图标按钮必须有 `aria-label`
- 表单输入必须有关联 `<label>` 或 `aria-labelledby`
- 动态更新区域使用 `aria-live="polite"`
- 模态框打开时焦点锁定，关闭后返回触发元素

### 7.3 对比度
- 正文与背景对比度 ≥ 4.5:1
- 大文字（18px+）与背景对比度 ≥ 3:1
- 错误/警告信息不仅依赖颜色，必须有图标辅助

---

## 8. Figma 协作规范

### 8.1 文件组织
```
📁 AI Nexus Platform
├── 📁 01-Design Tokens
│   └── Color / Typography / Shadow / Spacing
├── 📁 02-Components
│   ├── 📁 Base（Button, Input, Select...）
│   ├── 📁 Composite（Table, Form, Chart...）
│   └── 📁 Business（UserCard, OrderList...）
├── 📁 03-Patterns
│   ├── Empty State / Loading / Error
│   └── Layout Templates
├── 📁 04-Pages
│   ├── Login / Dashboard / Settings...
└── 📁 05-Archive
```

### 8.2 命名规范
- 页面: `[模块] 页面名 / 状态` → `[用户] 登录页 / 错误提示`
- 组件: `类型/名称/变体` → `Button/Primary/Large/Loading`
- 图层: 语义化命名，禁止 `Frame 127` → 改为 `Card Container`

### 8.3 交付标准
- 设计稿必须包含「暗黑模式」画板
- 所有组件必须链接到 Design Token，禁止硬编码色值
- 导出标注必须包含: 尺寸、间距、颜色 Token、字体 Token、动效时长
- 图标以 SVG 导出，插图以 2x PNG / WebP 导出

---

## 9. web-design-guidelines Skill 集成

当 AI 执行 UI 相关变更时，必须：
1. 激活 `web-design-guidelines` Skill 进行设计审查
2. 检查是否符合本规范中的 Token 和组件定义
3. 输出设计审查报告，包含：
   - 使用的 Token 列表
   - 组件符合性检查
   - 响应式适配建议
   - 无障碍合规性检查
