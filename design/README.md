# Mana 界面设计说明

> 关联需求：`docs/requirements.md` ｜ 可交互原型：`design/mana-prototype.html` ｜ 截图：`design/screens/`

## 1. 设计概要

| 项 | 规格 |
|---|---|
| 产品 | Mana · 开发项目管理系统（桌面端） |
| 设计体系 | 紫罗兰 · 企业紧凑风（沿用 Ant Design 5 的信息层级，重设品牌色与密度） |
| 画板 | 1440 × 900（web-1440），内容区 1440 × 838 |
| 窗口 | 默认 1440 × 900，最小 1120 × 720；单窗口、左侧导航 + 顶部栏 |
| 主题 | 浅色、白容器、冷灰工作区底；主色 #7C5CFF 只给关键操作与当前选中 |
| 字体 | Inter（latin）+ Noto Sans SC（hans），Commercial 可用；等宽 JetBrains Mono / Consolas 兜底 |
| 图标 | 全部使用 adago-design 内置 SVG 图标（24 网格、1.8 描边、currentColor），无 emoji、无字符图标 |

## 2. 页面清单

| 文件 | 页面 | 说明 |
|---|---|---|
| `screens/01-home.png` | 首页 | 问候、四项统计、最近更新项目、项目组概览、最近文档 |
| `screens/02-group-management.png` | 项目组管理 · 项目 | 左栏项目组列表，右栏组内项目表格与组文档摘要 |
| `screens/03-group-documents.png` | 项目组管理 · 文档 | 同一左右布局，右栏切换为组文档列表 |
| `screens/04-project-detail.png` | 项目详情 | 项目元信息、项目文档列表、右侧信息栏 |
| `screens/05-document-editor.png` | 文档编辑 | Markdown 源码 + 实时预览、保存状态、字数 |
| `screens/06-dialog-new-project.png` | 新建项目弹窗 | 名称、组、版本、标签、路径、备注 |
| `screens/07-dialog-new-group.png` | 新建项目组弹窗 | 名称、描述、标识色 |
| `screens/08-dialog-delete.png` | 删除项目组确认 | 级联影响说明、危险操作二次确认 |

原型支持真实点击：左侧导航、项目组切换、项目/文档页签、进入项目详情、打开文档编辑器、三类弹窗、Esc 关闭、toast。截图通过 `?screen=home|groups|groups-docs|detail|editor|project-modal|group-modal|delete-modal` 直接定位。

## 3. 布局规格

| 区域 | 尺寸 / 规则 |
|---|---|
| 左侧导航 | 228 px；白底，右边框 1 px #EEF0F2；品牌区 30 px 渐变 Logo |
| 顶部栏 | 56 px；白底，底边框 1 px；左侧标题/面包屑，中部 300 px 搜索框，右侧通知、主操作、头像 |
| 内容区 | 1440 × 810；内边距 18 / 20 / 20；白色卡片圆角 6 px，边框 #E3E5E8，无外层投影 |
| 项目组左栏 | 280 px + 14 px 间距；搜索 30 px，列表行 38 px |
| 项目组右栏 | 自适应；头部 92 px，页签 44 px，表格行 62 px，页脚 34 px |
| 项目详情 | 主列自适应 + 右侧信息栏 320 px；文档行 46 px |
| 文档编辑 | 顶部 32 px 面包屑 + 58 px 标题栏；编辑区左右 1:1；底部 30 px 状态栏 |
| 弹窗 | 项目 640 px、项目组 520 px、删除确认 448 px；圆角 6 px；遮罩 rgba(15,23,42,.42) |
| 栅格 | 间距取 4 的倍数：4 / 8 / 12 / 14 / 16 / 18 / 22 / 24 |

## 4. 设计 Token

### 4.1 颜色

| Token | 值 | 用途 |
|---|---|---|
| Primary | `#7C5CFF` | 主按钮、当前选中的组/页签、链接 |
| Primary BG | `#F3F0FF` | 选中底、图标底、提示底 |
| Text | `#1F2329` | 标题、主要正文 |
| Text Secondary | `#5C6670` | 正文、次要操作 |
| Text Tertiary | `#8F959E` | 说明、表头、时间 |
| Text Quaternary | `#B3B9C2` | 占位、辅助 |
| Border | `#E3E5E8` | 输入框、分隔 |
| Border Light | `#EEF0F2` | 卡片边框、行分隔 |
| Background | `#F4F5F7` | 工作区底 |
| Surface | `#FFFFFF` | 卡片、顶栏、侧栏 |
| Success | `#00B42A` / `#E8FFEA` | 成功、已保存、增长 |
| Warning | `#FF7D00` / `#FFF7E8` | 路径提示、风险说明 |
| Danger | `#F53F3F` / `#FFECE8` | 删除、错误 |
| Purple | `#7C5CFF` / `#F3F0FF` | 品牌强调、接口文档 |
| Cyan | `#0FC6C2` / `#E8FFFB` | 设计类文档、第二标签 |

