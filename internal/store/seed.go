package store

import (
	"context"
	"fmt"
	"time"

	"mana/internal/model"
)

// SeedDemo inserts the sample groups, projects and documents used by the
// design and by `mana -demo`. It does nothing when groups already exist.
func (s *Store) SeedDemo(ctx context.Context) error {
	groups, err := s.ListGroups(ctx)
	if err != nil {
		return err
	}
	if len(groups) > 0 {
		return nil
	}

	groupSpecs := []struct {
		Name, Description, Color string
		Docs                     []string
	}{
		{"客户端应用", "面向桌面与移动端的产品线，覆盖 macOS、iOS 与 Web。", "#7C5CFF",
			[]string{"客户端设计规范", "双端发布流程", "Q4 迭代目标", "移动端埋点方案", "组内技术评审记录"}},
		{"后端服务", "对内对外的 HTTP、任务与数据服务；统一部署在自建集群。", "#52C41A",
			[]string{"服务分层约定", "API 错误码规范", "发布回滚手册", "数据库变更流程"}},
		{"基础架构", "构建、发布、网关与密钥等基础设施，支撑各产品线。", "#722ED1",
			[]string{"CI 基线", "镜像与制品规范", "密钥轮换手册"}},
		{"实验项目", "探索型项目与原型验证，允许快速试错；文档归档到组级知识库。", "#FA8C16",
			[]string{"实验记录 2026-11", "沙箱使用说明"}},
		{"文档与工具", "面向内部效率的文档工具与状态页。", "#13C2C2",
			[]string{"文档模板"}},
	}

	groupIDs := map[string]int64{}
	for _, spec := range groupSpecs {
		g, err := s.CreateGroup(ctx, model.GroupInput{Name: spec.Name, Description: spec.Description, Color: spec.Color})
		if err != nil {
			return fmt.Errorf("seed group %s: %w", spec.Name, err)
		}
		groupIDs[spec.Name] = g.ID
		for i, title := range spec.Docs {
			kind := []string{"design", "other", "plan", "meeting"}[i%4]
			_, err := s.CreateDocument(ctx, model.DocumentInput{
				OwnerType: model.OwnerGroup, OwnerID: g.ID, Title: title, Kind: kind,
				Content: fmt.Sprintf("# %s\n\n%s 的组级文档，记录规范与流程。\n", title, spec.Name),
			})
			if err != nil {
				return fmt.Errorf("seed group doc %s: %w", title, err)
			}
		}
	}

	projectSpecs := []struct {
		Group       string
		Name        string
		Version     string
		Tags        []string
		Path        string
		Description string
		UpdatedAgo  time.Duration
		DocCount    int
	}{
		{"客户端应用", "Mana Desktop", "v0.9.2", []string{"Go", "MyGo", "桌面端"}, `E:\codes\go_projects\mana_project`, "本地优先的项目管理桌面端，Go + MyGo 原生 UI，SQLite 存储。", 2 * time.Hour, 5},
		{"客户端应用", "Orbit Mobile", "v2.3.1", []string{"Flutter", "iOS", "Android"}, `E:\codes\mobile\orbit`, "跨端移动应用，Flutter 渲染与原生能力桥接。", 26 * time.Hour, 3},
		{"客户端应用", "Pulse CLI", "v1.4.0", []string{"Go", "CLI"}, `E:\codes\go_projects\pulse`, "面向服务端开发者的命令行工具集。", 3 * 24 * time.Hour, 2},
		{"客户端应用", "Nova Player", "v0.6.0", []string{"Swift", "macOS"}, `E:\codes\mac\nova-player`, "macOS 原生音频播放器，实验性项目。", 5 * 24 * time.Hour, 1},
		{"客户端应用", "Ember Notes", "v1.2.3", []string{"TypeScript", "Tauri"}, `E:\codes\desktop\ember-notes`, "轻量本地笔记应用，Tauri + React。", 7 * 24 * time.Hour, 1},
		{"后端服务", "Atlas API", "v3.1.0", []string{"Go", "SQLite", "HTTP"}, `E:\codes\go_projects\atlas`, "核心业务 API，REST + JSON，本地缓存与审计日志。", 4 * time.Hour, 3},
		{"后端服务", "Relic Worker", "v1.8.2", []string{"Go", "Queue"}, `E:\codes\go_projects\relic`, "异步任务与定时作业的执行器。", 28 * time.Hour, 2},
		{"后端服务", "Sable Auth", "v2.0.4", []string{"Go", "JWT"}, `E:\codes\go_projects\sable`, "统一认证与令牌签发服务。", 3 * 24 * time.Hour, 1},
		{"后端服务", "Kite Search", "v0.4.7", []string{"Rust", "Tantivy"}, `E:\codes\rust\kite`, "站内全文检索服务，增量索引。", 8 * 24 * time.Hour, 1},
		{"基础架构", "Forge CI", "v1.9.0", []string{"Go", "Docker"}, `E:\codes\infra\forge-ci`, "自建持续集成流水线，支持多平台构建。", 30 * time.Hour, 2},
		{"基础架构", "Lumen Proxy", "v0.8.5", []string{"Rust", "Proxy"}, `E:\codes\infra\lumen`, "边缘代理与流量治理。", 2 * 24 * time.Hour, 1},
		{"基础架构", "Terraform Modules", "v1.2.0", []string{"HCL", "IaC"}, `E:\codes\infra\tf-modules`, "可复用的基础设施模块集合。", 6 * 24 * time.Hour, 1},
		{"实验项目", "Echo Lab", "v0.0.3", []string{"Python", "ML"}, `E:\codes\lab\echo`, "语音克隆与降噪的实验管线。", 2 * 24 * time.Hour, 1},
		{"实验项目", "Drift Sandbox", "v0.1.0", []string{"Go", "WASM"}, `E:\codes\lab\drift`, "浏览器内 WASM 沙箱运行时。", 9 * 24 * time.Hour, 1},
		{"文档与工具", "Quill Docs", "v1.0.0", []string{"Markdown", "Docs"}, `E:\codes\tools\quill`, "团队文档站生成器。", 7 * 24 * time.Hour, 1},
		{"文档与工具", "Beacon Status", "v0.5.2", []string{"Go", "Status"}, `E:\codes\tools\beacon`, "内部服务状态页与事件记录。", 14 * 24 * time.Hour, 1},
	}

	for _, spec := range projectSpecs {
		gid := groupIDs[spec.Group]
		p, err := s.CreateProject(ctx, model.ProjectInput{
			Name: spec.Name, GroupID: &gid, Tags: spec.Tags, Path: spec.Path,
			Version: spec.Version, Description: spec.Description,
		})
		if err != nil {
			return fmt.Errorf("seed project %s: %w", spec.Name, err)
		}
		if _, err := s.db.ExecContext(ctx, `UPDATE projects SET updated_at = ? WHERE id = ?`,
			time.Now().Add(-spec.UpdatedAgo).UTC().Unix(), p.ID); err != nil {
			return err
		}
		for i := 0; i < spec.DocCount; i++ {
			title, kind, content := demoProjectDoc(spec.Name, i)
			d, err := s.CreateDocument(ctx, model.DocumentInput{
				OwnerType: model.OwnerProject, OwnerID: p.ID, Title: title, Kind: kind, Content: content,
			})
			if err != nil {
				return fmt.Errorf("seed doc %s/%s: %w", spec.Name, title, err)
			}
			_ = d
		}
	}

	for i, spec := range []struct {
		Name, Version string
		Tags          []string
		Path          string
	}{
		{"Sandbox Kit", "v0.2.0", []string{"Go", "CLI"}, `E:\codes\lab\sandbox-kit`},
		{"Pixel Probe", "v0.1.0", []string{"Rust", "Graphics"}, `E:\codes\lab\pixel-probe`},
	} {
		p, err := s.CreateProject(ctx, model.ProjectInput{
			Name: spec.Name, Tags: spec.Tags, Path: spec.Path, Version: spec.Version,
			Description: "尚未归组的实验项目。",
		})
		if err != nil {
			return fmt.Errorf("seed ungrouped %s: %w", spec.Name, err)
		}
		if _, err := s.db.ExecContext(ctx, `UPDATE projects SET updated_at = ? WHERE id = ?`,
			time.Now().Add(-time.Duration(9+i*3)*24*time.Hour).UTC().Unix(), p.ID); err != nil {
			return err
		}
	}

	// Vary document timestamps so relative times read naturally.
	rows, err := s.ListDocuments(ctx, "", 0)
	if err != nil {
		return err
	}
	for i, d := range rows {
		ago := time.Duration(i+1) * 5 * time.Hour
		if d.OwnerName == "Mana Desktop" {
			switch d.Title {
			case "需求文档 v1.0":
				ago = 2 * time.Hour
			case "技术方案":
				ago = 26 * time.Hour
			case "接口设计":
				ago = 3 * 24 * time.Hour
			case "迭代计划":
				ago = 5 * 24 * time.Hour
			case "发布检查清单":
				ago = 7 * 24 * time.Hour
			}
		}
		if _, err := s.db.ExecContext(ctx, `UPDATE documents SET updated_at = ? WHERE id = ?`,
			time.Now().Add(-ago).UTC().Unix(), d.ID); err != nil {
			return err
		}
	}
	return nil
}

