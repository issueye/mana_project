# Mana 开发计划（Go + MyGo + SQLite）

| 项 | 内容 |
|---|---|
| 文档版本 | v1.0 |
| 日期 | 2026-11-18 |
| 需求基线 | `docs/requirements.md` v1.0 |
| 设计基线 | `design/README.md`、`design/mana-prototype.html` |
| 技术栈 | Go 1.27.3+ · MyGo v0.9.0（原生 UI） · `modernc.org/sqlite`（纯 Go，无 cgo） |
| 交付形态 | 单机桌面应用，Windows / macOS / Linux，无 cgo、无前端构建 |
| 预计工期 | 10.5 人日 + 2 人日缓冲（约 2.5 周，单人） |

> **实施记录（2026-11-18）**：M0–M5 的主要 P0 已落地，代码在 `internal/app`、`internal/store`、`internal/service`；`go test ./...`、`go vet ./...`、`CGO_ENABLED=0 go test ./...` 通过，`go tool mygo build` 已产出 `build/windows-amd64/Mana.exe`。SQLite 驱动由原计划的 MyGo 原生插件改为 `modernc.org/sqlite`（本机无法下载 GitHub 原生库），SQL/schema/事务不变。P1 中的全局文档搜索、设置/备份、行级“移入组”菜单尚未收尾。

---

## 1. 目标与交付范围

### 1.1 本轮目标

按需求文档实现 v1.0 全部 P0 功能，并完成 P1 中的项目移动、打开目录、复制路径、全局搜索、数据设置：

1. 项目 CRUD（名称、标签、备注、路径、版本、所属组）。
2. 项目组 CRUD（名称、描述、标识色），组内项目与组文档管理。
3. 项目文档与组文档 CRUD，Markdown 编辑 + 实时预览。
4. 首页统计与最近更新、最近文档。
5. 全局搜索、快捷键、空状态、错误反馈、删除二次确认。
6. SQLite 本地持久化、迁移、备份入口。
7. 可打包的三平台桌面应用与自动化测试。

### 1.2 不在本轮

- 多人协作、账号、权限、云同步。
- Git 集成、CI/CD、看板/甘特图。
- 附件上传、Office/PDF 预览、图片粘贴。
- 自动扫描磁盘、目录监听。
- 深色主题（保留 Token 结构，v1.1 再做）。

### 1.3 需求追踪

| 需求 | 主要模块 | 验证 |
|---|---|---|
| FR-01 框架与导航 | `internal/ui/shell.go`、`internal/app` | UI 测试 + 手动窗口检查 |
| FR-02 项目管理 | `internal/ui/projects.go`、`internal/service`、`internal/store` | Store/Service/UI 测试，AC-01 |
| FR-03 项目组管理 | `internal/ui/groups.go`、`internal/store/groups.go` | AC-02，级联事务测试 |
| FR-04 文档管理 | `internal/ui/documents.go`、`internal/ui/editor.go` | AC-03，编辑器测试 |
| FR-05 首页 | `internal/ui/home.go`、`internal/store/stats.go` | AC-04，统计一致性测试 |
| FR-06 全局搜索 | `internal/ui/search.go`、`internal/store/search.go` | 搜索用例 + UI 测试 |
| FR-07 设置与数据 | `internal/ui/settings.go`、`internal/platform` | 手动/集成测试 |
| NFR 性能/可靠/可访问 | 全局 | 基准、`ui.NewTester`、手动 |

---

## 2. 技术栈与版本

| 层 | 选型 | 版本/说明 |
|---|---|---|
| 语言 | Go | `go 1.27.3`（MyGo v0.9.0 的 `go.mod` 要求；本机 1.27.1 会由 toolchain 自动升级） |
| UI | MyGo 原生 UI | `github.com/egoist/mygo@v0.9.0`，`ui` 包；不启 webview、无 JS/CSS |
| 窗口/系统 | MyGo | `mygo.NewWindow`、`mygo.Dialog`、`mygo.Shell` |
| 数据库 | `database/sql` + `modernc.org/sqlite` | v1.62.0；纯 Go SQLite 驱动，无 cgo、无需下载原生模块。原计划的 MyGo SQLite 插件因本机无法拉取 GitHub 原生库（超时）改用此实现，SQL 与 schema 不变 |
| Markdown | MyGo | `github.com/egoist/mygo/ui/markdown` |
| 测试 | Go testing + MyGo Tester | `ui.NewTester` 无头渲染，`CGO_ENABLED=0` 可跑 |
| 工具链 | MyGo CLI | `go tool mygo dev/build`，`mygo.json`，`resources/icon.png` |

初始化方式（新建仓库时）：

```sh
go run github.com/egoist/mygo/cmd/mygo@v0.9.0 init -template native mana
cd mana
go get github.com/egoist/mygo@v0.9.0
go get -tool github.com/egoist/mygo/cmd/mygo@v0.9.0
go mod tidy
go tool mygo doctor
```

> 约束：所有 UI 图标使用 `ui.MustParseSVG` 解析设计稿中的 24 网格 SVG；不引入 emoji 或字符图标。字体使用 Inter / Noto Sans SC（Commercial 可用），系统字体只做回退。

---

## 3. 总体架构

