package app

import (
	"context"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"

	"mana/internal/gitcmd"
	"mana/internal/model"
)

// gitQuick bounds the local git queries; pushes rely on gitcmd's own
// longer timeout.
const gitQuick = 30 * time.Second

// repoState is the Git panel of one project. The git commands run on
// goroutines that only produce data; every write to this struct happens
// on the UI thread (win.Update, or inline in headless tests).
type repoState struct {
	projectID int64
	loading   bool
	isRepo    bool
	checked   bool
	hasRemote bool

	branch   string
	files    []gitcmd.FileStatus
	commits  []gitcmd.LogEntry
	tags     []gitcmd.Tag
	errText  string
	dirtyMsg string // last successful action, until the next one
	gen      int

	// Form state that must survive frames.
	commitMsg  string
	tagName    string
	tagMessage string
}

// repoData is what one fetch produces, merged into repoState on the UI
// thread.
type repoData struct {
	isRepo    bool
	hasRemote bool
	branch    string
	files     []gitcmd.FileStatus
	commits   []gitcmd.LogEntry
	tags      []gitcmd.Tag
	errText   string
}

// openRepoView resets the panel for a project and starts the fetch.
func (a *App) openRepoView(p model.Project) {
	a.repo = repoState{
		projectID: p.ID, loading: true, gen: a.repo.gen + 1,
		commitMsg: a.repo.commitMsg, tagName: a.repo.tagName, tagMessage: a.repo.tagMessage,
	}
	id := a.repo.gen
	if a.win == nil {
		a.repoMerge(id, fetchRepoData(p))
		return
	}
	go func() {
		data := fetchRepoData(p)
		a.win.Update(func() { a.repoMerge(id, data) })
	}()
}

// repoMerge folds a finished fetch into the panel state.
func (a *App) repoMerge(id int, data repoData) {
	if id != a.repo.gen {
		return
	}
	r := a.repo
	r.loading = false
	r.checked = true
	r.isRepo = data.isRepo
	r.hasRemote = data.hasRemote
	r.branch = data.branch
	r.files = data.files
	r.commits = data.commits
	r.tags = data.tags
	if data.errText != "" {
		r.errText = data.errText
	}
	a.repo = r
}

// fetchRepoData runs the git queries. It touches nothing shared, so it is
// safe on a goroutine.
func fetchRepoData(p model.Project) repoData {
	if p.Path == "" {
		return repoData{errText: "该项目未登记路径，请先在「编辑项目」中登记"}
	}
	if !gitcmd.IsRepo(p.Path) {
		return repoData{}
	}
	ctx, cancel := context.WithTimeout(context.Background(), gitQuick)
	defer cancel()
	data := repoData{isRepo: true}
	branch, files, err := gitcmd.Status(ctx, p.Path)
	if err != nil {
		data.errText = err.Error()
		return data
	}
	data.branch, data.files = branch, files
	data.hasRemote = gitcmd.HasRemote(p.Path)
	data.commits, _ = gitcmd.Log(ctx, p.Path, 50)
	data.tags, _ = gitcmd.Tags(ctx, p.Path)
	return data
}

// runRepoOp runs one mutating git command off the frame and refreshes.
func (a *App) runRepoOp(p model.Project, label string, op func(context.Context) error) {
	a.repo.gen++
	id := a.repo.gen
	apply := func(err error) {
		r := a.repo
		if id != r.gen {
			return
		}
		if err != nil {
			r.errText = err.Error()
			a.repo = r
			return
		}
		r.dirtyMsg = label
		r.errText = ""
		a.repo = r
		a.openRepoView(p)
	}
	if a.win == nil {
		ctx, cancel := context.WithTimeout(context.Background(), gitQuick)
		defer cancel()
		apply(op(ctx))
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), gitQuick)
		defer cancel()
		err := op(ctx)
		a.win.Update(func() { apply(err) })
	}()
}

// stageRepoFiles adds (or unadds) files, then refreshes the status.
func (a *App) stageRepoFiles(p model.Project, paths []string, stage bool) {
	verb := "已暂存"
	if stage {
		verb = "已暂存所选改动"
	} else {
		verb = "已取消暂存"
	}
	a.runRepoOp(p, verb, func(ctx context.Context) error {
		if stage {
			return gitcmd.Stage(ctx, p.Path, paths)
		}
		return gitcmd.Unstage(ctx, p.Path, paths)
	})
}

