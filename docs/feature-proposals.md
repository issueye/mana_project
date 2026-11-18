# Mana 功能建议与设计 · v2.1

> 日期：2026-11-18 ｜ 基线：`docs/requirements.md` v1.0、`docs/development-plan.md` v1.0、`design/README-v2.md`（索引卡重设计）
> 状态：**F1–F4 已在本轮实现**，详见 §4 各节的「实现状态」与 §6 测试记录。

---

## 1. 现状盘点

### 1.1 已完成

| 能力 | 位置 | 状态 |
|---|---|---|
| 项目 / 项目组 / 文档 CRUD + 级联删除 | `internal/store`、`internal/service` | ✅ 有测试 |
| `MoveProject`（数据层） | `store/projects.go`、`service.go:117` | ✅ 已实现，**UI 无入口** |
| Markdown 写作纸编辑器（行号、预览、二次确认） | `internal/app/editor.go` | ✅ |
| 索引卡视觉（图标轨 + 目录 + 状态条） | `internal/app/shell.go` 等 | ✅ v2 重设计已落地 |
| 快捷键 CTRL N / SHIFT+N / F / S | `app.go shortcuts` | ✅ |
| 迁移、WAL、外键、内存库测试 | `internal/store/store.go` | ✅ |

### 1.2 缺口（开发计划 §0 实施记录明确的 P1 未收尾项 + UI 占位）

| # | 缺口 | 证据 |
|---|---|---|
| G1 | 全局搜索无 `store.Search` / `service.Search`；顶栏搜索只做页内转发 | 开发计划实施记录「全局文档搜索…尚未收尾」；FR-06 |
| G2 | 设置与备份完全缺失；顶栏 ⚙、轨底 ⚙ 均弹「后续版本开放」 | FR-07；`shell.go barIcon/railGhost` |
| G3 | 行级「移入组」无 UI 入口（数据层已就绪） | 开发计划实施记录 |
| G4 | 统计页占位（图标轨 grid 图标弹 toast） | `shell.go railItem("统计")` |
| G5 | 顶栏 `CTRL K` 命令提示是死入口 | `shell.go barKey` |

---

## 2. 功能建议清单

按「用户价值 × 工作量」排序；P0 为本轮实现。

| ID | 功能 | 优先级 | 理由 | 工作量 |
|---|---|---|---|---|
| F1 | **全局搜索 · 命令面板（CTRL K）** | P0 | FR-06 明确要求；数据量上来后是主要检索路径；同时救活 CTRL K 死入口 | 1 人日 |
| F2 | **项目移入组** | P0 | 数据层已实现，只差 UI；修正分组是登记后的高频修正动作 | 0.5 人日 |
| F3 | **设置与数据备份** | P0 | FR-07；本地数据库产品必须有「我的数据在哪、能否带走」的答案 | 0.5 人日 |
| F4 | **统计页** | P0 | 图标轨已有占位；让「索引」有全局视图，纯内存聚合零新 SQL | 0.5 人日 |
| F5 | 文档回收站（删除软标记 + 恢复） | P1 | 误删保护，需要 schema 迁移（`deleted_at`） | 1.5 人日 |
| F13 | **附件上传与查看（PDF / Word / Excel / PPT / 图片 / Markdown）** | **P0（2026-11-19 追加）** | 文档不止 Markdown：登记产物与资料原件，应用内可看，系统程序可打开 | 1.5 人日 |
| F14 | **项目文件系统（文件浏览器）** | **P0（2026-11-19 追加）** | 项目登记的是磁盘路径；不看代码文件，索引就只有元数据 | 1 人日 |
| F15 | **代码仓库检测与管理（Git）** | **P0（2026-11-19 追加）** | 状态、暂存、提交、历史、标签发版、推送——把「项目管理」落到工作区 | 1.5 人日 |
| F6 | 文档全文搜索（正文 LIKE / FTS5） | P1 | FR-04-8 的正文部分；数据量大后需要 FTS | 1 人日 |
| F7 | 项目置顶与排序记忆 | P1 | 手动 `sort_order` 列 + 拖拽或快捷操作 | 1 人日 |
| F8 | 导入导出（JSON 快照导出 / 导入） | P1 | 换机、分享模板；与备份互补 | 1 人日 |
| F9 | 深色主题 | P1 | design/README-v2 §8 已预留 Ink/Sheet 反转 | 0.5 人日 |
| F10 | 文档模板（新建时选预置骨架） | P2 | 需求/设计/计划三类模板，纯 UI | 0.5 人日 |
| F11 | 最近访问历史 | P2 | 开发计划 v1.1 候选；需新表 | 1 人日 |
| F12 | Git 仓库状态徽标 | P2 | 开发计划 v1.2 候选；依赖磁盘路径探测 | 2 人日 |

