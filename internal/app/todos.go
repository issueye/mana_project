package app

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/egoist/mygo/ui"

	"mana/internal/model"
	"mana/internal/service"
)

// projectTodos filters the snapshot's todos for one project.
func (a *App) projectTodos(projectID int64) []model.Todo {
	out := make([]model.Todo, 0, 8)
	for _, t := range a.todos {
		if t.ProjectID == projectID {
			out = append(out, t)
		}
	}
	return out
}

// undoneTodos counts a project's open items.
func (a *App) undoneTodos(projectID int64) int {
	n := 0
	for _, t := range a.todos {
		if t.ProjectID == projectID && !t.Done {
			n++
		}
	}
	return n
}

// reloadSnapshot refreshes the in-memory view after a todo write.
func (a *App) reloadSnapshot() { a.reload() }

// addTodo creates an item from the input box.
func (a *App) addTodo(c *ui.Context, p model.Project) {
	content := strings.TrimSpace(a.todoInput)
	if content == "" {
		c.Toast("请输入待办内容")
		return
	}
	if _, err := a.svc.CreateTodo(context.Background(), p.ID, content); err != nil {
		c.Toast(service.FormatError(err))
		return
	}
	a.todoInput = ""
	a.todoFilter = 1 // show the new open item right away
	a.reloadSnapshot()
}

// toggleTodo flips an item's done flag.
func (a *App) toggleTodo(p model.Project, t model.Todo) {
	if _, err := a.svc.SetTodoDone(context.Background(), t.ID, !t.Done); err != nil {
		a.notice = service.FormatError(err)
		return
	}
	a.reloadSnapshot()
}

// deleteTodo removes an item.
func (a *App) deleteTodo(p model.Project, t model.Todo) {
	if err := a.svc.DeleteTodo(context.Background(), t.ID); err != nil {
		a.notice = service.FormatError(err)
		return
	}
	a.reloadSnapshot()
}

// clearDoneTodos removes every finished item of the project.
func (a *App) clearDoneTodos(c *ui.Context, p model.Project) {
	n, err := a.svc.ClearDoneTodos(context.Background(), p.ID)
	if err != nil {
		c.Toast(service.FormatError(err))
		return
	}
	if n == 0 {
		c.Toast("没有已完成的待办")
		return
	}
	a.todoFilter = 0
	a.reloadSnapshot()
	c.Toast(fmt.Sprintf("已清理 %d 项", n))
}

// todosView is the 待办 tab: an add box, a filter row and the checklist.
func (a *App) todosView(c *ui.Context, p model.Project) {
	todos := a.projectTodos(p.ID)
	undone := 0
	done := 0
	for _, t := range todos {
		if t.Done {
			done++
		} else {
			undone++
		}
	}
	ui.Column(c).Grow(1).Padding(12, 24, 0, 24).Children(func() {
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.TextInput(c, &a.todoInput).Grow(1).Height(30).
				Placeholder("添加一条待办，回车或点「添加」…").
				OnSubmit(func() { a.addTodo(c, p) })
			a.primaryButton(c, icons.plus, "添加", func() { a.addTodo(c, p) })
		})
		ui.Row(c).Height(34).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Segmented(c, &a.todoFilter, "全部", "未完成", "已完成").Height(26)
			ui.Text(c, fmt.Sprintf("%d 项未完成 · %d 项已完成", undone, done)).
				FontSize(10).TextColor(colInk3)
			ui.Spacer(c)
			if done > 0 {
				a.secondaryButton(c, nil, fmt.Sprintf("清理已完成 (%d)", done), func() { a.clearDoneTodos(c, p) })
			}
		})
		hairline(c)

		filtered := make([]model.Todo, 0, len(todos))
		for _, t := range todos {
			switch a.todoFilter {
			case 1:
				if t.Done {
					continue
				}
			case 2:
				if !t.Done {
					continue
				}
			}
			filtered = append(filtered, t)
		}
		if len(todos) == 0 {
			ui.Box(c).Grow(1).Children(func() {
				a.emptyState(c, "还没有待办", "把下一步要做的事记在这里，完成一项勾一项。",
					[2]any{"添加待办", func() {
						a.todoInput = "第一步："
					}})
			})
			return
		}
		if len(filtered) == 0 {
			ui.Box(c).Grow(1).Children(func() {
				a.emptyState(c, "这个筛选下没有待办", "换一个筛选，或添加一条新的待办。")
			})
			return
		}
		ui.Scroll(c).Grow(1).Children(func() {
			ui.Column(c).Gap(0).Children(func() {
				for i, t := range filtered {
					item := t
					idx := i
					a.todoRow(c, p, item, idx)
				}
			})
		})
		ui.Box(c).Height(12)
	})
}