// commitRepo records the staged changes.
func (a *App) commitRepo(p model.Project) {
	msg := strings.TrimSpace(a.repo.commitMsg)
	if msg == "" {
		r := a.repo
		r.errText = "请填写提交说明"
		a.repo = r
		return
	}
	a.runRepoOp(p, "提交完成", func(ctx context.Context) error {
		hash, err := gitcmd.Commit(ctx, p.Path, msg)
		if err == nil && hash != "" {
			a.winNotify("已提交 " + hash)
		}
		return err
	})
}

// createRepoTag makes an annotated tag.
func (a *App) createRepoTag(p model.Project) {
	name := strings.TrimSpace(a.repo.tagName)
	if name == "" {
		r := a.repo
		r.errText = "请填写标签名称"
		a.repo = r
		return
	}
	a.runRepoOp(p, "已创建标签 "+name, func(ctx context.Context) error {
		return gitcmd.CreateTag(ctx, p.Path, name, strings.TrimSpace(a.repo.tagMessage))
	})
}

// pushRepo sends the branch (and a chosen tag) to origin.
func (a *App) pushRepo(c *ui.Context, p model.Project, tag string) {
	a.repo.gen++
	id := a.repo.gen
	apply := func(err error) {
		r := a.repo
		if id != r.gen {
			return
		}
		if err != nil {
			r.errText = err.Error()
			a.repo = r
			return
		}
		r.dirtyMsg = "已推送到 origin"
		r.errText = ""
		a.repo = r
		a.openRepoView(p)
	}
	if a.win == nil {
		apply(gitcmd.Push(context.Background(), p.Path, tag))
		return
	}
	go func() {
		apply(gitcmd.Push(context.Background(), p.Path, tag))
	}()
}

// winNotify queues a toast when a background op has something to say.
func (a *App) winNotify(msg string) {
	if a.win == nil {
		return
	}
	a.win.Update(func() { a.notice = msg })
}

// initRepo creates a repository in the project path.
func (a *App) initRepo(p model.Project) {
	a.repo.gen++
	id := a.repo.gen
	apply := func(err error) {
		r := a.repo
		if id != r.gen {
			return
		}
		if err != nil {
			r.errText = err.Error()
			a.repo = r
			return
		}
		r.dirtyMsg = "已初始化 Git 仓库"
		a.repo = r
		a.openRepoView(p)
	}
	if a.win == nil {
		apply(gitcmd.Init(p.Path))
		return
	}
	go func() { apply(gitcmd.Init(p.Path)) }()
}

// refreshRepo re-runs the snapshot with a new generation.
func (a *App) refreshRepo(p model.Project) {
	a.openRepoView(p)
}

// repoView is the 仓库 tab.
func (a *App) repoView(c *ui.Context, p model.Project) {
	r := a.repo
	if r.projectID != p.ID {
		a.openRepoView(p)
		r = a.repo
	}
	ui.Column(c).Grow(1).Padding(12, 24, 0, 24).Children(func() {
		ui.Row(c).Height(34).Gap(8).AlignItems(ui.Center).Children(func() {
			if r.isRepo && r.branch != "" {
				ui.Row(c).Height(24).Padding(0, 9).Radius(2).Gap(6).AlignItems(ui.Center).
					Border(1, colSeal).Children(func() {
					monoText(c, r.branch, 10).FontWeight(600).TextColor(colSeal)
				})
			}
			if text := repoRemoteText(r); text != "" {
				ui.Text(c, text).FontSize(10).TextColor(colInk3)
			}
			ui.Spacer(c)
			if r.dirtyMsg != "" {
				ui.Text(c, r.dirtyMsg).FontSize(10).FontWeight(600).TextColor(colMoss)
				ui.Box(c).Width(8)
			}
			a.ghostButton(c, icons.refresh, "刷新", func() { a.refreshRepo(p) })
		})
		hairline(c)

		switch {
		case r.loading && !r.checked:
			ui.Box(c).Grow(1).Children(func() {
				a.emptyState(c, "正在检测 Git 仓库…", p.Path)
			})
		case r.errText != "" && !r.isRepo:
			ui.Box(c).Grow(1).Children(func() {
				a.emptyState(c, "无法读取仓库", r.errText)
			})
		case !r.isRepo:
			ui.Box(c).Grow(1).Children(func() {
				a.emptyState(c, "还没有检测到 Git 仓库",
					"在项目路径下初始化一个仓库，或确认登记的路径是否正确。",
					[2]any{"初始化仓库", func() { a.initRepo(p) }})
			})
		default:
			a.repoBody(c, p, r)
		}
		ui.Box(c).Height(12)
	})
}