```
main.go
        │
        ▼
internal/app        组合根：窗口、路由、异步加载、错误呈现
        │
        ├── internal/ui          视图与交互（只读状态，不直接碰 SQL）
        │        │
        │        ▼
        ├── internal/service     校验、业务规则、事务编排
        │        │
        │        ▼
        ├── internal/store       SQLite 连接、迁移、仓储查询
        │        │
        │        ▼
        └── internal/model       纯数据结构与枚举
```

依赖规则：

- `ui` 不得导入 `store`；只依赖 `service` 的输入/输出类型。
- `store` 不得导入 `ui`；可在测试中独立使用。
- `service` 持有 `*store.Store`，不持有 `*ui.Context`、`*mygo.Window`。
- `app` 是唯一知道窗口与数据库生命周期的地方。
- 所有 SQL 只出现在 `internal/store`，页面里不拼 SQL。

---

## 4. 目录结构

```
mana/
├── main.go                         # 入口：App.WhenReady、建窗、Run
├── go.mod
├── go.sum
├── mygo.json                       # 应用名称、identifier、版本
├── resources/
│   └── icon.png                    # 1024×1024
├── internal/
│   ├── app/
│   │   ├── app.go                  # App 结构、Start/AttachWindow
│   │   ├── view.go                 # 根视图与布局
│   │   ├── load.go                 # 异步加载、generation 防陈旧
│   │   └── shortcuts.go            # Ctrl+N / Ctrl+Shift+N / Ctrl+F
│   ├── model/
│   │   ├── group.go
│   │   ├── project.go
│   │   ├── document.go
│   │   └── stats.go
│   ├── store/
│   │   ├── store.go                # Open、Close、migrate、helpers
│   │   ├── groups.go
│   │   ├── projects.go
│   │   ├── documents.go
│   │   ├── search.go
│   │   ├── stats.go
│   │   └── migrations/
│   │       └── 001_init.sql
│   ├── service/
│   │   ├── service.go
│   │   ├── validation.go
│   │   ├── errors.go               # FieldError / ErrNotFound / ErrConflict
│   │   └── *_test.go
│   ├── ui/
│   │   ├── shell.go                # 侧栏 + 顶栏 + 内容容器
│   │   ├── theme.go                # Ant 5 Token → MyGo Theme
│   │   ├── icons.go                # SVG 缓存
│   │   ├── home.go
│   │   ├── groups.go
│   │   ├── projects.go             # 项目表格、详情
│   │   ├── documents.go            # 文档表格
│   │   ├── editor.go               # Markdown 编辑 + 预览
│   │   ├── dialogs.go              # 新建/编辑项目、项目组、删除确认
│   │   ├── search.go
│   │   ├── widgets.go              # emptyState、tag、statCard、card
│   │   └── *_test.go
│   └── platform/
│       ├── folder.go               # mygo.Dialog.Open(Directory)
│       └── reveal.go               # mygo.Shell.ShowItemInFolder
└── docs/
    ├── requirements.md
    └── development-plan.md
```

---

## 5. 数据层设计

### 5.1 数据库位置与连接

```go
dir, err := mygo.App.Path(mygo.PathUserData)
if err != nil { return fmt.Errorf("user data dir: %w", err) }
path := filepath.Join(dir, "mana.db")

db, err := sql.Open("sqlite", path)
if err != nil { return fmt.Errorf("open db: %w", err) }
db.SetMaxOpenConns(1)
```

- 单个 `*sql.DB` 连接（`SetMaxOpenConns(1)`），避免 SQLite 多连接写锁竞争。
- 启动时 `ExecContext` 执行 `PRAGMA foreign_keys = ON`、`PRAGMA busy_timeout = 5000`、`PRAGMA journal_mode = WAL`（文件库）。
- 外键由 pragma 开启；迁移脚本不写 PRAGMA，避免在事务中执行 `foreign_keys` / `journal_mode`。
- `Close` 在 `App.OnQuit` 或窗口关闭时调用；连接生命周期归 `app`。
- 测试用 `sql.Open("sqlite", ":memory:")`，每个测试独立连接。

### 5.2 迁移

```go
//go:embed migrations/*.sql
var migrationFS embed.FS
```

启动流程：

1. `CREATE TABLE IF NOT EXISTS schema_migrations(version INTEGER PRIMARY KEY, applied_at INTEGER NOT NULL);`
2. 读取 `migrations/*.sql`，按文件名升序。
3. 对每个未应用版本：`splitStatements(sql)`（去掉 `--` 注释、按 `;` 切分；本项目 SQL 不含字符串内分号）→ `db.BeginTx` → 逐条 `tx.ExecContext` → 写入版本号 → `tx.Commit`。
4. 任何一步失败整体回滚，应用显示错误并退出，避免半迁移。

`001_init.sql` 以需求文档 8.2 为准：

- `project_groups`：`ux_project_groups_name`。
- `projects`：`ux_projects_group_name(IFNULL(group_id,0), name COLLATE NOCASE)`、`ix_projects_group`、`ix_projects_updated`。
- `documents`：`ux_documents_owner_title(owner_type, owner_id, title COLLATE NOCASE)`、`ix_documents_owner`。
- 时间统一 `INTEGER`（Unix 秒，UTC）；Go 侧用 `time.Time`。

> 迁移只增不改：后续变更新增 `002_*.sql`，不得修改已发布脚本。

### 5.3 仓储接口