标签颜色控制在紫/青/灰三主色 + 文档类型的橙/绿补充；同屏避免超过 3 种颜色块。

### 4.2 字阶

| 层级 | 字号 / 行高 | 字重 | 用途 |
|---|---|---|---|
| Display | 25 / 30 | 680 | 首页问候 |
| H1 | 22 / 26 | 680 | 项目详情标题 |
| H2 | 19 / 24 | 680 | 组标题 |
| Title | 14 / 20 | 650 | 卡片标题 |
| Body | 13 / 20 | 400 | 正文、表格 |
| Body Strong | 13 / 20 | 600 | 项目名、关键值 |
| Small | 11.5 / 18 | 400 | 说明、表头 |
| Caption | 10.5 / 16 | 400 | 时间、路径、脚注 |
| Mono | 11 / 20 | 400 | 路径、版本号、Markdown |

### 4.3 圆角与层级

- 控件 4 px；卡片 6 px；弹窗 6 px；标签 3 px；头像/色点圆形。
- 不使用外层投影；卡片仅 1 px 边框 + 背景分层。
- 弹窗仅 1 px 边框与半透明遮罩，不用阴影。
- 不使用重投影与玻璃拟态；层级靠边框、留白与背景明度。

## 5. 组件与 MyGo 映射

| 界面部件 | 设计规格 | MyGo 落地 |
|---|---|---|
| 左侧导航 | 228 px，36 px 行高，选中紫底 + 主色文字 | `ui.Column` + 自绘导航项（`ui.Button` 无脸变体 / `SidebarBase`），或 `ui.Sidebar` |
| 顶部搜索 | 340 px，34 px 高，Ctrl F 快捷键 | `ui.SearchField` + `OnShortcut` |
| 主操作按钮 | 30 px 高，主色，唯一 | `ui.PrimaryButton` |
| 卡片 | 6 px 圆角、1 px 边框、无投影 | `ui.Box` + 主题色（`t.Surface` / `t.Border` / `t.Radius`） |
| 项目组左右布局 | 280 px + 14 px 间距 | `ui.Row` + 固定左栏 |
| 项目表格 | 62 px 行高，表头吸顶，行点击进入详情 | `ui.Table` + `ui.TableColumn` + `ListState.Key` / `Submitted` |
| 组文档 / 项目文档 | 54–56 px 行高，标题 + 类型 + 字数 + 时间 | 同上；列宽按设计表 |
| 页签 | 40 px，下划线 2 px 主色，计数徽标 | `ui.Tabs` 或 `TabsBase` + 自绘计数 |
| 新建项目 / 组 | 640 / 520 px 弹窗，水平表单 | `ui.DialogBase` + `ui.Form` + `Field` / `TextInput` / `TextArea` / `Select` / `TokenField` |
| 删除确认 | 448 px，危险图标 + 级联说明 | `ui.AlertDialog`（`Danger`） |
| 标签 | 18 px 高、3 px 圆角 | 自绘 `Badge`/`Tag` 小组件（主题色派生浅底） |
| 文档编辑器 | 左 Markdown、右预览，底部状态 | `ui.TextArea` + `github.com/egoist/mygo/ui/markdown.View` + 状态栏 |
| toast | 底部居中 38 px，2.4 s | `c.Toast` / `c.AddToast` |
| 空状态 | 图标 + 一句说明 + 主操作 | 项目内 `emptyState` 组件 |
| 图标 | 24 网格 1.8 描边 | `ui.MustParseSVG` + `ui.Icon`，颜色 `currentColor` |

## 6. 关键视觉决策

1. **左栏同时承载「全部项目」「项目组」「未分组」**：主菜单保持两项（首页、项目组管理），全局列表不额外开页；未分组是虚拟组，视觉上用灰点、虚线分隔。
2. **项目组右栏用表格而不是卡片瀑布**：项目有路径、版本、标签、更新时间等强结构信息，表格在 1280–1440 宽度下密度更好，也直接对应 `ui.Table` 的排序、行菜单能力。
3. **组文档在「文档」页签内完整管理，同时在「项目」页签底部保留最近三条摘要**：不打断浏览项目的上下文，又能让组级知识库被看见。
4. **项目详情采用「主列文档 + 右列信息」**：写文档是主任务，元信息是辅助；路径用等宽字体并在反斜杠处换行，避免截断关键目录名。
5. **文档编辑采用源码 + 实时预览双栏**：符合开发者习惯；「编辑 / 预览」分段控件让窄窗口可切单栏；状态栏明确字数、保存状态和快捷键。
6. **深色遮罩 + 轻模糊**只用于弹窗，不使用大面积玻璃效果；危险操作必须有级联数量说明，不靠颜色本身传达风险。
7. **无装饰插画、无 emoji**：信息密度与工具感优先，空状态也使用图标与短文案。