// repoBody is the working panel of a detected repository.
func (a *App) repoBody(c *ui.Context, p model.Project, r repoState) {
	ui.Scroll(c).Grow(1).Children(func() {
		ui.Column(c).FillWidth().Gap(10).Children(func() {
			a.repoStatusSection(c, p, r)
			a.repoCommitSection(c, p, r)
			a.repoLogSection(c, r)
			a.repoTagSection(c, p, r)
		})
	})
}

// repoStatusSection lists the changed files and the staging controls.
func (a *App) repoStatusSection(c *ui.Context, p model.Project, r repoState) {
	ui.Row(c).Height(28).AlignItems(ui.Center).Children(func() {
		sectionTitle(c, "改动")
		ui.Box(c).Width(6)
		monoText(c, itoaPad(len(r.files)), 9.5).TextColor(colInk4)
		ui.Spacer(c)
		if len(r.files) > 0 {
			a.secondaryButton(c, nil, "全部暂存", func() { a.stageRepoFiles(p, nil, true) })
		}
	})
	if len(r.files) == 0 {
		ui.Row(c).Height(30).Children(func() {
			ui.Text(c, "工作区是干净的，没有待提交的改动。").FontSize(10.5).TextColor(colInk3)
		})
		return
	}
	for i, f := range r.files {
		file := f
		row := ui.Row(c).MinHeight(32).Padding(0, 8).Gap(9).AlignItems(ui.Center).
			Focusable().Role(ui.RoleButton).Label(file.Path)
		if i%2 == 1 {
			row.Background(colPaper)
		}
		row.Children(func() {
			monoText(c, file.Code, 10).FontWeight(600).TextColor(statusColor(file.Code))
			ui.Text(c, file.Path).FontSize(11).TextColor(colInk2).SingleLine()
			ui.Spacer(c)
			label := "暂存"
			if file.Staged {
				label = "取消暂存"
			}
			b := ui.Row(c).Height(22).Padding(0, 8).Radius(2).AlignItems(ui.Center).
				Border(1, colLine).
				Focusable().Role(ui.RoleButton).Label(label + " " + file.Path)
			if b.Hovered() {
				b.Background(colSheet2)
			}
			b.OnClick(func() {
				a.stageRepoFiles(p, []string{file.Path}, !file.Staged)
			})
			b.Children(func() {
				ui.Text(c, label).FontSize(9.5).TextColor(colInk2)
			})
		})
	}
}

// statusColor maps porcelain letters to the index palette.
func statusColor(code string) ui.Color {
	switch {
	case strings.Contains(code, "D"):
		return colSeal
	case code == "??":
		return colMoss
	case len(code) > 0 && code[0] != ' ':
		return colAmber
	default:
		return colInk3
	}
}

// repoCommitSection is the message box and the commit button.
func (a *App) repoCommitSection(c *ui.Context, p model.Project, r repoState) {
	ui.Column(c).Gap(4).Padding(0, 0, 4, 0).Children(func() {
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.TextInput(c, &a.repo.commitMsg).Grow(1).Height(30).
				Placeholder("提交说明（提交已暂存的改动）")
			a.primaryButton(c, icons.check, "提交", func() { a.commitRepo(p) })
		})
		if r.errText != "" {
			ui.Row(c).Gap(5).AlignItems(ui.Center).Children(func() {
				ui.Icon(c, icons.warning).FontSize(11).TextColor(colSeal)
				ui.Text(c, r.errText).FontSize(9.5).TextColor(colSeal).SingleLine()
			})
		}
	})
}