func demoProjectDoc(project string, i int) (title, kind, content string) {
	base := []struct {
		Title, Kind string
	}{
		{"需求文档 v1.0", "requirement"},
		{"技术方案", "design"},
		{"接口设计", "interface"},
		{"迭代计划", "plan"},
		{"发布检查清单", "other"},
	}
	if i < len(base) {
		title, kind = base[i].Title, base[i].Kind
	} else {
		title, kind = fmt.Sprintf("文档 %d", i+1), "other"
	}
	if kind == "interface" {
		kind = "api"
	}
	if project == "Mana Desktop" && i == 0 {
		content = `# 需求文档 v1.0

## 1. 背景
Mana 是本地优先的开发项目管理桌面端，把项目、项目组与文档收拢到同一个窗口。
数据仅存本机 SQLite，不依赖账号与网络。

## 2. 目标
- 项目信息一处登记，路径与版本可查
- 用项目组表达归属，支持组内文档
- 常用操作三步内完成

## 3. 范围
| 模块 | v1.0 | 说明 |
| --- | --- | --- |
| 项目 | 支持 | 增删改查、标签、路径、版本 |
| 项目组 | 支持 | 左右布局、组内项目与文档 |
| 文档 | 支持 | Markdown 编辑与实时预览 |

## 4. 关键约束
删除项目会级联删除其文档；删除项目组时，组内项目移至「未分组」，组文档一并删除。

## 5. 验收
- 新建项目后重启仍存在
- 组内项目计数与列表一致
- 未保存离开时二次确认
`
	} else {
		content = fmt.Sprintf("# %s\n\n%s 的项目文档，记录当前迭代的目标与实现要点。\n\n## 进展\n- 已完成基础结构\n- 待补充测试与发布说明\n", title, project)
	}
	return
}
