package service

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"mana/internal/model"
	"mana/internal/store"
)

// FieldError is a validation error attached to a form field.
type FieldError struct {
	Field   string
	Message string
}

func (e *FieldError) Error() string { return e.Message }

// Service applies validation and business rules over the store.
type Service struct {
	store *store.Store
}

// New returns a Service.
func New(s *store.Store) *Service { return &Service{store: s} }

// Store exposes the store for diagnostics and tests.
func (s *Service) Store() *store.Store { return s.store }

// Snapshot loads groups, projects and document metadata.
func (s *Service) Snapshot(ctx context.Context) (model.Snapshot, error) {
	return s.store.Snapshot(ctx)
}

// Groups returns all groups.
func (s *Service) Groups(ctx context.Context) ([]model.Group, error) {
	return s.store.ListGroups(ctx)
}

// CreateGroup validates and inserts a group.
func (s *Service) CreateGroup(ctx context.Context, in model.GroupInput) (model.Group, error) {
	in, err := validateGroup(in)
	if err != nil {
		return model.Group{}, err
	}
	g, err := s.store.CreateGroup(ctx, in)
	if errors.Is(err, store.ErrConflict) {
		return model.Group{}, &FieldError{Field: "name", Message: "该名称已被其他项目组使用"}
	}
	return g, err
}

// UpdateGroup validates and updates a group.
func (s *Service) UpdateGroup(ctx context.Context, id int64, in model.GroupInput) (model.Group, error) {
	in, err := validateGroup(in)
	if err != nil {
		return model.Group{}, err
	}
	g, err := s.store.UpdateGroup(ctx, id, in)
	if errors.Is(err, store.ErrConflict) {
		return model.Group{}, &FieldError{Field: "name", Message: "该名称已被其他项目组使用"}
	}
	return g, err
}

// DeleteGroup removes a group and migrates its projects.
func (s *Service) DeleteGroup(ctx context.Context, id int64) error {
	return s.store.DeleteGroup(ctx, id)
}

// Projects returns all projects.
func (s *Service) Projects(ctx context.Context) ([]model.Project, error) {
	return s.store.ListProjects(ctx)
}

// Project returns a project.
func (s *Service) Project(ctx context.Context, id int64) (model.Project, error) {
	return s.store.GetProject(ctx, id)
}

// CreateProject validates and inserts a project.
func (s *Service) CreateProject(ctx context.Context, in model.ProjectInput) (model.Project, error) {
	in, err := s.validateProject(ctx, in)
	if err != nil {
		return model.Project{}, err
	}
	p, err := s.store.CreateProject(ctx, in)
	if errors.Is(err, store.ErrConflict) {
		return model.Project{}, &FieldError{Field: "name", Message: "该组下已存在同名项目"}
	}
	return p, err
}

// UpdateProject validates and updates a project.
func (s *Service) UpdateProject(ctx context.Context, id int64, in model.ProjectInput) (model.Project, error) {
	in, err := s.validateProject(ctx, in)
	if err != nil {
		return model.Project{}, err
	}
	p, err := s.store.UpdateProject(ctx, id, in)
	if errors.Is(err, store.ErrConflict) {
		return model.Project{}, &FieldError{Field: "name", Message: "该组下已存在同名项目"}
	}
	return p, err
}

// DeleteProject removes a project and its documents.
func (s *Service) DeleteProject(ctx context.Context, id int64) error {
	return s.store.DeleteProject(ctx, id)
}

// MoveProject moves a project to a group; nil means ungrouped.
func (s *Service) MoveProject(ctx context.Context, id int64, groupID *int64) error {
	if groupID != nil {
		if _, err := s.store.GetGroup(ctx, *groupID); err != nil {
			return err
		}
	}
	return s.store.MoveProject(ctx, id, groupID)
}

// Documents returns documents for an owner, or all metadata when ownerType is empty.
func (s *Service) Documents(ctx context.Context, ownerType string, ownerID int64) ([]model.Document, error) {
	return s.store.ListDocuments(ctx, ownerType, ownerID)
}

// Document loads one document with content.
func (s *Service) Document(ctx context.Context, id int64) (model.Document, error) {
	return s.store.GetDocument(ctx, id)
}

// SaveDocument creates (id == 0) or updates a document.
func (s *Service) SaveDocument(ctx context.Context, id int64, in model.DocumentInput) (model.Document, error) {
	in, err := s.validateDocument(ctx, in)
	if err != nil {
		return model.Document{}, err
	}
	if id == 0 {
		d, err := s.store.CreateDocument(ctx, in)
		if errors.Is(err, store.ErrConflict) {
			return model.Document{}, &FieldError{Field: "title", Message: "该归属下已存在同名文档"}
		}
		return d, err
	}
	d, err := s.store.UpdateDocument(ctx, id, in)
	if errors.Is(err, store.ErrConflict) {
		return model.Document{}, &FieldError{Field: "title", Message: "该归属下已存在同名文档"}
	}
	return d, err
}

// DeleteDocument removes a document.
func (s *Service) DeleteDocument(ctx context.Context, id int64) error {
	return s.store.DeleteDocument(ctx, id)
}

