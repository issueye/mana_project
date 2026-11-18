package app

import (
	"context"
	"fmt"

	"github.com/egoist/mygo/ui"

	"mana/internal/model"
	"mana/internal/service"
)

// openMoveProject asks where a project should live; the select preselects
// its current group.
func (a *App) openMoveProject(id int64) {
	p, ok := a.projectByID(id)
	if !ok {
		a.notice = "项目不存在或已被删除"
		return
	}
	a.moveProjectID = id
	a.moveChoice = "未分组"
	if p.GroupID != nil {
		if g, ok := a.groupByID(*p.GroupID); ok {
			a.moveChoice = g.Name
		}
	}
	a.fieldErr = map[string]string{}
	a.showMoveDialog = true
}

// chosenMoveID translates the select's label into a group ID; nil moves
// the project out of every group.
func (a *App) chosenMoveID() *int64 {
	if a.moveChoice == "" || a.moveChoice == "未分组" {
		return nil
	}
	if g, ok := a.groupByName(a.moveChoice); ok {
		return &g.ID
	}
	return nil
}

// submitMove applies the move and refreshes everything that shows counts.
func (a *App) submitMove(c *ui.Context) {
	if a.svc == nil {
		return
	}
	if err := a.svc.MoveProject(context.Background(), a.moveProjectID, a.chosenMoveID()); err != nil {
		c.Toast(service.FormatError(err))
		return
	}
	dest := a.moveChoice
	if dest == "" {
		dest = "未分组"
	}
	a.showMoveDialog = false
	a.reload()
	if p, ok := a.projectByID(a.moveProjectID); ok {
		a.currentProject = p
	}
	c.Toast(fmt.Sprintf("已移动到 %s", dest))
}

// moveDialog is the small sheet that re-homes a project.
func (a *App) moveDialog(c *ui.Context) {
	ui.DialogBase(c, &a.showMoveDialog, func(backdrop, panel ui.Element) {
		p, ok := a.projectByID(a.moveProjectID)
		name := "项目"
		if ok {
			name = p.Name
		}
		a.dialogShell(c, backdrop, panel, 480,
			func() {
				a.dialogTop(c, "MOVE RECORD · 项目", "移动 “"+name+"”",
					"项目的路径、文档与标签保持不变，只更换所属项目组。",
					func() { a.showMoveDialog = false })
			},
			func() {
				ui.Form(c, func() {
					current := itoa(len(a.projectsInGroup(a.chosenMoveID())))
					ui.Field(c, "目标项目组", func() {
						options := []string{"未分组"}
						for _, g := range a.groups {
							options = append(options, g.Name)
						}
						ui.Select(c, &a.moveChoice, options)
					}).Error(a.fieldErr["group"])
					ui.Row(c).Gap(5).AlignItems(ui.Center).Children(func() {
						ui.Icon(c, icons.info).FontSize(11).TextColor(colInk4)
						ui.Text(c, "目标组现有 " + current + " 个项目。").FontSize(9.5).TextColor(colInk4)
					})
				}).Padding(4, 20, 14, 20)
				a.dialogFoot(c, "CTRL ENTER 确认 · ESC 取消", "取消", "移动项目", false, func() { a.submitMove(c) })
			})
	})
}

// projectsInGroup counts the projects currently in a group (nil = none).
func (a *App) projectsInGroup(groupID *int64) []model.Project {
	out := make([]model.Project, 0)
	for _, p := range a.projects {
		if groupID == nil {
			if p.GroupID == nil {
				out = append(out, p)
			}
			continue
		}
		if p.GroupID != nil && *p.GroupID == *groupID {
			out = append(out, p)
		}
	}
	return out
}