```go
type Store struct {
    db  *sql.DB
    now func() time.Time
}

// groups
func (s *Store) ListGroups(ctx context.Context) ([]model.Group, error)
func (s *Store) GetGroup(ctx context.Context, id int64) (model.Group, error)
func (s *Store) CreateGroup(ctx context.Context, in model.GroupInput) (model.Group, error)
func (s *Store) UpdateGroup(ctx context.Context, id int64, in model.GroupInput) (model.Group, error)
func (s *Store) DeleteGroup(ctx context.Context, id int64) error

// projects
func (s *Store) ListProjects(ctx context.Context, f model.ProjectFilter) ([]model.Project, error)
func (s *Store) GetProject(ctx context.Context, id int64) (model.Project, error)
func (s *Store) CreateProject(ctx context.Context, in model.ProjectInput) (model.Project, error)
func (s *Store) UpdateProject(ctx context.Context, id int64, in model.ProjectInput) (model.Project, error)
func (s *Store) DeleteProject(ctx context.Context, id int64) error
func (s *Store) MoveProject(ctx context.Context, id, groupID int64) error

// documents
func (s *Store) ListDocuments(ctx context.Context, ownerType string, ownerID int64) ([]model.Document, error)
func (s *Store) GetDocument(ctx context.Context, id int64) (model.Document, error)
func (s *Store) CreateDocument(ctx context.Context, in model.DocumentInput) (model.Document, error)
func (s *Store) UpdateDocument(ctx context.Context, id int64, in model.DocumentInput) (model.Document, error)
func (s *Store) DeleteDocument(ctx context.Context, id int64) error

// stats & search
func (s *Store) Dashboard(ctx context.Context) (model.Dashboard, error)
func (s *Store) Search(ctx context.Context, q string, limit int) (model.SearchResult, error)
```

辅助函数：

```go
func queryOne[T any](ctx context.Context, db *sql.DB, sqlText string, args []any, scan func(*sql.Row) (T, error)) (T, error)
func queryMany[T any](ctx context.Context, db *sql.DB, sqlText string, args []any, scan func(*sql.Rows) (T, error)) ([]T, error)
```

`sql.Rows` 与 `sql.Row` 按列名/顺序扫描；把 `NULL` 的 `group_id` 读为 `sql.NullInt64`，避免依赖列顺序的脆弱代码。

UI 用 `-1` 表示「未分组」、`0` 表示「全部项目」；在 service 边界转换：`groupID <= 0` 写入 `NULL`，`groupID > 0` 校验组存在。避免把 0 当成真实组 ID。

### 5.4 关键 SQL 与事务

**组列表 + 计数（一次查询，避免 N+1）**

```sql
SELECT g.id, g.name, g.description, g.color, g.created_at, g.updated_at,
       (SELECT COUNT(*) FROM projects p WHERE p.group_id = g.id) AS project_count,
       (SELECT COUNT(*) FROM documents d
          WHERE d.owner_type='group' AND d.owner_id = g.id) AS document_count
FROM project_groups g
ORDER BY g.name COLLATE NOCASE;
```

**删除项目组（事务）**

```go
tx, err := s.db.BeginTx(ctx, nil)
if err != nil { return err }
defer tx.Rollback()
for _, statement := range []string{
    `UPDATE projects SET group_id = NULL WHERE group_id = ?`,
    `DELETE FROM documents WHERE owner_type='group' AND owner_id = ?`,
    `DELETE FROM project_groups WHERE id = ?`,
} {
    if _, err := tx.ExecContext(ctx, statement, id); err != nil { return err }
}
return tx.Commit()
```

顺序固定：先迁项目，再删组文档，最后删组。任何一步失败整体回滚。

**删除项目（事务）**

```go
tx, err := s.db.BeginTx(ctx, nil)
if err != nil { return err }
defer tx.Rollback()
for _, statement := range []string{
    `DELETE FROM documents WHERE owner_type='project' AND owner_id = ?`,
    `DELETE FROM projects WHERE id = ?`,
} {
    if _, err := tx.ExecContext(ctx, statement, id); err != nil { return err }
}
return tx.Commit()
```

**标签 JSON**

- 存储：`encoding/json` 序列化为 `["Go","MyGo"]`。
- 读取：`json.Unmarshal`；损坏数据回退为空切片并记录日志，不阻断列表。
- 筛选标签：v1.0 在 Go 侧过滤，或 `LIKE '%"tag"%'`；数据量 < 5k 时 Go 侧更简单、可控。

**唯一性冲突**

```go
func mapError(op string, err error) error {
    if strings.Contains(err.Error(), "UNIQUE constraint failed") {
        return fmt.Errorf("%w: %s", ErrConflict, err)
    }
    return fmt.Errorf("%s: %w", op, err)
}
```

同时在 service 层先查一次，返回带字段的错误，给用户可读提示；数据库唯一索引是最后防线。

### 5.5 并发与错误

- 所有写操作串行在单连接上；跨表写使用 `Transaction`。
- `context.WithTimeout(ctx, 3*time.Second)` 包裹页面级查询；超时返回可读错误。
- SQL 层错误统一包一层 `fmt.Errorf("store: ...: %w")`；service 转成 `FieldError` 或 `ErrNotFound`/`ErrConflict`；UI 只消费后者。
- 不在 UI 线程做文件 IO、系统对话框或大查询；放到 goroutine，结果回到 `win.Update`。

---

## 6. 业务服务层

### 6.1 输入模型