> 取舍：F5–F12 本轮不做；F5/F6 涉及 schema 迁移与查询语义，放在下一轮单独评审。

---

## 3. 共性设计约定

- 视觉：全部沿用 `design/README-v2.md` 的纸墨令牌；新弹层复用 `dialogShell`（3px 朱砂条 + 2px 圆角 + 56% 墨色遮罩）。
- 架构：SQL 只进 `internal/store`；校验进 `internal/service`；UI 不拼 SQL、不做文件 IO（对话框与备份走 goroutine + `win.Update`）。
- 无障碍：新交互均带 `Label`；键盘路径与鼠标路径等价。
- 验收：新增交互有 `ui.NewTester` 用例；store 逻辑有内存库用例；1440×900 与 1120×720 不裁切。

---

## 4. 本轮功能设计

### 4.1 F1 全局搜索 · 命令面板（CTRL K）

**形态**：模态命令面板（640px，顶部搜索框自动聚焦），替代「下拉结果」——桌面端一次键盘调用即达，同时承载全局动作（新建项目/组、备份、设置、统计），救活 G1+G5。

**数据层**（`store/search.go` 新建）：

```go
type SearchResult struct {
    Projects []model.Project   // 命中 name / path / tags JSON，LIMIT 5
    Groups   []model.Group     // 命中 name，LIMIT 5
    Docs     []model.Document  // 命中 title（元数据，不含正文），LIMIT 5
}
func (s *Store) Search(ctx context.Context, q string) (model.SearchResult, error)
```

- SQL：`LIKE '%'||?||'%'`（COLLATE NOCASE）；标签用 `tags LIKE '%"q"%'`；每类独立 LIMIT；`q` 经 `strings.TrimSpace`，空串返回空结果。
- 大小写：SQLite `LIKE` 对 ASCII 不区分大小写；中文不受影响。

**服务层**：`service.Search` 透传（查询词 ≤ 80 字符校验）。

**UI 层**（`internal/app/palette.go` 新建）：

| 项 | 设计 |
|---|---|
| 打开 | `CTRL K`（全局快捷键）、顶栏 CTRL K 芯片点击 |
| 状态 | `paletteOpen bool`、`paletteQuery string`、`paletteResults model.SearchResult`、`paletteSel int`（跨行选择索引）、`paletteSearched string`（已查询词） |
| 防抖 | 查询词变化时起 goroutine `sleep 150ms` 后查库，回来对比 `paletteQuery` 未变才 `win.Update` 回填；输入 <2 rune 用内存过滤兜底（仅项目/组/文档名前缀） |
| 结果 | 三组：项目（名 + 组名/路径）、文档（标题 + 归属）、项目组（名 + 计数）；随后「操作」组：新建项目、新建项目组、打开统计、备份数据、打开设置 |
| 键盘 | ↑↓ 移动（跨组扁平索引）、Enter 打开、Esc 关闭（DialogBase 自带遮罩 Esc） |
| 打开动作 | 项目 → 详情；文档 → 编辑器；组 → 组页；动作 → 对应入口 |
| 空态 | 「未找到与 “xxx” 相关的内容」 |

**验收**：
1. CTRL K 打开；输入命中项目名/路径/标签/文档标题/组名，三类结果 ≤5 条；
2. ↑↓+Enter 与点击行为一致；
3. 无结果显示空态文案；
4. 打开结果后面板关闭、页面正确。

### 4.2 F2 项目移入组

