# 前端开发规范 (FE Spec)

> **版本**: v1.0  
> **适用技术栈**: Vue 3 + TypeScript + Vite  
> **配套工具**: ESLint + Prettier + Tailwind CSS + Pinia  
> **最后更新**: 2026-05-05

---

## 1. 技术栈与版本锁定

| 依赖 | 版本要求 | 说明 |
|------|---------|------|
| Vue | ^3.4.x | 必须使用 `<script setup>` + Composition API |
| TypeScript | ^5.4.x | `strict: true`，禁止 `any` |
| Vite | ^5.x | 构建工具 |
| Vue Router | ^4.x | 路由管理 |
| Pinia | ^2.x | 状态管理（唯一选择，禁止 Vuex） |
| Tailwind CSS | ^3.4.x | 原子化 CSS（唯一选择，禁止手写 CSS 文件） |
| Axios | ^1.6.x | HTTP 客户端 |
| Vitest | ^1.x | 单元测试框架 |
| Cypress | ^13.x | E2E 测试框架 |

---

## 2. 项目目录结构

```
src/
├── api/                    # API 接口层（按模块组织）
│   ├── user.ts
│   └── types/              # API 类型定义
├── assets/                 # 静态资源（图片、字体）
│   ├── images/
│   └── icons/
├── components/             # 组件
│   ├── __ui__/             # 基础 UI 组件（Button, Input, Modal）
│   ├── __layout__/         # 布局组件（Header, Sidebar, Footer）
│   └── business/           # 业务组件（UserCard, OrderList）
├── composables/            # 组合式函数
│   ├── useAuth.ts
│   └── usePermission.ts
├── directives/             # 自定义指令
├── layouts/                # 页面级布局
├── plugins/                # 插件注册
├── router/                 # 路由配置
│   ├── index.ts
│   └── modules/            # 按业务模块拆分
├── stores/                 # Pinia Store
│   ├── user.ts
│   └── app.ts
├── styles/                 # 全局样式（仅允许 tailwind.css + variables）
├── utils/                  # 工具函数
│   ├── request.ts          # Axios 封装
│   ├── validator.ts        # 表单校验
│   └── format.ts           # 格式化
├── views/                  # 页面视图
│   ├── login/
│   └── dashboard/
└── App.vue
```

**禁止**: 在 `src/` 根目录下直接创建零散文件；所有代码必须有明确归属目录。

---

## 3. 代码风格与格式

### 3.1 ESLint 规则（不可修改）
使用项目根目录 `.eslintrc.cjs`，核心规则：
- `@typescript-eslint/no-explicit-any`: `error` —— 禁止 `any`
- `@typescript-eslint/explicit-function-return-type`: `warn` —— 公共函数必须声明返回类型
- `vue/multi-word-component-names`: `off` —— 允许单单词组件名（仅限 `__ui__`）
- `vue/require-default-prop`: `error` —— 可选 Prop 必须有默认值
- `import/order`: `error` —— Import 强制分组排序
- `no-console`: `warn` —— 生产环境禁止 `console.log`

### 3.2 Import 排序（强制）
```ts
// 1. builtin
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'

// 2. external
import axios from 'axios'
import { ElMessage } from 'element-plus'

// 3. internal（别名 @/）
import { useUserStore } from '@/stores/user'
import { request } from '@/utils/request'

// 4. sibling（相对路径同级）
import { validatePhone } from './validator'

// 5. index（相对路径父级）
import type { UserInfo } from '../types'
```

### 3.3 Prettier 配置
```json
{
  "semi": false,
  "singleQuote": true,
  "tabWidth": 2,
  "trailingComma": "es5",
  "printWidth": 100,
  "arrowParens": "avoid"
}
```

---

## 4. Vue 组件规范

### 4.1 组件结构（强制 `<script setup>`）
```vue
<script setup lang="ts">
// 1. 类型导入
import type { PropType } from 'vue'

// 2. Vue / 外部库
import { ref, computed, watch } from 'vue'
import { useRoute } from 'vue-router'

// 3. 内部模块
import { useUserStore } from '@/stores/user'
import { formatDate } from '@/utils/format'

// 4. 同级模块
import { useForm } from './composables/useForm'

// 5. Props & Emits（显式定义）
interface Props {
  title: string
  visible?: boolean
}
const props = withDefaults(defineProps<Props>(), {
  visible: false,
})

const emit = defineEmits<{
  confirm: [value: string]
  cancel: []
}>()

// 6. 组合式函数
const route = useRoute()
const userStore = useUserStore()

// 7. 响应式状态（按用途分组）
const loading = ref(false)
const formData = ref({ name: '', age: 0 })

// 8. 计算属性
const displayTitle = computed(() => props.title || '默认标题')

// 9. 方法
async function handleSubmit() {
  // ...
}

// 10. 生命周期 & Watch
watch(() => props.visible, (val) => {
  if (val) initForm()
})
</script>

<template>
  <div class="flex flex-col gap-4 p-6">
    <h1 class="text-xl font-bold text-gray-900">{{ displayTitle }}</h1>
    <!-- 内容 -->
  </div>
</template>
```

### 4.2 命名规范
| 类型 | 规范 | 示例 |
|------|------|------|
| 组件文件 | PascalCase | `UserProfile.vue` |
| 组件目录 | kebab-case | `user-profile/` |
| 组合式函数 | `use` + camelCase | `useAuth.ts` |
| Props | camelCase | `userName` |
| Event | camelCase | `@userChange` |
| CSS Class | Tailwind 工具类为主 | `class="flex items-center gap-2"` |
| 自定义指令 | `v` + 小写 | `vPermission` |