```go
type ProjectInput struct {
    Name        string
    GroupID     *int64
    Tags        []string
    Description string
    Path        string
    Version     string
}

type DocumentInput struct {
    OwnerType string // "project" | "group"
    OwnerID   int64
    Title     string
    Kind      string
    Content   string
    Link      string
}
```

### 6.2 校验规则（与需求 BR-01..BR-07 对齐）

| 字段 | 规则 | 错误提示 |
|---|---|---|
| 项目名称 | `TrimSpace` 后 1–80 rune；同组唯一（忽略大小写） | 「请输入项目名称」/「该组下已存在同名项目」 |
| 标签 | 0–5 个，每个 1–16 rune；去重、去首尾空格 | 「最多添加 5 个标签」/「标签不能超过 16 个字符」 |
| 备注 | 0–1000 rune | 「备注最多 1000 个字符」 |
| 路径 | 0–512；允许不存在；若存在必须是绝对路径 | 「路径不存在，仍可保存登记」为提示，不阻断 |
| 版本 | 0–32 | 「版本最多 32 个字符」 |
| 组名 | 1–60 rune；全局唯一 | 「该名称已被其他项目组使用」 |
| 组描述 | 0–300 | 同上 |
| 文档标题 | 1–120；同 owner 唯一 | 「该归属下已存在同名文档」 |
| 文档类型 | 六类枚举 | 「请选择文档类型」 |
| 文档正文 | 0–200,000 rune | 「文档内容过长，请拆分」 |
| 关联链接 | 0–512 | 同上 |

字符计数统一用 `utf8.RuneCountInString(strings.TrimSpace(v))`，不要用 `len()`（中文会算成 3）。

### 6.3 服务接口

```go
type Service struct{ store *store.Store }

func (s *Service) Dashboard(ctx) (model.Dashboard, error)
func (s *Service) Groups(ctx) ([]model.Group, error)
func (s *Service) CreateGroup(ctx, in model.GroupInput) (model.Group, error)
func (s *Service) UpdateGroup(ctx, id, in model.GroupInput) (model.Group, error)
func (s *Service) DeleteGroup(ctx, id) error
func (s *Service) Projects(ctx, f model.ProjectFilter) ([]model.Project, error)
func (s *Service) Project(ctx, id) (model.Project, error)
func (s *Service) CreateProject(ctx, in model.ProjectInput) (model.Project, error)
func (s *Service) UpdateProject(ctx, id, in model.ProjectInput) (model.Project, error)
func (s *Service) DeleteProject(ctx, id) error
func (s *Service) MoveProject(ctx, id, groupID) error
func (s *Service) Documents(ctx, ownerType, ownerID) ([]model.Document, error)
func (s *Service) SaveDocument(ctx, id int64, in model.DocumentInput) (model.Document, error) // id=0 新建
func (s *Service) DeleteDocument(ctx, id) error
func (s *Service) Search(ctx, q, limit) (model.SearchResult, error)
```

### 6.4 删除语义

- 删除项目组：服务层先取组与计数，生成确认文案（项目数、组文档数），再调用 `store.DeleteGroup`。
- 删除项目：先取文档数，确认后 `store.DeleteProject`；项目文档随项目删除。
- 删除文档：确认后直接删除。
- 所有删除确认由 UI 触发；service 不做「未确认」判断（确认是交互层职责）。

---

## 7. UI 架构与 MyGo 落地

### 7.1 应用状态

```go
type Page int

const (
    PageHome Page = iota
    PageGroups
    PageProject
    PageEditor
)

type App struct {
    svc  *service.Service
    win  *mygo.Window
    page Page

    // 导航
    selectedGroupID   int64 // 0 = 全部项目，-1 = 未分组
    selectedProjectID int64
    selectedDocID     int64
    editorReturn      Page

    // 数据
    groups    []model.Group
    projects  []model.Project
    documents []model.Document
    dashboard model.Dashboard
    loading   bool
    loadErr   error

    // 控件状态（必须持久，不能放局部变量）
    groupList   ui.ListState
    projectList ui.ListState
    docList     ui.ListState
    tab         int

    // 弹窗与表单
    projectDialogOpen bool
    groupDialogOpen   bool
    deleteDialogOpen  bool
    deleteKind        string
    projectForm       model.ProjectInput
    groupForm         model.GroupInput
    fieldErrs         map[string]string
    saving            bool
}
```

规则：

- 元素与 `*Base` 返回值在构建帧结束后失效；只把 `ListState`、`bool`、字符串、ID 存进 `App`。
- 表单绑定到 `app.projectForm` 的字段；不要在 view 内重新计算并丢弃输入。
- `ui.Local` 只用于单页瞬时状态（如 hover 提示），跨页状态放 `App`。
- `*ui.Context` 不进入 goroutine、不存字段。

### 7.2 根布局与导航

```go
func (a *App) View(c *ui.Context) {
    ui.Column(c).Fill().Children(func() {
        a.windowBar(c) // 36 DIP，DragWindow + 自绘窗口控制
        ui.Row(c).Grow(1).FillWidth().Children(func() {
            a.sidebar(c) // 228 DIP
            ui.Column(c).Grow(1).FillHeight().Children(func() {
                a.topbar(c) // 62 DIP
                a.pageContent(c)
            })
        })
    })
}
```