**入口**（两处）：
1. 项目详情「记录卡」的「所属项目组」行尾部 ghost 按钮「移动」；
2. 项目组页项目表格行尾 `⋯` 按钮 → 直接打开移动弹层（当前行为是死图标）。

**弹层**（480px，复用 dialogShell）：
- 标题「移动 “X” 到…」，`ui.Select` 列出 全部组 + 未分组，预选当前组；
- 提示行：目标组的现有项目数；确认 → `service.MoveProject(id, groupID)`；
- 成功：关弹层 + reload + toast「已移动到 Y」；若当前在详情页则刷新 `currentProject`；失败 toast。

**验收**：
1. 记录卡与组表格两个入口都能打开弹层，预选正确；
2. 移动后组计数、记录卡、状态条同步变化；
3. 移到「未分组」合法（`groupID=nil`）。

### 4.3 F3 设置与数据备份

**形态**：560px 设置弹层（⚙ 两处入口均打开它），两个分区。

| 分区 | 内容 |
|---|---|
| 数据 | 数据文件路径（等宽字，末尾「在文件管理器中显示」按钮）；数据库大小（os.Stat，打开弹层时异步取）；「立即备份」主按钮 |
| 关于 | Mana v1.0.0 · Go + MyGo 原生 UI · 本地 SQLite · MIT |

**备份**：
- `store.Backup(ctx, dest string) error`：`VACUUM INTO ?`（单语句生成一致性快照，不打断使用；modernc.org/sqlite 支持）。dest 已存在时先报错（SQLite 语义），UI 预检同名文件。
- service.Backup 校验 dest 非空、父目录存在。
- UI：点「立即备份」→ `mygo.Dialog.Open(Directory)` 选目录（goroutine）→ 目标名 `mana-backup-20060102-150405.db` → 执行 → toast「已备份到 …」/ 失败 toast 明确错误（FR-07-3）。
- 「在文件管理器中显示」：`mygo.Shell.OpenPath(数据目录)`。

**验收**：
1. 设置弹层显示真实路径与大小；
2. 备份生成的文件可用 SQLite 打开且包含当前数据（store 测试断言备份库可查询出同量项目）；
3. 失败（目录不可写）有明确提示。

### 4.4 F4 统计页

**入口**：图标轨 grid 图标（替换 toast 占位）。

**布局**（沿用首页语言，无卡片墙）：
1. 页眉：面包屑「工作台 / 统计」+ 标题「数据统计」+ 一句话口径说明；
2. 账本 4 格：项目 / 项目组 / 文档 / 本周更新（复用 `ledgerCell`）；
3. 项目组分布：每行 组色点 + 组名 + 项目数条 + 文档数条 + 等宽数字；
4. 文档类型分布：复用 `docTypeLedger`；
5. 近 7 天活跃：按天的（项目更新数 + 文档改动数）双列迷你条；
6. 标签 Top：前 8 个标签 + 计数。

**数据**：全部从 `Snapshot` 内存聚合（新 `stats.go` 辅助函数），零新 SQL。

**上下文目录**：显示文档类型分布 + 组列表（复用现有组件）。

**验收**：
1. 轨道 grid 进入/离开统计页，active 态正确；
2. 数字与首页账本一致（同一 `stats()` 来源）；
3. 1120×720 无横向滚动。

---

## 5. 实现顺序

1. `store/search.go` + service.Search + 命令面板（F1，最大件）
2. F2 移入组弹层（顺手接 F1 的动作项）
3. F3 设置 + 备份（store.Backup + dialog）
4. F4 统计页（纯 UI）
5. 测试与截图对照，更新本文档状态
6. （2026-11-19 追加）F13 附件上传与查看，见 §4.5
7. （2026-11-19 追加）F14 项目文件系统 + F15 Git 仓库管理，见 §4.6

---

### 4.6 F14 项目文件系统 + F15 Git 仓库管理

两个功能共用一个入口：**项目详情页主列改为页签**「文档 / 文件 / 仓库」（`ui.Tabs`，`a.projectView`）；右栏记录卡不变。

#### F14 文件浏览器

