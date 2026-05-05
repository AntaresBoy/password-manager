# 前端测试规范 (FE Testing Spec)

> **版本**: v1.0  
> **测试框架**: Vitest (单元) + Cypress (E2E)  
> **配套工具**: @vue/test-utils, MSW (Mock Service Worker), c8 / istanbul  
> **最后更新**: 2026-05-05

---

## 1. 测试哲学

### 1.1 TDD 铁律（强制）
```
1. RED:  编写一个失败的测试（明确期望行为）
2. GREEN: 编写最小实现代码使测试通过
3. REFACTOR: 重构代码，保持测试通过
```

**禁止**: 先写业务代码再补测试。所有代码提交前，对应测试必须已存在且通过。

### 1.2 测试金字塔
```
      /\
     /  \
    / E2E \      ← 少量（关键路径）
   /─────────\
  / Component \   ← 中量（组件交互）
 /─────────────\
/    Unit       \ ← 大量（业务逻辑）
───────────────────
```

- **单元测试**: 70% —— 工具函数、Composables、Store、纯逻辑
- **组件测试**: 20% —— 组件渲染、Props、Events、Slots、用户交互
- **E2E 测试**: 10% —— 完整用户流程、跨页面导航、真实网络

---

## 2. 单元测试规范（Vitest）

### 2.1 文件组织
```
src/
├── utils/
│   ├── validator.ts
│   └── validator.test.ts          # 同目录，*.test.ts
├── composables/
│   ├── useAuth.ts
│   └── useAuth.test.ts
├── stores/
│   ├── user.ts
│   └── __tests__/user.test.ts     # 或 __tests__ 子目录
└── components/__ui__/
    ├── Button.vue
    └── Button.spec.ts             # *.spec.ts 亦可
```

### 2.2 测试文件模板
```ts
// utils/validator.test.ts
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { validatePhone, validateEmail } from './validator'

describe('validator', () => {
  describe('validatePhone', () => {
    // 表驱动测试（推荐）
    const cases = [
      { input: '13800138000', expected: true, desc: '标准手机号' },
      { input: '1380013800',  expected: false, desc: '少一位' },
      { input: '23800138000', expected: false, desc: '非 1 开头' },
      { input: '',            expected: false, desc: '空字符串' },
      { input: null,          expected: false, desc: 'null' },
    ]

    cases.forEach(({ input, expected, desc }) => {
      it(`should return ${expected} for ${desc}`, () => {
        expect(validatePhone(input)).toBe(expected)
      })
    })
  })

  describe('validateEmail', () => {
    it('should return true for valid email', () => {
      expect(validateEmail('user@example.com')).toBe(true)
    })

    it('should return false for missing @', () => {
      expect(validateEmail('userexample.com')).toBe(false)
    })
  })
})
```

### 2.3 Composables 测试
```ts
// composables/useAuth.test.ts
import { describe, it, expect, vi } from 'vitest'
import { ref } from 'vue'
import { useAuth } from './useAuth'

// Mock 依赖
vi.mock('@/api/user', () => ({
  login: vi.fn(),
}))

describe('useAuth', () => {
  it('should set user after successful login', async () => {
    const { user, login } = useAuth()

    expect(user.value).toBeNull()

    await login({ username: 'test', password: '123456' })

    expect(user.value).toEqual({ id: 1, name: 'test' })
  })

  it('should throw error when login fails', async () => {
    const { login } = useAuth()

    await expect(
      login({ username: 'bad', password: 'wrong' })
    ).rejects.toThrow('用户名或密码错误')
  })
})
```

### 2.4 Store 测试（Pinia）
```ts
// stores/user.test.ts
import { setActivePinia, createPinia } from 'pinia'
import { useUserStore } from './user'
import { describe, it, expect, beforeEach, vi } from 'vitest'

vi.mock('@/api/user', () => ({
  getUserInfo: vi.fn(() => Promise.resolve({ data: { id: 1, name: 'test' } })),
}))

describe('User Store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('should initialize with null user', () => {
    const store = useUserStore()
    expect(store.userInfo).toBeNull()
    expect(store.isLogin).toBe(false)
  })

  it('should fetch and set user info', async () => {
    const store = useUserStore()
    await store.fetchUserInfo()

    expect(store.userInfo).toEqual({ id: 1, name: 'test' })
    expect(store.isLogin).toBe(true)
  })
})
```