- 无边框：`TitleBarStyle: mygo.TitleBarHidden` + `HideWindowButtons: true`，保留窗口缩放边缘；自绘 36 DIP 标题栏、拖拽区与最小化/最大化/关闭按钮。
- 侧栏用 `ui.Column` + 自绘导航项（`ui.Button` 无脸态或 `SidebarBase`），当前项浅蓝底 + 左侧 3 DIP 主色条。
- 顶部栏：标题/面包屑 + `ui.SearchField`（宽 340）+ 通知/主操作/头像。
- 主操作「新建项目」唯一主色按钮；其余为描边/文字按钮。
- 路由用 `Page` 枚举 + 选中 ID；返回时恢复来源页（编辑器从详情进入则回详情，否则回组页）。v1.0 不引入 URL/Router，避免与表单状态耦合。

### 7.3 页面落地

#### 首页

| 设计元素 | MyGo 实现 |
|---|---|
| 问候 + 日期 | `ui.Text`（Display 25/680） |
| 统计卡 ×4 | `ui.Row` + 四个 `statCard`（Box + 图标底 + 数值） |
| 最近更新项目 | `ui.Table`，6 行，列：项目/版本/标签/更新时间；行点击进入 `PageProject` |
| 项目组概览 | 自绘 groupRow（色点 + 名称 + 计数 + 箭头），点击进入组 |
| 最近文档 | 2×2 网格 `ui.Grid` 或两行 `Row`；点击进入编辑器 |
| 主操作 | 顶栏 `ui.PrimaryButton("新建项目")` |

#### 项目组管理

```go
// 左右两栏 + 14 DIP 间距；不做拖拽分割条，避免原生分隔条破坏视觉。
ui.Row(c).Fill().Gap(14).AlignItems(ui.Stretch).Children(func() {
    ui.Box(c).Width(280).FillHeight().Children(func() { a.groupPanel(c) })
    ui.Box(c).Grow(1).FillHeight().Children(func() { a.groupMain(c) })
})
```

- 左栏：`ui.Sidebar`（或 `SidebarBase`）承载「全部项目 / 各组 / 未分组」；每项带组色文件夹 SVG + 计数 Badge；底部「新建项目组」描边按钮。
- 右栏：组头（46 DIP 色块 + 名称 + 描述 + 计数 + 操作）；`ui.Tabs` 切「项目 / 文档」。
- 项目表格：`ui.Table` + `TableColumn`（项目自适应、版本 78、标签 142、更新时间 84、操作 26）；`ListState.Key` 用项目 ID，行菜单用 `CellMenu`/`ContextMenu`。行点击用单元格 `RowClicked()` 或表体 `Changed()` 读取 `Selected` 后进入详情；键盘 Enter 用 `Submitted()` 覆盖。
- 文档表格列：文档自适应、类型 108、字数 82、更新时间 104、操作 26。
- 组文档页签底部放「组文档是组级知识库」提示卡（仅视觉辅助，不阻断）。
- 空组、无组、搜索无结果分别走 `emptyState`。

#### 项目详情

- 面包屑 + 返回；头部 52 DIP 项目图标、名称、版本 chip、标签、操作（打开目录/编辑/更多）。
- 主体 `ui.Row`：文档区 `Grow(1)` + 信息栏 320 DIP（<1280 时 288）。
- 文档区：搜索框 + 新建文档 + `ui.Table`（标题/类型/字数/更新时间/操作）。
- 信息栏：`ui.Field` 风格的只读行；路径用等宽字体 + 复制按钮（`c.WriteClipboard(p.Path)`）。
- 窗口宽度 < 1280 时：左栏收至 240 DIP，项目表格隐藏「标签」列（标签仍可在详情查看），信息栏收至 300 DIP；表格与卡片不出现横向滚动。
- 打开目录：

```go
if _, err := os.Stat(p.Path); err != nil {
    c.Toast("项目路径不存在，无法打开")
    return
}
go func() {
    if err := mygo.Shell.OpenPath(p.Path); err != nil {
        // 回到 UI 线程提示
    }
}()
```

#### 文档编辑器

- 顶部：返回、面包屑、标题 `ui.TextInput`（大字）、类型 `ui.Select`、保存 `ui.PrimaryButton`。
- 「编辑 / 预览」用 `ui.Segmented`；编辑态左右各占一半（固定 50/50，不做拖拽分隔）。
- 源码：`ui.TextArea`（等宽字体，行号可后续加）；预览：`markdown.View(c, content, nil)`。
- 状态栏：字数、保存状态、`Ctrl+S`；未保存离开用 `ui.AlertDialog` 二次确认。
- 超过 20,000 字符时预览延迟 300ms 更新（`c.After` 或 goroutine），避免每帧重排 Markdown。

#### 弹窗与表单

```go
ui.DialogBase(c, &a.projectDialogOpen, func(backdrop, panel ui.Element) {
    panel.Width(640).Children(func() {
        ui.Form(c, func() {
            ui.Field(c, "项目名称", func() {
                ui.TextInput(c, &a.projectForm.Name)
            }).Description(a.fieldErrs["name"]).Error(a.fieldErrs["name"] != "")
            // 组 / 版本 / 标签 / 路径 / 备注 ...
        })
    })
})
```

