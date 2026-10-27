# codi

**AI 编码代理的质量基础设施**

[English](README.md)

---

## 设计理念：工程的三条边

每个工程问题都存在于三条边界之内：

```
         理论极限（最佳实践）
                    ▲
                   ╱│╲
                  ╱ │ ╲
                 ╱  │  ╲
                ╱   │   ╲
               ╱  阶梯   ╲
              ╱  (CR+AI)  ╲
             ╱      │       ╲
            ╱       │        ╲
           ╱        │         ╲
          ╱─────────┼──────────╲
         能力基线 (make lint/test)
```

### 1. 能力基线（下界）
**工具自动保证的质量底线。**

- `make lint` - golangci-lint、ruff、类型检查
- `make test` - 自动化测试套件
- Git hooks - 提交前验证

**作用：** 在人工审查前挡住低级错误。

### 2. 理论极限（上界）
**基于现有知识能达到的最高水平。**

- Google Go 风格指南 + 生产实践模式
- PEP 8 + 框架最佳实践
- 安全基线、性能模式

**作用：** 知道什么是好代码。

### 3. 攀登的阶梯（路径）
**如何从基线到达极限。**

```
写代码 → CR（发现差距）→ AI 改（应用标准）→ CR（验证）→ 重复
```

**codi 的作用：**
- 提供**极限**（精选标准库）
- 降低**阶梯成本**（AI 自动应用标准）
- 确保**基线**（带工具链的项目模板）

**专业人士与业余爱好者的差别：**
- 业余者不知道极限存在
- 初学者基线很低
- 专业人士知道两条边界，系统性地攀登阶梯

---

## codi 是什么？

AI 辅助开发的统一代码质量系统。

**核心功能：**
- **标准库** - 精选编码标准（Go、Python）
- **项目模板** - 预配置工具链（lint、test、build）
- **AI 集成** - 自动生成 Claude Code、Codex、Cursor 配置
- **代码评审工作流** - AI 自动检查 + 人工评审指南

**为什么重要：**
- AI 写代码速度前所未有
- 质量标准防止 AI 大规模生产烂代码
- codi 让标准对 AI 代理机器可读

---

## 安装

### 从源码安装

```bash
git clone https://github.com/rakuv3r/codi.git
cd codi
make build
sudo cp bin/codi /usr/local/bin/
```

或者不使用 sudo：
```bash
cp bin/codi ~/.local/bin/  # 确保 ~/.local/bin 在你的 PATH 中
```

### 从 Release 安装（即将推出）

```bash
curl -fsSL https://raw.githubusercontent.com/rakuv3r/codi/main/install.sh | bash
```

---

## 快速开始

### 集成现有项目

```bash
cd your-project
codi adopt
```

**做了什么：**
1. 检测语言（Go/Python）
2. 分析依赖
3. 生成 AI 工具配置文件（`.claude/CLAUDE.md`、`.codex/AGENTS.md`）
4. 创建 `.codi/config.yaml` 存储项目元数据
5. 添加缺失的质量工具（Makefile、linters）

**结果：** AI 代理现在自动看到你的标准。

---

### 创建新项目

```bash
codi init my-api
cd my-api
make setup && make build
```

**交互式提示：**
- 语言？（Go、Python）
- 类型？（CLI、API）
- 模块？（Gin、Cobra、FastAPI 等）
- AI 工具？（Claude、Codex、Cursor）

**生成的结构：**
```
my-api/
├── .codi/
│   └── config.yaml          # 项目元数据
├── .claude/
│   ├── CLAUDE.md            # Claude Code 标准
│   ├── review-checklist.md  # AI 自动检查清单
│   └── review-manual.md     # 人工评审指南
├── Makefile                 # lint、test、build、run
├── .golangci.yml            # Linter 配置
└── cmd/main.go              # 脚手架代码
```

---

## 阶梯：代码评审工作流

### 第 1 步：基线检查（自动化）
```bash
make lint  # 格式、类型、常见问题
make test  # 单元测试通过
```

### 第 2 步：AI 评审（review-checklist.md）
**AI 自动完成 70% 的检查：**
- 构建/测试质量
- 错误处理模式
- 安全基础（SQL 注入、XSS）
- 性能反模式

### 第 3 步：人工评审（review-manual.md）
**聚焦 AI 无法判断的：**
- 架构决策
- 业务逻辑正确性
- API 设计
- 可扩展性考虑

### 第 4 步：AI 修复问题
**粘贴 CR 反馈 → AI 读标准 → AI 改代码**

**重复步骤 2-4，直到代码接近极限。**

---

## 标准体系

### 通用标准
- Git 工作流（分支、提交）
- 文档规范（README、API 文档、ADR）
- 安全基线（密钥、依赖）

