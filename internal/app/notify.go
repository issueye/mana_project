package app

import (
	"context"
	"fmt"
	"time"

	"github.com/egoist/mygo/ui"

	"mana/internal/gitcmd"
	"mana/internal/model"
)

// The remote-update watcher (F19): a background ticker fetches every
// git project with an origin remote and turns "behind" counts into
// top-right notification banners.
const (
	checkFirstDelay = 20 * time.Second
	checkInterval   = 5 * time.Minute
	checkQuick      = 90 * time.Second
)

// updateNotice is one banner: a project's remote moved ahead.
type updateNotice struct {
	key       string // projectID:behind, deduplicates repeats
	projectID int64
	name      string
	branch    string
	behind    int
	at        time.Time
}

// remoteTarget is what a check cycle needs per project, collected on the
// UI thread so the workers never touch shared state.
type remoteTarget struct {
	id   int64
	name string
	path string
}

// remoteResult is what one worker reports back.
type remoteResult struct {
	target    remoteTarget
	isRepo    bool
	hasRemote bool
	behind    int // -1 when unknown (no upstream, fetch failed…)
	errText   string
}

// startUpdateWatcher launches the periodic check loop. It is a no-op
// without a window (headless tests call the pieces directly).
func (a *App) startUpdateWatcher() {
	if a.win == nil {
		return
	}
	go func() {
		time.Sleep(checkFirstDelay)
		for {
			a.scheduleRemoteCheck()
			time.Sleep(checkInterval)
		}
	}()
}

// scheduleRemoteCheck collects the check targets on the UI thread, then
// lets one worker per project hit the network.
func (a *App) scheduleRemoteCheck() {
	a.win.Update(func() {
		if a.checkRunning {
			return
		}
		a.checkRunning = true
		targets := make([]remoteTarget, 0, len(a.projects))
		for _, p := range a.projects {
			if p.Path == "" {
				continue
			}
			targets = append(targets, remoteTarget{id: p.ID, name: p.Name, path: p.Path})
		}
		go func() {
			results := make([]remoteResult, 0, len(targets))
			for _, target := range targets {
				results = append(results, a.checkOneRemote(target))
			}
			a.win.Update(func() {
				a.checkRunning = false
				for _, res := range results {
					a.applyRemoteResult(res)
				}
			})
		}()
	})
}

// checkOneRemote fetches and measures one project. It touches nothing
// shared; the results merge on the UI thread.
func (a *App) checkOneRemote(target remoteTarget) remoteResult {
	res := remoteResult{target: target, behind: -1}
	if !gitcmd.IsRepo(target.path) {
		return res
	}
	res.isRepo = true
	ctx, cancel := context.WithTimeout(context.Background(), checkQuick)
	defer cancel()
	if err := gitcmd.Fetch(ctx, target.path); err != nil {
		res.errText = err.Error()
		return res
	}
	res.hasRemote = true
	branch, _, err := gitcmd.Status(ctx, target.path)
	if err != nil || branch == "" {
		return res
	}
	if ahead, behind, err := gitcmd.AheadBehind(ctx, target.path, branch); err == nil {
		res.behind = behind
		_ = ahead
	}
	return res
}

// applyRemoteResult folds one worker's outcome into the notifications.
func (a *App) applyRemoteResult(res remoteResult) {
	a.knownRepos[res.target.id] = res.isRepo && res.hasRemote
	if a.repo.projectID == res.target.id && res.isRepo {
		r := a.repo
		r.remoteBehind = res.behind
		if res.errText != "" {
			r.remoteBehind = -1
		}
		a.repo = r
	}
	if res.behind <= 0 {
		// Everything seen: drop this project's stale banners.
		a.removeNoticesFor(res.target.id)
		return
	}
	if res.errText != "" {
		return
	}
	key := fmt.Sprintf("%d:%d", res.target.id, res.behind)
	for _, n := range a.notices {
		if n.projectID == res.target.id {
			if n.key == key {
				return // already telling the user about this
			}
			a.removeNoticesFor(res.target.id) // stale count, re-announce below
			break
		}
	}
	if a.seenNotices[key] {
		return // the user already looked at this exact state
	}
	a.notices = append(a.notices, updateNotice{
		key: key, projectID: res.target.id, name: res.target.name,
		branch: a.projectBranch(res.target.id), behind: res.behind, at: time.Now(),
	})
}

// projectBranch looks up the cached branch of a project from the repo
// panel state, falling back to "master" wording later.
func (a *App) projectBranch(projectID int64) string {
	if p, ok := a.projectByID(projectID); ok {
		branch, _, err := gitStatusQuick(p)
		if err == nil && branch != "" {
			return branch
		}
	}
	return ""
}

// gitStatusQuick reads the branch name off the frame.
func gitStatusQuick(p model.Project) (string, []gitcmd.FileStatus, error) {
	ctx, cancel := context.WithTimeout(context.Background(), gitQuick)
	defer cancel()
	return gitcmd.Status(ctx, p.Path)
}

// removeNoticesFor drops every banner of one project.
func (a *App) removeNoticesFor(projectID int64) {
	kept := a.notices[:0]
	for _, n := range a.notices {
		if n.projectID != projectID {
			kept = append(kept, n)
		}
	}
	a.notices = kept
}