- 项目弹窗：640 DIP；水平表单（标签左列 + 控件），标签 `ui.TokenField`，路径 `TextInput` + 浏览按钮（`mygo.Dialog.Open` 选目录，goroutine + `win.Update` 回填）。
- 项目组弹窗：520 DIP；名称、描述、6 个色块（自绘 swatch，选中圆环）。
- 删除确认：项目用 `ui.AlertDialog`；项目组用 `DialogBase`，正文展示「N 个项目移至未分组 / M 篇组文档删除」，确认按钮为 danger 色。
- 焦点：打开时聚焦第一个输入框；Esc 关闭；遮罩点击不关闭（与需求 7.5 一致）。

#### 搜索

- 顶栏 `ui.SearchField` 绑定 `app.query`；输入 ≥2 字符后 150ms 防抖查询。
- 结果用 `ui.Popover` 或页面内下拉面板，分「项目 / 文档 / 项目组」三组，各最多 5 条。
- ↑↓ 选择、Enter 打开、Esc 关闭；结果为空显示「未找到与 “xxx” 相关的内容」。
- 搜索 SQL 复用 store 的 `Search`，避免 UI 层拼条件。

### 7.4 主题与图标

- `internal/app/theme.go` 把设计 Token 映射为 MyGo Theme：主色 `#7C5CFF`、背景 `#F4F5F7`、表面 `#FFFFFF`、文本四级、边框 `#E3E5E8`、控件圆角 4 / 卡片 6、间距基数 4。
- 所有尺寸用 `t.Space(n)` / `t.Rem(n)` 派生；固定值只出现在布局契约（侧栏 228、顶栏 56、左栏 280、信息栏 320、无边框标题栏 34）并注释原因。
- 图标在 `icons.go` 里用 `sync.OnceValue` 缓存 `ui.MustParseSVG` 结果；不每帧解析 SVG。
- 卡片用 `Box` + `Background(t.Surface)` + `Border(1, t.Border)` + `Radius(t.Radius)`；阴影能力不足时以边框和留白替代，不堆叠重投影。
- 字体：将 Inter / Noto Sans SC 的子集放入 `resources/`，用 `ui.RegisterFont` 注册；系统字体仅作回退，避免中文落到衬线体。
- 主题固定浅色（v1.0 不跟随系统深色），但 Token 结构保留 `dark` 映射，供 v1.1 切换。

### 7.5 异步加载与防陈旧

```go
func (a *App) loadProjects() {
    a.loadID++
    id := a.loadID
    filter := a.currentFilter()
    go func() {
        ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
        defer cancel()
        list, err := a.svc.Projects(ctx, filter)
        a.win.Update(func() {
            if id != a.loadID { return } // 迟到的结果丢弃
            a.projects, a.loadErr, a.loading = list, err, false
            a.win.Invalidate()
        })
    }()
}
```

- `loadID` 只在 UI 线程读写；goroutine 只读拷贝。
- 初始加载：打开数据库/迁移同步完成（< 50ms），窗口先显示骨架，随后异步加载首页数据。
- 所有写操作完成后调用对应的 reload；失败保留旧列表并提示，不清空。

### 7.6 快捷键与无障碍

| 快捷键 | 行为 |
|---|---|
| `Ctrl/Cmd+N` | 打开新建项目弹窗 |
| `Ctrl/Cmd+Shift+N` | 打开新建项目组弹窗 |
| `Ctrl/Cmd+F` | 聚焦顶栏搜索 |
| `Ctrl/Cmd+S` | 保存文档（编辑器内） |
| `Esc` | 关闭弹窗/退出搜索 |
| `Enter` | 表单提交 / 打开选中结果 |
| `↑↓` | 搜索结果与列表移动 |

- 所有图标按钮 `.Label("…")`；表格 `.Label("项目列表")`；表单 `Field` 自带标签关联。
- 焦点顺序：侧栏 → 顶栏搜索 → 主操作 → 内容区；弹窗打开后焦点锁在弹窗内。
- 状态变化（保存成功、删除完成）用 `c.Toast` 或 `c.Announce`，不只写日志。

---

## 8. 关键流程

### 8.1 启动

```
main → mygo.App.WhenReady
  → app.Start:
      PathUserData → sql.Open("sqlite") → PRAGMA → migrate
      service := service.New(store)
  → NewWindow(StateKey, MinWidth/Height, Content: ui.View(app.View))
  → app.AttachWindow(win)
  → go app.loadInitial() → win.Update → 首帧数据
  → mygo.App.Run()
```

失败处理：数据库打不开或迁移失败时，用 `mygo.Dialog.Error` 显示路径与错误，然后退出；不带着半初始化状态建窗。

### 8.2 新建项目

```
点击新建 → projectDialogOpen=true，form 重置，聚焦名称
填写 → 绑定 projectForm
提交 → service.CreateProject
  ├─ 校验失败：fieldErrs 写入，弹窗不关，聚焦首个错误字段
  ├─ 唯一冲突：fieldErrs["name"] = "该组下已存在同名项目"
  └─ 成功：关闭弹窗 → reloadGroups/Projects/Dashboard → c.Toast("已创建项目")
```

### 8.3 删除项目组

```
点击删除 → service 取组计数 → deleteDialogOpen=true，展示影响数量
确认 → service.DeleteGroup（事务：迁项目 → 删组文档 → 删组）
  ├─ 失败：toast + 保留弹窗，允许重试
  └─ 成功：关闭弹窗 → 选中「未分组」→ reload → toast
```

### 8.4 编辑文档