## 7. 状态与响应式

- 选中态：浅紫底 `#F3F0FF` + 主色文字；悬浮态：`#FAFAFB` 或 `#F8F6FF`。
- 焦点态：输入框主色描边 + 3 px 10% 主色外圈；键盘可达，图标按钮均有 aria-label。
- 最小窗口 1120 × 720：项目组左栏收至 280 px，项目表格隐藏「标签」列，详情右栏收至 288 px；不出现横向滚动。
- 文案上限：项目名 80 字符、组名 60、文档标题 120；超长单行省略，悬浮显示完整值。
- 空状态：无项目、无文档、搜索无结果各有一套文案与主操作，不展示空表格。

## 8. 交付文件

```
design/
├── mana-prototype.html        # 可交互高保真原型（自包含，无外部依赖）
├── README.md                  # 本设计说明
└── screens/
    ├── 01-home.png
    ├── 02-group-management.png
    ├── 03-group-documents.png
    ├── 04-project-detail.png
    ├── 05-document-editor.png
    ├── 06-dialog-new-project.png
    ├── 07-dialog-new-group.png
    └── 08-dialog-delete.png
```

## 9. 还原说明与已知差异

- 原型用 HTML/CSS 表达视觉与交互，落地到 MyGo 原生 UI 时以本说明的 px/DIP 数值为准；颜色、间距、圆角通过主题 Token 统一，不写死在页面里。
- 原型中的浏览器字体栈以 Inter + Noto Sans SC 为主；MyGo 落地使用系统 UI 字体（Windows 为 Segoe UI / 微软雅黑），中文渲染正常，未随包分发字体文件。
- 原型中「全局搜索」「设置」「备份」为占位交互（toast），v1.0 需求中分别属于 P1/P1/P1，开发计划已标注。

## 10. 实现落地（v1.0）

实际应用位于 `internal/app`，由 MyGo 原生 UI 绘制；实现截图在 [`docs/screens/`](../docs/screens/)。与设计稿的差异：

| 项 | 设计稿 | 实现 |
|---|---|---|
| 窗口 | 1440×900 内容区 | `TitleBarHidden + HideWindowButtons` 无边框标题栏；自绘 34 px 顶栏与窗口控制，保留窗口缩放边缘和系统圆角 |
| 首页 | 固定两栏 + 最近文档 | 同布局；窗口较矮时整页可滚动，不裁切 |
| 项目组管理 | 左右 280 px | 左右两栏固定 280 px + 14 px 间距；为保证清晰度不使用 Split 分割条，可拖拽缩放暂缓 |
| 项目详情 | 右侧信息栏 320 px | <1280 px 时收至 288 px；路径在反斜杠处换行 |
| 表格 | 行高 54–62 px | 表格外层增加 12 px 横向留白、4 px 纵向留白，按钮与行内容不贴边 |
| 卡片与弹窗 | 浅阴影 + 描边 | 实际不使用外层投影，改为 1 px 边框 + 背景分层；弹窗用半透明遮罩与 1 px 边框，避免原生阴影模糊 |
| 弹窗 | 标题 + 描述 + 说明 footer | header 40 px 只留标题（主色），footer 44 px 只留取消/确认；表单改为 `ui.Form` 水平标签列，必填项标 `*`；header/footer 分割线左右拉通到弹窗边缘 |
| 文档编辑 | 源码 + 预览，无链接字段 | 增加「关联链接（可选）」输入，满足需求 FR-04 字段；源码区高度随窗口计算 |
| 组文档 | 文档页签 + 新建按钮 | 已实现“新建文档”，并支持删除 |
| 行点击 | 整行进入详情 | `ui.Table` 的 `ListState.Selected` + `Changed` 驱动导航；保留键盘选择 |
| 删除组 | 遮罩 + 影响说明 | `ui.Modal` 二次确认，文案动态显示项目/文档数量 |
| 主题 | 紫罗兰紧凑 Token | 自定义 `ui.Theme` 映射主色、文字四级、边框、圆角、间距；固定浅色 |

已知实现取舍：

- 使用 `modernc.org/sqlite`（纯 Go）替代原计划的 MyGo SQLite 原生插件：本机无法下载 GitHub 原生库；schema/SQL/事务一致。
- 全局搜索目前把关键字应用到“全部项目”页；文档标题的分组结果与设置/备份列入 P1 收尾。
- Table 的表头排序未接 `ListState.Sort`，当前提供“更新时间 / 名称”下拉排序；后续可切到表头排序。
- 未实现行级“移入项目组”菜单，编辑项目时可直接改组；独立入口列入 P1。