---

## 3. 组件测试规范（@vue/test-utils）

### 3.1 测试文件模板
```ts
// components/__ui__/Button.test.ts
import { describe, it, expect, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import Button from './Button.vue'

describe('Button', () => {
  // Props 测试
  it('should render with correct text', () => {
    const wrapper = mount(Button, {
      props: { label: '提交' },
    })
    expect(wrapper.text()).toContain('提交')
  })

  it('should apply primary class by default', () => {
    const wrapper = mount(Button)
    expect(wrapper.classes()).toContain('bg-primary-500')
  })

  it('should apply danger class when type is danger', () => {
    const wrapper = mount(Button, {
      props: { type: 'danger' },
    })
    expect(wrapper.classes()).toContain('bg-error-500')
  })

  // Events 测试
  it('should emit click event when clicked', async () => {
    const wrapper = mount(Button)
    await wrapper.find('button').trigger('click')
    expect(wrapper.emitted()).toHaveProperty('click')
    expect(wrapper.emitted('click')).toHaveLength(1)
  })

  it('should not emit click when disabled', async () => {
    const wrapper = mount(Button, {
      props: { disabled: true },
    })
    await wrapper.find('button').trigger('click')
    expect(wrapper.emitted('click')).toBeUndefined()
  })

  // Slots 测试
  it('should render default slot content', () => {
    const wrapper = mount(Button, {
      slots: { default: '自定义内容' },
    })
    expect(wrapper.text()).toContain('自定义内容')
  })

  // Loading 状态
  it('should show spinner and disable button when loading', () => {
    const wrapper = mount(Button, {
      props: { loading: true },
    })
    expect(wrapper.find('[data-testid="spinner"]').exists()).toBe(true)
    expect(wrapper.find('button').attributes('disabled')).toBeDefined()
  })
})
```

### 3.2 复杂组件测试（表单）
```ts
// components/LoginForm.test.ts
import { mount } from '@vue/test-utils'
import LoginForm from './LoginForm.vue'
import { describe, it, expect, vi } from 'vitest'

describe('LoginForm', () => {
  it('should validate fields before submit', async () => {
    const wrapper = mount(LoginForm)

    // 不填任何内容直接提交
    await wrapper.find('form').trigger('submit')

    // 应该显示错误提示
    expect(wrapper.find('[data-testid="username-error"]').text()).toBe('用户名不能为空')
    expect(wrapper.find('[data-testid="password-error"]').text()).toBe('密码不能为空')
  })

  it('should call login API with correct params', async () => {
    const mockLogin = vi.fn(() => Promise.resolve({ token: 'abc' }))
    const wrapper = mount(LoginForm, {
      global: {
        provide: { loginService: mockLogin },
      },
    })

    await wrapper.find('input[name="username"]').setValue('admin')
    await wrapper.find('input[name="password"]').setValue('123456')
    await wrapper.find('form').trigger('submit')

    expect(mockLogin).toHaveBeenCalledWith({ username: 'admin', password: '123456' })
  })
})
```

---

## 4. E2E 测试规范（Cypress）

### 4.1 文件组织
```
cypress/
├── e2e/
│   ├── auth/
│   │   └── login.cy.ts          # 按模块组织
│   ├── order/
│   │   └── create-order.cy.ts
│   └── dashboard/
│       └── overview.cy.ts
├── fixtures/                    # 测试数据
│   └── users.json
├── support/
│   ├── commands.ts              # 自定义命令
│   └── e2e.ts                   # 全局配置
└── tsconfig.json
```