// repoLogSection is the recent history.
func (a *App) repoLogSection(c *ui.Context, r repoState) {
	ui.Row(c).Height(28).AlignItems(ui.Center).Children(func() {
		sectionTitle(c, "历史")
		ui.Box(c).Width(6)
		monoText(c, itoaPad(len(r.commits)), 9.5).TextColor(colInk4)
		ui.Spacer(c)
		monoText(c, "最近 50 条", 9).TextColor(colInk4)
	})
	if len(r.commits) == 0 {
		ui.Row(c).Height(30).Children(func() {
			ui.Text(c, "还没有提交；写好提交说明后点「提交」。").FontSize(10.5).TextColor(colInk3)
		})
		return
	}
	for i, cm := range r.commits {
		entry := cm
		row := ui.Row(c).MinHeight(30).Padding(0, 8).Gap(10).AlignItems(ui.Center)
		if i%2 == 1 {
			row.Background(colPaper)
		}
		row.Children(func() {
			monoText(c, entry.Hash, 10).FontWeight(600).TextColor(colSeal)
			ui.Text(c, entry.Subject).FontSize(11).TextColor(colInk).SingleLine()
			ui.Spacer(c)
			ui.Text(c, entry.Author).FontSize(9.5).TextColor(colInk3)
			ui.Box(c).Width(10)
			ui.Text(c, model.Relative(entry.Time)).FontSize(9.5).TextColor(colInk4)
		})
	}
}

// repoTagSection is the release area: existing tags and the create form.
func (a *App) repoTagSection(c *ui.Context, p model.Project, r repoState) {
	ui.Row(c).Height(28).AlignItems(ui.Center).Children(func() {
		sectionTitle(c, "标签 / 发版")
		ui.Box(c).Width(6)
		monoText(c, itoaPad(len(r.tags)), 9.5).TextColor(colInk4)
		ui.Spacer(c)
		if r.hasRemote {
			a.secondaryButton(c, icons.arrowRight, "推送分支与标签", func() { a.pushRepo(c, p, a.repo.tagName) })
		} else {
			ui.Text(c, "未配置远端 origin").FontSize(9.5).TextColor(colInk4)
		}
	})
	ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
		ui.TextInput(c, &a.repo.tagName).Width(180).Height(28).Placeholder("标签名，如 v1.2.0")
		ui.TextInput(c, &a.repo.tagMessage).Grow(1).Height(28).Placeholder("发版说明（可选）")
		a.secondaryButton(c, icons.bookmark, "创建标签", func() { a.createRepoTag(p) })
	})
	if len(r.tags) == 0 {
		ui.Row(c).Height(26).Children(func() {
			ui.Text(c, "还没有标签；打一个标签即是一次发版记录。").FontSize(10).TextColor(colInk4)
		})
		return
	}
	for i, tg := range r.tags {
		tag := tg
		row := ui.Row(c).MinHeight(28).Padding(0, 8).Gap(10).AlignItems(ui.Center)
		if i%2 == 1 {
			row.Background(colPaper)
		}
		row.Children(func() {
			monoText(c, tag.Name, 10.5).FontWeight(600).TextColor(colInk)
			if tag.Subject != "" {
				ui.Text(c, tag.Subject).FontSize(10).TextColor(colInk2).SingleLine()
			}
			ui.Spacer(c)
			monoText(c, shortHash(tag.Hash), 9.5).TextColor(colInk4)
			ui.Box(c).Width(10)
			ui.Text(c, model.Relative(tag.Created)).FontSize(9.5).TextColor(colInk4)
		})
	}
}

// shortHash trims a commit hash for display.
func shortHash(hash string) string {
	if len(hash) > 7 {
		return hash[:7]
	}
	return hash
}

// repoRemoteText describes the remote configuration in one line.
func repoRemoteText(r repoState) string {
	if !r.checked {
		return ""
	}
	if r.isRepo && r.hasRemote {
		return "远端 origin 已配置"
	}
	if r.isRepo {
		return "本地仓库 · 未配置远端"
	}
	return ""
}