- **导航**：当前目录面包屑（等宽字，项目根 ▸ 子目录）+「上级」+「在文件管理器中显示」+「刷新」。
- **读取**：`os.ReadDir` 离帧执行，带代际防陈旧（`filesGen`）；目录项含 名称/大小/修改时间；**目录在前、文件在后**，名称升序；默认隐藏 `.` 开头条目（`.git`、`.idea` 等）。
- **交互**：点文件夹进入；点文件/行尾按钮用系统程序打开；路径**永不越出项目根**（`filepath.Rel` 校验，越界回落根目录）。
- **空态**：未登记路径 → 引导去编辑项目登记；目录为空 / 读取失败（目录不存在）→ 明确提示。

#### F15 Git 仓库管理（系统 git CLI）

**选型**：`exec.Command("git", args…)` 调用系统 git（新包 `internal/gitcmd`），不用 go-git——
1. 推送凭证交系统 Credential Manager（go-git 无法复用，反而要处理交互）；
2. 目标用户是开发者，git 必在；
3. 参数一律走 argv，不经过 shell，无注入面；输出超时 10s（推送 120s）。

**能力**：

| 命令 | 实现 |
|---|---|
| 检测 | `<path>/.git` 存在即视为仓库；面板按需异步检测 |
| 状态 | `status --porcelain=v1 -z --branch`（`core.quotepath=false` 保中文路径）；分支行 + 每文件 XY 状态 |
| 暂存 | 单文件 `add -- <path>`；「全部暂存」`add -A` |
| 提交 | `commit -m <msg>`（消息必填） |
| 历史 | `log -n 50 --pretty=%h%x09%an%x09%at%x09%s` |
| 标签 | `for-each-ref refs/tags`（名称/hash/日期/说明）；`tag -a <name> -m <msg>` |
| 初始化 | `init`（未检测到仓库时给入口） |
| 推送 | `push -u origin HEAD` + `push origin <tag>`（有 origin 时；无远端提示先在终端添加） |

**UI**（仓库页签，全部离帧 + `win.Update` 回填，代际防陈旧）：

```
分支芯片（mono + 朱砂描边）· 远端状态 · 刷新
─ 改动 (N)：状态字母芯片（M/A/D/??）+ 等宽路径 + [暂存|取消] ／「全部暂存」
─ 提交框：消息输入 + [提交]（无消息/无暂存禁用态提示）
─ 历史 01–50：hash(朱砂 mono) · 主题 · 作者 · 相对时间
─ 标签/发版：列表 + 名称/说明输入 + [创建标签] + [推送分支与标签]
```

- 空仓（无提交）：历史区提示「还没有提交」；提交后自动刷新。
- 非 git 目录：空态 +「初始化仓库」按钮。

**验收**：

1. 文件页签能进入子目录、返回、打开文件，面包屑同步；路径越界不可能；
2. 改动文件列表与 `git status` 一致；暂存→提交→历史立即可见新提交；
3. 创建标签出现在列表；推送在无远端时给出明确提示而不是静默失败；
4. 中文文件名不乱码；长列表不卡帧（历史限 50 条）；
5. 所有命令超时有错误提示；UI 线程无阻塞调用。

### 4.5 F13 附件上传与查看（PDF / Word / Excel / PPT / 图片 / Markdown）

**目标**：文档分两种载体——笔记（Markdown 正文）与附件（上传的文件）。附件保存在数据目录内（`attachments/<文档ID>/<原文件名>`），应用内可预览，系统程序可打开；删除文档时附件一并清理。

**数据层**（迁移 `002_attachments.sql`）：

```sql
ALTER TABLE documents ADD COLUMN file_name TEXT NOT NULL DEFAULT '';
ALTER TABLE documents ADD COLUMN file_size INTEGER NOT NULL DEFAULT 0;
```

- `model.Document` 增加 `FileName string`、`FileSize int64`；`documentSelect` 与全部扫描同步。
- `store.SaveFile(ctx, in DocumentInput, srcPath string) (Document, error)`：插行 → 拷贝文件到 `attachments/<id>/<清洗后的原名>` → 回写 `file_size`；拷贝失败回滚删行。
- `store.SetFilesDir(dir)` 由 `app.Start` 指到用户数据目录；`:memory:` 测试指到临时目录。