```
打开文档 → 载入 title/kind/content/link → PageEditor
编辑 → TextArea 绑定 content；预览按 300ms 防抖刷新
保存 → service.SaveDocument（校验 + 唯一性）
  ├─ 失败：保留内容，状态栏显示“保存失败”，toast 可重试
  └─ 成功：状态栏“已保存 HH:mm”，返回来源页
返回/关闭 → 若 dirty，AlertDialog「放弃未保存的修改？」
```

### 8.5 全局搜索

```
输入 ≥2 字符 → 150ms 防抖 → service.Search（项目名/标签/路径/文档标题/组名）
→ 分组结果面板 → ↑↓ 选择 → Enter 打开详情/编辑器/组页
```

---

## 9. 测试计划

### 9.1 Store（内存 SQLite）

| 用例 | 断言 |
|---|---|
| 迁移幂等 | 连续 `Migrate` 两次不报错，版本表只有 1 行 |
| 组 CRUD | 创建/读取/更新/删除；名称唯一冲突 |
| 项目 CRUD | 组内重名冲突；跨组同名允许；`group_id` 为空允许 |
| 标签 JSON | 中文/特殊字符往返一致；空数组不为 NULL |
| 删除项目 | 项目与其文档均删除，组文档不受影响 |
| 删除组 | 项目 `group_id` 变 NULL；组文档删除；项目文档保留 |
| 统计 | 项目/组/文档/本周更新数与插入数据一致 |
| 搜索 | 命中名称、标签、路径、文档标题；LIMIT 生效 |
| 时间 | `updated_at` 在写操作后严格递增 |

### 9.2 Service

- 名称/标签/备注/路径/版本的边界值（0、1、上限、上限+1）。
- 全角/中文 rune 计数。
- 唯一性预检与数据库冲突双路径。
- 删除组/项目前的计数与确认文案。
- 文档 owner 不合法时拒绝。

### 9.3 UI（`ui.NewTester`）

| 用例 | 操作 |
|---|---|
| 首页渲染 | 1440×900 与 1120×720 各一帧；统计/表格/文档都在视口内 |
| 侧栏导航 | 点击「项目组管理」→ `PageGroups`；当前项高亮 |
| 组切换 | 点击「后端服务」→ 右栏标题、项目数、表格行变化 |
| 页签切换 | 点「文档」→ 文档表出现；点「项目」→ 项目表恢复 |
| 项目详情 | 点击项目行 → 面包屑/标题/文档列表正确 |
| 新建项目弹窗 | 点顶栏按钮 → 弹窗可见；空名称提交 → 错误提示，弹窗不关闭 |
| 删除确认 | 点删除 → 影响数量文案正确；取消不删 |
| 编辑器 | 打开文档 → 标题/正文载入；保存 → toast；返回未保存 → 确认弹窗 |
| 空状态 | 空组/搜索无结果显示引导，不显示空表格 |
| 快捷键 | Ctrl+N / Ctrl+F 在 tester 的快捷键路径可触发 |

### 9.4 手动验证（无头测试覆盖不到）

- 真实文件选择器（`mygo.Dialog.Open`）在 Windows/macOS/Linux 的行为。
- `mygo.Shell.OpenPath` / `ShowItemInFolder` 定位目录。
- 中文 IME 在 `TextInput` / `TextArea` 的候选词与组合输入。
- 高 DPI（125% / 150% / 200%）下侧栏、表格、弹窗不裁切。
- 窗口关闭后重新打开，`StateKey` 恢复尺寸；最小尺寸下无重叠。
- 打包后的 SQLite 原生模块加载（不能只在 `go run` 下验证）。

### 9.5 命令

```sh
go build ./...
go vet ./...
go test ./...
CGO_ENABLED=0 go test ./...     # 发布构建不能依赖 cgo
go tool mygo vet .
go tool mygo dev -background    # 手动检查原生窗口
go tool mygo build              # 打包三平台产物
```

---

## 10. 性能与可靠性

| 关注点 | 措施 |
|---|---|
| 列表首屏 | 索引覆盖 `group_id` / `updated_at`；`ui.Table` 只构建可见行；查询 `LIMIT 200` |
| 首页聚合 | 一次 `Dashboard` 查询（子查询计数），不做 4 次往返 |
| 搜索 | 150ms 防抖；`LIMIT 20`；输入 <2 字符不查 |
| 文档预览 | 300ms 防抖；超过 20k 字符降级为按需渲染；编辑页不每帧重解析 |
| 启动 | 迁移与 `PRAGMA` 同步完成；数据异步加载，先骨架后内容 |
| 写操作 | 事务包裹；失败回滚；UI 保留输入并允许重试 |
| 数据安全 | 迁移前如检测到旧版本，先复制 `mana.db` 为 `mana.db.bak-<timestamp>`；不覆盖未知 schema |
| 崩溃恢复 | WAL 模式；应用异常退出后 SQLite 自动恢复；启动时校验 `schema_migrations` |
| 内存 | 不缓存全量文档正文；列表只载标题/类型/字数/时间；正文按需读取 |

---

## 11. 里程碑与排期

单人开发，单位为「人日」，含自测；缓冲 2 人日。