### 4.2 E2E 测试模板
```ts
// cypress/e2e/auth/login.cy.ts
describe('登录流程', () => {
  beforeEach(() => {
    // 拦截 API，控制响应
    cy.intercept('POST', '/api/v1/login', {
      statusCode: 200,
      body: { code: 0, data: { token: 'fake-jwt-token' } },
    }).as('loginRequest')

    cy.visit('/login')
  })

  it(' should login successfully with valid credentials', () => {
    // 填写表单
    cy.get('[data-testid="username-input"]').type('admin')
    cy.get('[data-testid="password-input"]').type('correct-password')

    // 提交
    cy.get('[data-testid="login-button"]').click()

    // 等待接口响应
    cy.wait('@loginRequest')

    // 验证跳转
    cy.url().should('include', '/dashboard')

    // 验证本地存储
    cy.window().its('localStorage.token').should('eq', 'fake-jwt-token')

    // 验证欢迎消息
    cy.contains('欢迎回来，admin').should('be.visible')
  })

  it('should show error message with invalid credentials', () => {
    cy.intercept('POST', '/api/v1/login', {
      statusCode: 200,
      body: { code: 20002, message: '密码错误' },
    }).as('loginFail')

    cy.get('[data-testid="username-input"]').type('admin')
    cy.get('[data-testid="password-input"]').type('wrong-password')
    cy.get('[data-testid="login-button"]').click()

    cy.wait('@loginFail')
    cy.get('[data-testid="error-message"]').should('contain', '密码错误')
    cy.url().should('include', '/login')  // 未跳转
  })

  it('should validate empty fields', () => {
    cy.get('[data-testid="login-button"]').click()

    cy.get('[data-testid="username-error"]').should('contain', '不能为空')
    cy.get('[data-testid="password-error"]').should('contain', '不能为空')
  })
})
```

### 4.3 关键路径定义（必须覆盖）

| 模块 | 黄金路径 | 异常路径 |
|------|---------|---------|
| 认证 | 注册 → 登录 → 登出 | 密码错误、Token 过期、重复注册 |
| 订单 | 选商品 → 填地址 → 支付 → 查看订单 | 库存不足、支付失败、取消订单 |
| 用户 | 修改资料 → 上传头像 → 保存 | 图片过大、格式错误、网络中断 |
| 权限 | 管理员登录 → 进入后台 → 分配角色 | 无权限访问、越权操作 |

### 4.4 自定义命令
```ts
// cypress/support/commands.ts
Cypress.Commands.add('login', (username: string, password: string) => {
  cy.session([username, password], () => {
    cy.request('POST', '/api/v1/login', { username, password }).then((res) => {
      window.localStorage.setItem('token', res.body.data.token)
    })
  })
})

Cypress.Commands.add('dataTestId', (value: string) => {
  return cy.get(`[data-testid="${value}"]`)
})

// 使用
cy.login('admin', '123456')
cy.visit('/dashboard')
cy.dataTestId('user-menu').click()
```

---

## 5. Mock 规范（MSW）

### 5.1 使用场景
- 单元/组件测试中的 API 调用
- 开发环境未就绪接口
- E2E 中需要控制边界响应

### 5.2 Mock 文件组织
```
src/
└── mocks/
    ├── browser.ts           # 开发环境启用
    ├── server.ts            # 测试环境启用
    ├── handlers.ts          # 处理器列表
    └── data/                # 模拟数据
        ├── users.ts
        └── orders.ts
```

### 5.3 Handler 示例
```ts
// mocks/handlers.ts
import { http, HttpResponse } from 'msw'

export const handlers = [
  http.get('/api/v1/users', () => {
    return HttpResponse.json({
      code: 0,
      data: {
        list: [
          { id: 1, name: '张三', status: 1 },
          { id: 2, name: '李四', status: 2 },
        ],
        total: 2,
      },
    })
  }),

  http.post('/api/v1/login', async ({ request }) => {
    const body = await request.json() as { username: string; password: string }

    if (body.password === 'wrong') {
      return HttpResponse.json(
        { code: 20002, message: '密码错误' },
        { status: 401 }
      )
    }

    return HttpResponse.json({
      code: 0,
      data: { token: 'mock-token-12345' },
    })
  }),
]
```

---

## 6. 覆盖率规范