// todoRow draws one checklist line: the check tile toggles, the trash
// deletes.
func (a *App) todoRow(c *ui.Context, p model.Project, t model.Todo, idx int) {
	row := ui.Row(c).MinHeight(38).Padding(0, 10).Gap(10).AlignItems(ui.Center).
		Focusable().Role(ui.RoleButton).Label((map[bool]string{true: "完成：", false: "未完成："}[t.Done]) + t.Content)
	switch {
	case row.Hovered():
		row.Background(colSheet2)
	case idx%2 == 1:
		row.Background(ui.Hex("#F8F4EA"))
	}
	row.OnClick(func() { a.toggleTodo(p, t) })
	row.Children(func() {
		check := ui.Box(c).Size(16, 16).Radius(2).Center()
		if t.Done {
			check.Background(colSeal)
		} else {
			check.Border(1, colLine).Background(colSheet)
		}
		check.Children(func() {
			if t.Done {
				ui.Icon(c, icons.check).FontSize(11).TextColor(colSheet)
			}
		})
		text := ui.Text(c, t.Content).FontSize(12)
		if t.Done {
			text.Strikethrough().TextColor(colInk4)
		} else {
			text.TextColor(colInk)
		}
		text.SingleLine()
		ui.Spacer(c)
		monoText(c, model.Relative(t.UpdatedAt), 9.5).TextColor(colInk4)
		del := ui.Box(c).Size(24, 24).Radius(2).Center().
			Focusable().Role(ui.RoleButton).Label("删除待办 " + t.Content).Tooltip("删除")
		if del.Hovered() {
			del.Background(colSealSoft)
		}
		del.OnClick(func() { a.deleteTodo(p, t) })
		del.Children(func() {
			ui.Icon(c, icons.trash).FontSize(13).TextColor(colInk4)
		})
	})
}

// noteSection is the editable 备注 block of the record card.
func (a *App) noteSection(c *ui.Context, p model.Project) {
	ui.Column(c).Padding(6, 0, 8, 0).Gap(6).Children(func() {
		ui.Row(c).Height(20).AlignItems(ui.Center).Children(func() {
			microLabel(c, "备注 · NOTE")
			ui.Spacer(c)
			if a.noteEditing {
				save := ui.Row(c).Height(20).Padding(0, 8).Radius(2).
					Background(colSeal).AlignItems(ui.Center).
					Focusable().Role(ui.RoleButton).Label("保存备注")
				save.OnClick(func() { a.saveNote(c, p) })
				save.Children(func() {
					ui.Text(c, "保存").FontSize(9.5).TextColor(colSheet)
				})
				ui.Box(c).Width(6)
				cancel := ui.Row(c).Height(20).Padding(0, 8).Radius(2).AlignItems(ui.Center).
					Focusable().Role(ui.RoleButton).Label("取消编辑备注")
				if cancel.Hovered() {
					cancel.Background(colSheet)
				}
				cancel.OnClick(func() { a.noteEditing = false })
				cancel.Children(func() {
					ui.Text(c, "取消").FontSize(9.5).TextColor(colInk3)
				})
			} else {
				edit := ui.Row(c).Height(20).Padding(0, 8).Radius(2).Gap(4).AlignItems(ui.Center).
					Focusable().Role(ui.RoleButton).Label("编辑备注")
				if edit.Hovered() {
					edit.Background(colSheet)
				}
				edit.OnClick(func() {
					a.noteDraft = p.Description
					a.noteEditing = true
				})
				edit.Children(func() {
					ui.Icon(c, icons.edit).FontSize(10).TextColor(colInk4)
					ui.Text(c, "编辑").FontSize(9.5).TextColor(colInk3)
				})
			}
		})
		if a.noteEditing {
			ui.TextArea(c, &a.noteDraft).Height(96).FontSize(11)
			if n := utf8.RuneCountInString(a.noteDraft); n > 900 {
				monoText(c, fmt.Sprintf("%d / 1000", n), 9).
					TextColor(map[bool]ui.Color{true: colSeal, false: colInk4}[n > 1000])
			}
		} else {
			note := p.Description
			if note == "" {
				note = "还没有备注；点「编辑」随手记一条。"
				ui.Text(c, note).FontSize(10.5).TextColor(colInk4)
				return
			}
			ui.Text(c, note).FontSize(10.5).TextColor(colInk2)
		}
	})
}

// saveNote writes the draft back through UpdateProject, keeping every
// other field as it is.
func (a *App) saveNote(c *ui.Context, p model.Project) {
	draft := strings.TrimSpace(a.noteDraft)
	if n := utf8.RuneCountInString(draft); n > 1000 {
		c.Toast("备注最多 1000 个字符")
		return
	}
	in := model.ProjectInput{
		Name: p.Name, GroupID: p.GroupID, Tags: p.Tags,
		Description: draft, Path: p.Path, Version: p.Version,
	}
	if _, err := a.svc.UpdateProject(context.Background(), p.ID, in); err != nil {
		c.Toast(service.FormatError(err))
		return
	}
	a.noteEditing = false
	a.reloadSnapshot()
	if cur, ok := a.projectByID(p.ID); ok {
		a.currentProject = cur
	}
	c.Toast("备注已保存")
}