| 里程碑 | 内容 | 工期 | 出口条件 |
|---|---|---|---|
| M0 脚手架 | `mygo init -template native`、`mygo.json`、主题骨架、图标缓存、CI 脚本 | 0.5 | `go run .` 出现 1440×900 窗口与侧栏 |
| M1 数据层 | model、`database/sql` + modernc SQLite、迁移、groups/projects/documents 仓储、统计 | 1.5 | Store 测试全绿；迁移幂等；CRUD 可用 |
| M2 应用骨架 + 首页 | App 状态、根布局、侧栏/顶栏、首页统计/最近项目/组概览/最近文档 | 1.5 | 首页在 1440/1120 下不裁切；Tester 通过 |
| M3 项目组管理 | 左右两栏（280 + 14 间距）、组列表、组 CRUD、组文档页签、项目表 | 2.0 | AC-02 通过；删除组事务测试通过 |
| M4 项目管理 | 项目 CRUD、详情、移动、删除、打开目录/复制路径 | 1.5 | AC-01 通过 |
| M5 文档管理 | 项目/组文档 CRUD、编辑器、Markdown 预览、保存状态 | 2.0 | AC-03 通过；预览 20k 字符不卡 |
| M6 搜索与打磨 | 全局搜索、快捷键、空状态、错误反馈、设置/备份 | 1.5 | AC-04 通过；所有空态有引导 |
| M7 测试与交付 | 性能、跨平台构建、打包、文档更新、缺陷修复 | 1.5 | AC-05 通过；`CGO_ENABLED=0 go test ./...` 绿 |
| 缓冲 | 平台差异、IME、打包问题 | 2.0 | 不挤压 P0 范围 |

**总排期**：10.5 + 2 = 12.5 人日；按每周 5 天计约 2.5 周。若只能投入半天/天，按里程碑顺序交付，M3 与 M4 可并行度低（共享 App 状态），不建议多人同时改同一批文件。

### 11.1 每周节奏建议

| 周 | 交付 |
|---|---|
| 第 1 周 | M0–M2：能启动、能看首页、数据库可读写 |
| 第 2 周 | M3–M5：项目组、项目、文档三条主链路可操作 |
| 第 3 周（半周） | M6–M7：搜索、打磨、测试、打包 |

---

## 12. 风险与对策

| 风险 | 影响 | 对策 |
|---|---|---|
| MyGo v0.9.0 API 与设计预期有差异（表格行点击、DialogBase、主题字段） | 返工 UI | M0 做 0.5 天 API spike：表格行点击、弹窗、TokenField、markdown.View、Tester 各写一个最小样例 |
| SQLite 原生模块在目标平台/打包后加载失败 | 应用不可用 | M1 就在 Windows + 一个第二平台跑 `go tool mygo build` 并启动；打包后验证 `mana.db` 创建 |
| 删除组的跨表事务漏删/多删 | 数据不一致 | 事务 + 专门的 Store 测试（项目迁移、组文档删除、项目文档保留） |
| 中文输入法在 TextInput/TextArea 中异常 | 无法输入 | M5 用真实 IME 手动验证；必要时改用 `TextInputBuffer` / `TextAreaBuffer` |
| 大文档 Markdown 预览卡顿 | 编辑器掉帧 | 300ms 防抖 + 20k 字符降级 + 单栏预览；用 200k 字符做压力测试 |
| 窗口最小尺寸下控件重叠 | 不可用 | `ui.NewTester` 在 1120×720 断言关键元素框；侧栏可收、表格列可隐藏 |
| 设计 Token 与 MyGo Theme 不完全一一对应 | 视觉偏差 | 以 `theme.go` 为单一映射；不能支持的阴影用边框替代，记录偏差 |
| 数据库迁移损坏旧数据 | 数据丢失 | 迁移前备份；迁移单事务；失败不启动写操作 |
| 需求中「项目组增删开发」理解偏差 | 范围偏差 | 已按 CRUD 实现（A-01），评审时确认 |

---

## 13. 完成定义（DoD）

一个功能只有同时满足以下条件才算完成：

1. 需求条目与验收标准（AC-01..AC-05）有对应实现与测试。
2. `go build ./...`、`go vet ./...`、`go test ./...`、`CGO_ENABLED=0 go test ./...` 全部通过。
3. 新增/修改的交互有 `ui.NewTester` 测试；关键 Store 逻辑有内存库测试。
4. 1440×900 与 1120×720 两个尺寸无裁切、无重叠、无横向滚动。
5. 成功/失败都有界面反馈；删除有二次确认与影响说明。
6. 没有把 SQL、文件 IO、系统对话框放到 UI 帧内。
7. 文档更新：需求变更同步 `requirements.md`；设计偏差同步 `design/README.md`。
8. 打包产物在至少一个目标平台启动成功，数据库落在用户数据目录。

---

## 14. 交付物清单

- 可运行源码（`main.go`、`internal/…`）。
- `mygo.json`、`resources/icon.png`、打包说明。
- 测试：`internal/store/*_test.go`、`internal/service/*_test.go`、`internal/ui/*_test.go`。
- 文档：`docs/requirements.md`、`docs/development-plan.md`、`design/README.md`。
- 设计：`design/mana-prototype.html` + 8 张 1440×900 设计图。
- 一份已知问题与平台验证记录（发布前补）。

---

## 15. 后续版本（不在本轮）

| 版本 | 候选 |
|---|---|
| v1.1 | 深色主题、数据导入导出、全局文档中心、附件与图片、最近访问历史 |
| v1.2 | Git 仓库状态、README 预览、项目健康度、批量操作 |
| v2.0 | 多用户/同步（如确有需求）、插件化文档渲染、命令面板 |