**提取层**（新包 `internal/extract`，全部纯 Go、离线）：

| 类型 | 提取方式 | 应用内查看 |
|---|---|---|
| `.docx` | 解包 OOXML `word/document.xml`，按 `<w:p>` 段落收集 `<w:t>` | 阅读视图 |
| `.pptx` | 解包 `ppt/slides/slideN.xml`，按页收集 `<a:t>` | 分页阅读视图 |
| `.xlsx` | 解包 `xl/worksheets` + `sharedStrings.xml`，逐 sheet 生成单元格矩阵 | 简易表格（sheet 切换 + `ui.Table`） |
| `.pdf` | `github.com/ledongthuc/pdf`（纯 Go）逐页提取文本 | 阅读视图；失败则提示用系统程序打开 |
| `.md` | 原文 | `markdown.View` 渲染 |
| `.png/.jpg/.gif/.webp/.bmp` | — | `ui.Image` 原生显示（`ui.DecodeImage`） |
| 其他 | — | 文件卡 + 「在系统中打开」 |

- 提取文本（≤2 万字）存入 `content`，供列表摘要与后续全文搜索；预览按需重读文件，不依赖 content。

**服务层**：

- `service.SaveDocumentFile(ctx, ownerType, ownerID, srcPath)`：校验归属、扩展名与大小（≤100 MB）→ `extract` → `store.SaveFile`。
- `service.DeleteDocument` 在删行后清理 `attachments/<id>/`（best effort）。
- `service.DocumentFilePath(d)` 返回附件绝对路径，供「在系统中打开」。

**UI 层**：

- 上传入口：项目文档区与组文档页签头部各加「上传文档」描边按钮 → `mygo.Dialog.Open`（PDF/Word/Excel/PPT/图片/Markdown 过滤器）→ goroutine 上传 → reload + toast。
- 文档表格：附件行显示类型徽标（PDF/Word/Excel/PPT/图片）、大小列以 `KB/MB` 代替字数。
- 编辑器页对附件切换为**文件查看器**：标题/类型仍可编辑保存，正文区按类型渲染（见上表）；右上角「在系统中打开」。
- 记录卡「文档分布」等统计不变（附件计入文档数）。

**验收**：

1. 上传后列表立即出现，类型徽标与大小正确；数据目录出现 `attachments/<id>/`；
2. docx/pptx/pdf 阅读视图显示提取文本；xlsx 表格可切 sheet；图片直接显示；md 走渲染；
3. 「在系统中打开」调起系统程序；
4. 删除附件文档后 `attachments/<id>/` 被清理；
5. 提取失败（如加密 PDF）不阻断：显示文件卡与打开入口。

## 6. 测试记录（2026-11-19 实测）

- [x] `go build ./...`、`go vet ./...`、`go test ./...` 全绿
- [x] Store：`TestSearchFindsProjectsDocsGroups`（名称/标签/路径/文档标题/组名/空查询/未命中）、`TestSearchLimitsEachSection`（每类上限 5）
- [x] Store：`TestBackupCreatesQueryableCopy`（备份文件可被新连接打开且数据一致）、`TestBackupRefusesExistingDestination`、`TestBackupRejectsEmptyDestination`
- [x] UI：`TestCommandPalette`（CTRL K 打开 → 搜索命中项目与文档 → ↓ 选中 → ENTER 打开编辑器）、`TestCommandPaletteRunsActions`（空查询 ENTER 执行「新建项目」）
- [x] UI：`TestMoveProjectDialog`（记录卡入口 + 弹层预选当前组 + 确认后组名更新）、`TestSettingsDialog`（路径/大小/备份/关于俱全）、`TestStatsPage`（账本/组分布/周活跃/标签 Top）
- [x] 截图：`internal/app/testdata/screens/{palette,move-dialog,settings,stats}.png`

