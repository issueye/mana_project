package model

import "time"

// Group is a project group.
type Group struct {
	ID            int64
	Name          string
	Description   string
	Color         string
	ProjectCount  int
	DocumentCount int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Project is a development project. GroupID is nil for an ungrouped project.
type Project struct {
	ID            int64
	GroupID       *int64
	GroupName     string
	Name          string
	Tags          []string
	Description   string
	Path          string
	Version       string
	DocumentCount int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Document belongs to exactly one project or group. A note carries its
// text in Content; a file document carries it in FileName and FileSize
// (with extracted preview text in Content).
type Document struct {
	ID        int64
	OwnerType string // "project" or "group"
	OwnerID   int64
	OwnerName string
	Title     string
	Kind      string
	Content   string
	Link      string
	Words     int
	FileName  string
	FileSize  int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ProjectInput is the editable project data.
type ProjectInput struct {
	Name        string
	GroupID     *int64
	Tags        []string
	Description string
	Path        string
	Version     string
}

// GroupInput is the editable group data.
type GroupInput struct {
	Name        string
	Description string
	Color       string
}

// DocumentInput is the editable document data.
type DocumentInput struct {
	OwnerType string
	OwnerID   int64
	Title     string
	Kind      string
	Content   string
	Link      string
}

// Snapshot is the complete in-memory view of the database.
type Snapshot struct {
	Groups   []Group
	Projects []Project
	Docs     []Document
}

// SearchResult is the outcome of a global search, one capped section per
// kind of record.
type SearchResult struct {
	Query    string
	Projects []Project
	Docs     []Document
	Groups   []Group
}

// DocumentKind is one of the fixed document types.
type DocumentKind string

const (
	KindRequirement DocumentKind = "requirement"
	KindDesign      DocumentKind = "design"
	KindPlan        DocumentKind = "plan"
	KindAPI         DocumentKind = "api"
	KindMeeting     DocumentKind = "meeting"
	KindOther       DocumentKind = "other"
)

// KindLabels maps stored kinds to the Chinese labels used by the UI.
var KindLabels = map[string]string{
	"requirement": "需求",
	"design":      "设计",
	"plan":        "开发计划",
	"api":         "接口",
	"meeting":     "会议纪要",
	"other":       "其他",
}

// KindLabel returns the UI label for a kind.
func KindLabel(kind string) string {
	if label, ok := KindLabels[kind]; ok {
		return label
	}
	return "其他"
}

// NormalizeKind maps a label or a stored kind to a stored kind.
func NormalizeKind(s string) string {
	for k, v := range KindLabels {
		if s == k || s == v {
			return k
		}
	}
	return "other"
}

// OwnerType constants.
const (
	OwnerProject = "project"
	OwnerGroup   = "group"
)

// Date formats.
func Date(t time.Time) string { return t.Local().Format("2006-01-02") }

func DateTime(t time.Time) string { return t.Local().Format("2006-01-02 15:04") }

func Relative(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "刚刚"
	case d < time.Hour:
		return itoa(int(d.Minutes())) + " 分钟前"
	case d < 24*time.Hour:
		return itoa(int(d.Hours())) + " 小时前"
	case d < 48*time.Hour:
		return "昨天"
	case d < 7*24*time.Hour:
		return itoa(int(d.Hours()/24)) + " 天前"
	default:
		return t.Local().Format("01-02")
	}
}

func itoa(n int) string {
	if n <= 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