### Go 标准
- 基于 [Google Go 风格指南](https://google.github.io/styleguide/go/)
- 扩展：HTTP handler 模式、client-go 使用、Neo4j 批处理

### Python 标准
- 基于 PEP 8、[Google Python 风格指南](https://google.github.io/styleguide/pyguide.html)
- 扩展：Django ORM 模式、FastAPI 验证、异步最佳实践

### 依赖特定标准
**根据你的导入自动加载：**
- `github.com/gin-gonic/gin` → Gin 中间件模式
- `github.com/spf13/cobra` → CLI 设计模式
- `sqlalchemy` → ORM 事务管理
- `fastapi` → 依赖注入、验证

**零映射：** 依赖路径 = 标准路径
- `github.com/gin-gonic/gin` → `standards/go/github.com/gin-gonic/gin.md`

---

## CLI 参考

```bash
codi adopt              # 集成到现有项目
codi init <name>        # 创建新项目
codi sync               # 依赖变更后更新标准
codi status             # 显示项目配置
codi view [language]    # 交互式浏览标准
```

---

## 贡献

欢迎贡献！以下是扩展 codi 的方法：

### 添加新语言支持

**1. 创建语言结构：**
```bash
mkdir -p standards/{language}/core
mkdir -p templates/{language}/base/{cli,api}
mkdir -p templates/{language}/module
```

**2. 添加核心标准：**
```bash
# standards/rust/core/
- formatting.md
- error-handling.md
- testing.md
```

**3. 实现分析器：**
```go
// internal/analyzer/rust.go
type RustAnalyzer struct{}

func (a *RustAnalyzer) Analyze(projectPath string) (*DependencyList, error) {
    // 解析 Cargo.toml
    // 返回依赖列表
}
```

**4. 添加模板：**
```bash
# templates/rust/base/cli/
- Cargo.toml.tmpl
- main.rs.tmpl
- Makefile.tmpl
```

**5. 更新检测器：**
```go
// internal/analyzer/detector.go
case "Cargo.toml":
    return LanguageRust, nil
```

---

### 添加库标准

**示例：添加 `axum`（Rust web 框架）**

**1. 创建标准文件：**
```bash
# standards/rust/crates.io/axum.md
```

**2. 编写标准：**
````markdown
# axum

## 路由组织
```rust
// 使用 Router::nest 分组相关路由
let api = Router::new()
    .route("/users", get(list_users).post(create_user))
    .route("/users/:id", get(get_user));

let app = Router::new()
    .nest("/api/v1", api)
    .layer(TraceLayer::new_for_http());
```

## 错误处理
- 使用 `Result<T, AppError>` 处理可失败的 handler
- 为自定义错误类型实现 `IntoResponse`
- 在转换为 HTTP 响应前使用 tracing 记录错误
````

**3. 测试：**
```bash
# 在包含 axum 的 Rust 项目中
codi adopt
# 检查 .claude/CLAUDE.md 是否包含 axum 标准
```

**就这样！** 零映射意味着 codi 自动发现它。

---

### 添加代码评审检查

**编辑 `.claude/review-checklist.md`：**

```markdown
## 安全

### 输入验证
- [ ] 所有用户输入已验证
- [ ] SQL 查询使用参数化语句
- [ ] 文件路径已消毒（无 `../` 遍历）
- [ ] API 实现限流  ← 新增检查
```

**AI 现在会自动检查这一项。**

---

### 项目结构

```
codi/
├── standards/           # "极限" - 最佳实践
│   ├── universal/       # 语言无关
│   ├── go/
│   │   ├── core/        # 语言基础
│   │   └── github.com/  # 库特定（零映射）
│   └── python/
│       ├── core/
│       └── pypi/         # 包标准
│
├── templates/           # 项目脚手架
│   ├── go/
│   │   ├── base/        # cli、api
│   │   └── module/      # gin、cobra、testify
│   └── python/
│       ├── base/
│       └── module/      # fastapi、pytest
│
├── internal/
│   ├── analyzer/        # 检测语言、解析依赖
│   ├── generator/       # 生成项目脚手架
│   ├── standard/        # 加载和提供标准
│   └── command/         # CLI 命令
│
└── cmd/codi/main.go
```

---

### 验证：简单性测试

**能否在 10 分钟内添加新库支持？**

**可以：**
1. 创建 `standards/{lang}/{path}.md`（5 分钟编写）
2. 在测试项目中运行 `codi adopt`（1 分钟）
3. 验证 AI 看到标准（1 分钟）

**如果需要更长时间，说明设计失败。**

---

## 设计原则

### 1. 引用权威，不重复发明
- Google 风格指南
- PEP 8
- 语言特定习语

**仅添加：**
- 指南中没有的实战模式
- 依赖特定的使用方法

### 2. 机器可读的标准
- 标准是给 AI 代理的，不只是人类
- Markdown 格式：人类可读、AI 可解析
- 每个标准都有代码示例

### 3. 零映射
- 依赖路径 = 标准路径
- 无需维护配置文件
- 放入 `.md` 文件，就能工作

### 4. 可组合的质量
```
基线（工具）+ 标准（极限）+ CR（阶梯）= 质量
```

每一层独立但协同。

---

## 路线图

- [ ] Rust 支持
- [ ] TypeScript 支持
- [ ] Java 支持
- [ ] LSP 集成（IDE 中实时标准提示）
- [ ] CI 集成（GitHub Actions 自动 CR）
- [ ] 标准版本化（固定到特定标准版本）

---

## 许可证

MIT License - 见 [LICENSE](LICENSE)

---

## 致谢

**灵感来源：**
- Google 风格指南
- 《程序员修炼之道》
- 吴军《态度》 - 三条边框架

**构建目标：**
- Claude Code
- GitHub Copilot
- OpenAI Codex
- Cursor
- 任何读取项目上下文的 AI 编码助手