| F4 统计页 | ✅ | `app/stats.go`（零新 SQL，内存聚合） | `TestStatsPage` |
| F13 附件上传与查看 | ✅ | 迁移 `002_attachments.sql`、`store.SaveFile/RemoveAttachment/AttachmentFile`、`internal/extract`（docx/pptx/xlsx/pdf/md，纯 Go）、`service.SaveDocumentFile`、`app/attachments.go`（上传流程 + 文件查看器） | extract 5 件、store 3 件、service 2 件、`TestFileDocumentViewer` |
| F14 项目文件系统 | ✅ | `app/files.go`（异步目录读取 + 代际防陈旧 + 越界防护 + 面包屑）；入口：项目详情页「文件」页签 | `TestProjectFilesTab`（列根/进目录/返回） |
| F15 Git 仓库管理 | ✅ | `internal/gitcmd`（系统 git CLI，argv 传参 + 超时 + quotepath=false）、`app/repo.go`（检测/状态/暂存/提交/历史/标签/推送/初始化，离帧执行）；入口：「仓库」页签 | gitcmd 集成测试 5 件、`TestProjectRepoTab`（检测→暂存→提交→打标签全链路） |

实现备注：
- 命令面板的 Enter 经输入框 `OnSubmit`、方向键经面板 `OnShortcut`——`Context.Shortcut` 会被焦点输入框先消费，这是 mygo 的按键路由约定。
- 备份拒绝覆盖已存在文件（`VACUUM INTO` 语义），UI 侧文件名带时间戳天然避让。
- 附件查看器按扩展名路由：图片走 `ui.DecodeImage` 原生渲染，xlsx 走 sheet 矩阵 + `ui.Table`，md 走 `markdown.View`，其余走提取文本阅读视图；提取失败一律回退「文件卡 + 在系统中打开」，不阻断。
- PDF 文本提取用 `github.com/ledongthuc/pdf`（纯 Go，与 `CGO_ENABLED=0` 兼容）；加密/扫描件提取失败属预期。
- 后续候选 F5–F12 未动，见 §2。

### §6 追加（F13，2026-11-19 实测）

- [x] `TestDocxText`（段落/tab）、`TestPptxTextKeepsPageOrder`（页序与页标）、`TestXlsxSheets`（sharedStrings/inlineStr/关系映射）、`TestPDFText`（自建最小 PDF 逐字节构造 xref）、`TestKindFor`
- [x] `TestAttachmentColumnsExist`、`TestSaveFileCopiesAndCleansUp`、`TestSaveFileWithoutFilesDirFails`
- [x] `TestSaveDocumentFileUploadsAndExtracts`（上传/大小/提取/同名去重「(2)」/删除清树）、`TestSaveDocumentFileRejectsOversized`
- [x] `TestFileDocumentViewer`（表格徽标 + 查看器文本 + 在系统中打开按钮）；截图 `testdata/screens/{project-files,file-viewer}.png`
- [x] `go build` / `go vet` / `go test ./...` / `CGO_ENABLED=0 go test ./...` 全绿

### §6 追加（F14/F15，2026-11-19 实测）

- [x] gitcmd：`TestStatusBranchAndFiles`（分支 + 未跟踪/修改识别）、`TestStageCommitLog`（暂存标记/空消息拒绝/提交哈希/历史顺序与作者）、`TestTagsLifecycle`（空名拒绝/注释标签/元数据）、`TestPushWithoutRemoteFails`、`TestIsRepo`（含 worktree `.git` 文件）；git 缺失时整组跳过
- [x] UI：`TestProjectFilesTab`（根列表 → 进入 internal → 上一级回根）、`TestProjectRepoTab`（仓库检测 → 面板列未跟踪文件 → 面板内暂存 → 提交后历史 1→2 → 创建标签 v0.2.0 出现在列表）
- [x] 截图 `testdata/screens/{files-root,files-sub,repo,repo-after}.png`
- [x] 全套命令再次全绿

实现备注（F14/F15）：
- 项目详情主列改为 `文档 / 文件 / 仓库` 三个 `ui.Tabs`；切换项目时视图复位到文档页。
- git 命令一律 argv 传参 + `core.quotepath=false`（保中文文件名）；**goroutine 只取数，所有 `a.repo` 写入都在 UI 线程**，代际计数防陈旧。
- 测试教训：断言用的项目对象要在 `reload()` 后重新按 ID 取（旧对象 Path 为空会让 git 落到进程当前目录）。