// dismissNotice closes one banner and remembers the state as seen.
func (a *App) dismissNotice(n updateNotice) {
	a.seenNotices[n.key] = true
	kept := a.notices[:0]
	for _, item := range a.notices {
		if item.key != n.key {
			kept = append(kept, item)
		}
	}
	a.notices = kept
}

// openNotice jumps to the project's repository panel and dismisses.
func (a *App) openNotice(n updateNotice) {
	a.dismissNotice(n)
	if _, ok := a.projectByID(n.projectID); !ok {
		return
	}
	a.openProject(n.projectID)
	a.projectView = 2
	a.refreshRepo(a.currentProject)
}

// checkRemoteNow is the manual check button of the repo panel: fetch and
// measure one project right away.
func (a *App) checkRemoteNow(p model.Project) {
	go func() {
		res := a.checkOneRemote(remoteTarget{id: p.ID, name: p.Name, path: p.Path})
		a.win.Update(func() {
			a.applyRemoteResult(res)
			if a.repo.projectID == p.ID {
				r := a.repo
				if res.behind >= 0 {
					r.remoteBehind = res.behind
				} else if res.errText != "" {
					r.errText = res.errText
				}
				a.repo = r
			}
			if res.behind > 0 {
				a.notice = fmt.Sprintf("远端有 %d 个新提交", res.behind)
			} else if res.errText == "" {
				a.notice = "已是最新"
			}
		})
	}()
}

// notificationOverlay draws the top-right banners above everything else.
func (a *App) notificationOverlay(c *ui.Context) {
	if len(a.notices) == 0 {
		return
	}
	ui.Overlay(c, func() {
		w, _ := c.Size()
		cardW := float32(312)
		shown := a.notices
		extra := 0
		if len(shown) > 3 {
			extra = len(shown) - 3
			shown = shown[len(shown)-3:]
		}
		for i := len(shown) - 1; i >= 0; i-- {
			n := shown[i]
			pos := len(shown) - 1 - i
			y := 46 + float32(pos)*96
			card := ui.Column(c).Width(cardW).Background(colSheet).
				Border(1, colLine).Radius(2).
				Shadow(0, 10, 28, 0, ui.RGBA(23, 21, 15, 0.30)).
				Absolute().Left(w - cardW - 16).Top(y)
			card.Children(func() {
				ui.Box(c).Height(3).FillWidth().Background(colSeal)
				ui.Column(c).Padding(9, 12, 10, 12).Gap(4).Children(func() {
					ui.Row(c).Gap(7).AlignItems(ui.Center).Children(func() {
						ui.Icon(c, icons.refresh).FontSize(13).TextColor(colSeal)
						ui.Text(c, "远程仓库有更新").FontSize(11).FontWeight(600)
						ui.Spacer(c)
						closeBtn := ui.Box(c).Size(20, 20).Radius(2).Center().
							Focusable().Role(ui.RoleButton).Label("关闭通知")
						if closeBtn.Hovered() {
							closeBtn.Background(colSheet2)
						}
						closeBtn.OnClick(func() { a.dismissNotice(n) })
						closeBtn.Children(func() {
							ui.Icon(c, icons.close).FontSize(11).TextColor(colInk3)
						})
					})
					monoText(c, n.name+"  ·  "+n.branch, 10).TextColor(colInk2).SingleLine()
					ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
						ui.Text(c, fmt.Sprintf("远端领先 %d 个提交", n.behind)).
							FontSize(10.5).FontWeight(600).TextColor(colSeal)
						ui.Spacer(c)
						view := ui.Row(c).Height(22).Padding(0, 9).Radius(2).
							Border(1, colLine).AlignItems(ui.Center).
							Focusable().Role(ui.RoleButton).Label("查看 " + n.name + " 的更新")
						if view.Hovered() {
							view.Background(colSheet2)
						}
						view.OnClick(func() { a.openNotice(n) })
						view.Children(func() {
							ui.Text(c, "查看").FontSize(10).TextColor(colInk2)
						})
						ui.Text(c, model.Relative(n.at)).FontSize(9).TextColor(colInk4)
					})
					if extra > 0 && pos == len(shown)-1 {
						ui.Text(c, fmt.Sprintf("还有 %d 条通知", extra)).FontSize(9).TextColor(colInk4)
					}
				})
			})
		}
	})
}

// repoBehindChip is the remote freshness line of the repo panel.
func (a *App) repoBehindChip(c *ui.Context, p model.Project, r repoState) {
	ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
		switch {
		case r.remoteBehind > 0:
			chip := ui.Row(c).Height(22).Padding(0, 9).Radius(2).Gap(5).AlignItems(ui.Center).
				Border(1, colSeal).Background(colSealSoft)
			chip.Children(func() {
				monoText(c, fmt.Sprintf("落后 %d 个提交", r.remoteBehind), 9.5).FontWeight(600).TextColor(colSeal)
			})
		case r.remoteBehind == 0:
			ui.Text(c, "与远端一致").FontSize(9.5).TextColor(colMoss)
		case r.remoteBehind == -1:
			ui.Text(c, "无法确定与远端的差距").FontSize(9.5).TextColor(colInk4)
		default:
			ui.Text(c, "尚未与远端比对").FontSize(9.5).TextColor(colInk4)
		}
		a.secondaryButton(c, icons.refresh, "检查远程更新", func() { a.checkRemoteNow(p) })
	})
}