### 6.1 阈值配置（vitest.config.ts）
```ts
import { defineConfig } from 'vitest/config'

export default defineConfig({
  test: {
    coverage: {
      provider: 'v8',        // 或 'istanbul'
      reporter: ['text', 'json', 'html'],
      reportsDirectory: './coverage',
      thresholds: {
        statements: 80,
        branches: 70,
        functions: 80,
        lines: 80,
      },
      // 排除文件
      exclude: [
        'node_modules/',
        'src/**/*.d.ts',
        'src/main.ts',
        'src/router/**',
        'src/api/types/**',
        '**/*.config.*',
        '**/mock/**',
      ],
    },
  },
})
```

### 6.2 覆盖率解读与提升策略
| 指标 | 含义 | 提升方法 |
|------|------|---------|
| Statements | 语句执行比例 | 补充未覆盖分支的用例 |
| Branches | 条件分支覆盖 | 每个 if/else/switch 至少两个用例 |
| Functions | 函数调用比例 | 导出函数全部测试，私有函数通过公有接口间接覆盖 |
| Lines | 代码行覆盖 | 删除死代码，补充边界值测试 |

### 6.3 覆盖率报告输出
每次 CI 必须生成：
- `coverage/index.html` —— 本地可视化报告
- `coverage/coverage-summary.json` —— 自动化解析
- `openspec/reports/test-report-{timestamp}.html` —— OpenSpec 归档

---

## 7. 测试数据工厂（Factory）

### 7.1 使用 faker / chance 生成假数据
```ts
// tests/factories/user.ts
import { faker } from '@faker-js/faker/locale/zh_CN'
import type { UserInfo } from '@/stores/user'

export function createUser(overrides?: Partial<UserInfo>): UserInfo {
  return {
    id: faker.number.int({ min: 1, max: 10000 }),
    name: faker.person.fullName(),
    email: faker.internet.email(),
    phone: faker.phone.number('138########'),
    avatar: faker.image.avatar(),
    status: faker.helpers.arrayElement([1, 2]),
    ...overrides,
  }
}

// 使用
const mockUser = createUser({ name: '特定测试名', status: 2 })
```

### 7.2 测试数据隔离
- 每个测试独立数据，禁止测试间共享可变状态
- 数据库测试使用 `beforeEach` 清理 + `afterEach` 回滚
- 本地存储使用 `cy.clearLocalStorage()` 或 mock

---

## 8. CI/CD 集成

### 8.1 GitLab CI 阶段
```yaml
stages:
  - lint
  - test:unit
  - test:e2e
  - build

test:unit:
  stage: test:unit
  script:
    - npm ci
    - npm run test:unit -- --coverage
  coverage: '/All files[^|]*\|[^|]*\s+([\d\.]+)/'
  artifacts:
    reports:
      coverage_report:
        coverage_format: cobertura
        path: coverage/cobertura-coverage.xml
    paths:
      - coverage/

test:e2e:
  stage: test:e2e
  image: cypress/browsers:node-20.14.0-chrome-125.0.6422.141-1-ff-126.0.1-edge-125.0.2535.85-1
  script:
    - npm ci
    - npm run test:e2e:ci
  artifacts:
    paths:
      - cypress/screenshots/
      - cypress/videos/
    when: always
```

### 8.2 质量门禁
- `test:unit` 阶段覆盖率未达标 → Pipeline 失败
- `test:e2e` 有失败用例 → Pipeline 失败
- 测试报告自动上传到 GitLab Pages / S3

---

## 9. 测试编写 checklist

提交代码前自检：
- [ ] 每个新增/修改的函数都有对应单元测试
- [ ] 每个条件分支（if/else/switch/catch）至少一个测试用例
- [ ] 组件 Props、Events、Slots 已测试
- [ ] E2E 覆盖了至少一条黄金路径
- [ ] Mock 数据与真实数据结构一致
- [ ] 测试名称清晰描述「在什么情况下，期望什么结果」
- [ ] 无 `console.log`、无 `.only`、无 `.skip` 残留
- [ ] 覆盖率四项指标全部达标
- [ ] 测试执行时间合理（单文件 < 5s）