### 4.3 组件设计原则
- **单一职责**: 一个组件只做一件事，超过 200 行必须拆分
- **Props 向下**: 数据通过 Props 传递，禁止跨层级直接访问 Store
- **Events 向上**: 事件通过 `$emit` 向上传递，禁止子组件直接修改父组件状态
- **Slots 优先**: 可定制区域优先使用 Slots，而非 Props 传递 HTML 字符串
- **异步边界**: 所有异步操作必须包裹 `try/catch`，错误通过统一错误处理上报

---

## 5. TypeScript 规范

### 5.1 严格模式红线
```json
// tsconfig.json
{
  "compilerOptions": {
    "strict": true,
    "noImplicitAny": true,
    "strictNullChecks": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true
  }
}
```

### 5.2 类型定义规范
```ts
// ✅ 正确：接口命名使用 I 前缀或语义化命名
interface IUserInfo {
  id: number
  name: string
  avatar?: string  // 可选属性
}

// ✅ 正确：类型别名用于联合类型
type Status = 'pending' | 'success' | 'error'

// ✅ 正确：枚举用于固定取值
enum OrderStatus {
  Pending = 0,
  Paid = 1,
  Shipped = 2,
}

// ❌ 错误：禁止使用 any
function badFunc(data: any) { ... }

// ❌ 错误：禁止隐式 any
function badFunc2(data) { ... }
```

### 5.3 API 类型与业务类型分离
```ts
// api/types/user.ts —— 后端契约类型
export interface ApiUserResponse {
  user_id: number
  user_name: string
}

// stores/user.ts —— 业务模型类型
export interface UserInfo {
  id: number
  name: string
}

// utils/adapter.ts —— 适配器转换
export function adaptUser(apiData: ApiUserResponse): UserInfo {
  return {
    id: apiData.user_id,
    name: apiData.user_name,
  }
}
```

---

## 6. 状态管理（Pinia）

### 6.1 Store 结构
```ts
// stores/user.ts
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { getUserInfo } from '@/api/user'
import type { UserInfo } from '@/api/types'

export const useUserStore = defineStore('user', () => {
  // State
  const userInfo = ref<UserInfo | null>(null)
  const loading = ref(false)

  // Getters
  const isLogin = computed(() => !!userInfo.value)
  const displayName = computed(() => userInfo.value?.name || '访客')

  // Actions
  async function fetchUserInfo() {
    loading.value = true
    try {
      const res = await getUserInfo()
      userInfo.value = res.data
    } finally {
      loading.value = false
    }
  }

  function logout() {
    userInfo.value = null
    // 清理本地缓存
    localStorage.removeItem('token')
  }

  return {
    userInfo,
    loading,
    isLogin,
    displayName,
    fetchUserInfo,
    logout,
  }
})
```

### 6.2 使用规范
- **Setup Store 优先**: 使用 Composition API 风格（`defineStore('id', () => {...})`）
- **禁止在组件外直接修改 State**: 必须通过 Action 修改
- **Store 拆分**: 按领域拆分（`user`, `app`, `permission`），禁止万能 Store

---

## 7. 样式规范（Tailwind CSS）

### 7.1 核心原则
- **工具类优先**: 90% 场景使用 Tailwind 工具类，禁止写 `.css` 文件
- **语义化变量**: 颜色/间距使用 Design Token（见 `ui.md`）
- **响应式前缀**: `sm:`, `md:`, `lg:`, `xl:` 控制断点

### 7.2 允许使用 `@apply` 的场景
仅以下情况允许在 `<style>` 中使用 `@apply：
- 第三方组件覆盖
- 复杂动画关键帧
- 全局滚动条样式

```vue
<style scoped>
/* ✅ 允许：第三方组件深度定制 */
:deep(.el-table) {
  @apply border-collapse;
}

/* ❌ 禁止：简单布局也写 CSS */
.container {
  @apply flex flex-col gap-4;  /* 应直接在 template 写 class */
}
</style>
```

### 7.3 暗黑模式
```html
<!-- 使用 dark: 前缀 -->
<div class="bg-white dark:bg-gray-900 text-gray-900 dark:text-white">
  内容
</div>
```

---

## 8. 网络请求规范

### 8.1 Axios 封装
```ts
// utils/request.ts
import axios from 'axios'

const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  timeout: 10000,
})

// 请求拦截：注入 Token
request.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

// 响应拦截：统一错误处理
request.interceptors.response.use(
  (res) => res.data,
  (err) => {
    if (err.response?.status === 401) {
      // 统一跳转登录
      window.location.href = '/login'
    }
    return Promise.reject(err)
  }
)

export default request
```

### 8.2 API 文件组织
```ts
// api/user.ts
import request from '@/utils/request'
import type { ApiUserResponse, ApiLoginParams } from './types'

export function login(params: ApiLoginParams) {
  return request.post<ApiUserResponse>('/api/v1/login', params)
}

export function getUserInfo() {
  return request.get<ApiUserResponse>('/api/v1/user/info')
}
```

---

## 9. 性能优化

- **懒加载**: 路由组件必须使用 `() => import('@/views/xxx.vue')`
- **图片优化**: 使用 `vite-plugin-imagemin` 压缩，WebP 格式优先
- **虚拟滚动**: 列表超过 50 条必须使用虚拟滚动（`vue-virtual-scroller`）
- **防抖节流**: 搜索输入防抖 300ms，按钮点击节流 500ms
- **Watch 优化**: 复杂对象使用 `{ deep: false }` 或精确监听字段

---

## 10. 安全规范

- **XSS 防护**: 禁止直接使用 `v-html`，必须使用 `DOMPurify` 消毒
- **CSRF**: 敏感操作接口必须携带 `X-CSRF-Token`
- **敏感信息**: 禁止在代码中硬编码密钥、密码，使用环境变量
- **路由守卫**: 所有管理后台路由必须校验权限（`beforeEnter`）