// CountDocuments counts an owner's documents.
func (s *Service) CountDocuments(ctx context.Context, ownerType string, ownerID int64) (int, error) {
	return s.store.CountDocuments(ctx, ownerType, ownerID)
}

func validateGroup(in model.GroupInput) (model.GroupInput, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.Description = strings.TrimSpace(in.Description)
	if n := utf8.RuneCountInString(in.Name); n == 0 {
		return in, &FieldError{Field: "name", Message: "请输入项目组名称"}
	} else if n > 60 {
		return in, &FieldError{Field: "name", Message: "项目组名称最多 60 个字符"}
	}
	if n := utf8.RuneCountInString(in.Description); n > 300 {
		return in, &FieldError{Field: "description", Message: "描述最多 300 个字符"}
	}
	if in.Color == "" {
		in.Color = "#7C5CFF"
	}
	if !validColor(in.Color) {
		return in, &FieldError{Field: "color", Message: "请选择有效的标识色"}
	}
	return in, nil
}

func (s *Service) validateProject(ctx context.Context, in model.ProjectInput) (model.ProjectInput, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.Description = strings.TrimSpace(in.Description)
	in.Path = strings.TrimSpace(in.Path)
	in.Version = strings.TrimSpace(in.Version)
	if n := utf8.RuneCountInString(in.Name); n == 0 {
		return in, &FieldError{Field: "name", Message: "请输入项目名称"}
	} else if n > 80 {
		return in, &FieldError{Field: "name", Message: "项目名称最多 80 个字符"}
	}
	if n := utf8.RuneCountInString(in.Description); n > 1000 {
		return in, &FieldError{Field: "description", Message: "备注最多 1000 个字符"}
	}
	if n := utf8.RuneCountInString(in.Path); n > 512 {
		return in, &FieldError{Field: "path", Message: "项目路径最多 512 个字符"}
	}
	if in.Path != "" && !filepath.IsAbs(in.Path) {
		return in, &FieldError{Field: "path", Message: "请输入绝对路径"}
	}
	if n := utf8.RuneCountInString(in.Version); n > 32 {
		return in, &FieldError{Field: "version", Message: "版本最多 32 个字符"}
	}
	tags, err := normalizeTags(in.Tags)
	if err != nil {
		return in, err
	}
	in.Tags = tags
	if in.GroupID != nil {
		if _, err := s.store.GetGroup(ctx, *in.GroupID); err != nil {
			return in, &FieldError{Field: "group", Message: "所选项目组不存在"}
		}
	}
	return in, nil
}

func normalizeTags(tags []string) ([]string, error) {
	out := make([]string, 0, len(tags))
	seen := map[string]bool{}
	for _, raw := range tags {
		tag := strings.TrimSpace(raw)
		if tag == "" {
			continue
		}
		if n := utf8.RuneCountInString(tag); n > 16 {
			return nil, &FieldError{Field: "tags", Message: "单个标签最多 16 个字符"}
		}
		key := strings.ToLower(tag)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, tag)
		if len(out) > 5 {
			return nil, &FieldError{Field: "tags", Message: "最多添加 5 个标签"}
		}
	}
	return out, nil
}

func (s *Service) validateDocument(ctx context.Context, in model.DocumentInput) (model.DocumentInput, error) {
	in.Title = strings.TrimSpace(in.Title)
	in.Content = strings.ReplaceAll(in.Content, "\r\n", "\n")
	in.Link = strings.TrimSpace(in.Link)
	in.Kind = model.NormalizeKind(in.Kind)
	if in.OwnerType != model.OwnerProject && in.OwnerType != model.OwnerGroup {
		return in, &FieldError{Field: "owner", Message: "文档归属无效"}
	}
	if in.OwnerType == model.OwnerProject {
		if _, err := s.store.GetProject(ctx, in.OwnerID); err != nil {
			return in, &FieldError{Field: "owner", Message: "所属项目不存在"}
		}
	} else if _, err := s.store.GetGroup(ctx, in.OwnerID); err != nil {
		return in, &FieldError{Field: "owner", Message: "所属项目组不存在"}
	}
	if n := utf8.RuneCountInString(in.Title); n == 0 {
		return in, &FieldError{Field: "title", Message: "请输入文档标题"}
	} else if n > 120 {
		return in, &FieldError{Field: "title", Message: "文档标题最多 120 个字符"}
	}
	if n := utf8.RuneCountInString(in.Content); n > 200000 {
		return in, &FieldError{Field: "content", Message: "文档内容过长，请拆分"}
	}
	if n := utf8.RuneCountInString(in.Link); n > 512 {
		return in, &FieldError{Field: "link", Message: "关联链接最多 512 个字符"}
	}
	return in, nil
}

func validColor(s string) bool {
	if len(s) != 7 || s[0] != '#' {
		return false
	}
	for _, r := range s[1:] {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return false
		}
	}
	return true
}

// FormatError turns a service error into a user-facing string.
func FormatError(err error) string {
	if err == nil {
		return ""
	}
	var fe *FieldError
	if errors.As(err, &fe) {
		return fe.Message
	}
	if errors.Is(err, store.ErrNotFound) {
		return "记录不存在或已被删除"
	}
	return fmt.Sprintf("操作失败：%v", err)
}
