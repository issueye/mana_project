// Package gitcmd drives the system git executable. Everything goes
// through argv (never a shell), with timeouts so a stuck index or a
// hanging credential prompt cannot freeze the caller; errors carry the
// stderr text so the UI can show what git said.
package gitcmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ErrNoGit reports that the git executable is not on PATH.
var ErrNoGit = errors.New("没有找到 git 命令，请确认已安装并加入 PATH")

// Timeouts: normal queries are quick; pushes may legitimately take a
// while over slow links.
const (
	defaultTimeout = 10 * time.Second
	pushTimeout    = 120 * time.Second
)

// run executes git with args in dir and returns trimmed stdout.
func run(ctx context.Context, dir string, timeout time.Duration, args ...string) (string, error) {
	if _, err := exec.LookPath("git"); err != nil {
		return "", ErrNoGit
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	var out, errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	// Chinese file names arrive raw instead of \xNN-quoted.
	cmd.Env = append(os.Environ(), "core.quotepath=false", "GIT_OPTIONAL_LOCKS=0")
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(errOut.String())
		if msg == "" {
			msg = err.Error()
		}
		return out.String(), fmt.Errorf("git %s：%s", strings.Join(args, " "), msg)
	}
	return out.String(), nil
}

// IsRepo reports whether dir sits inside a git work tree.
func IsRepo(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, ".git"))
	if err == nil && (info.IsDir() || info.Mode().IsRegular()) { // regular .git file = worktree
		return true
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	_, err = run(ctx, dir, defaultTimeout, "rev-parse", "--is-inside-work-tree")
	return err == nil
}

// HasRemote reports whether the origin remote is configured.
func HasRemote(dir string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	out, err := run(ctx, dir, defaultTimeout, "remote")
	return err == nil && strings.Contains(out, "origin")
}

// Init creates a repository in dir.
func Init(dir string) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	_, err := run(ctx, dir, defaultTimeout, "init")
	return err
}

// FileStatus is one entry of `git status --porcelain`.
type FileStatus struct {
	Code     string // two letters, e.g. "M ", " M", "??"
	Path     string
	Staged   bool
	Deleted  bool
	Untracked bool
}

// Status reads the branch name and the working tree state.
func Status(ctx context.Context, dir string) (branch string, files []FileStatus, err error) {
	out, err := run(ctx, dir, defaultTimeout, "status", "--porcelain=v1", "-z", "--branch")
	if err != nil {
		return "", nil, err
	}
	parts := strings.Split(out, "\x00")
	for i, part := range parts {
		if i == 0 {
			branch = strings.TrimPrefix(strings.TrimSpace(part), "## ")
			if head, _, ok := strings.Cut(branch, "..."); ok {
				branch = head // drop the upstream when tracking
			}
			if branch == "" {
				branch = "main"
			}
			continue
		}
		if len(part) < 3 {
			continue
		}
		code, path := part[:2], part[3:]
		f := FileStatus{Code: code, Path: path}
		f.Staged = code[0] != ' ' && code[0] != '?'
		f.Deleted = strings.Contains(code, "D")
		f.Untracked = code == "??"
		files = append(files, f)
	}
	return branch, files, nil
}

// Stage adds paths to the index; an empty list stages everything.
func Stage(ctx context.Context, dir string, paths []string) error {
	args := []string{"add", "--"}
	if len(paths) == 0 {
		args = []string{"add", "-A", "--"}
	}
	args = append(args, paths...)
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()
	_, err := run(ctx, dir, defaultTimeout, args...)
	return err
}

// Unstage removes paths from the index, keeping the working tree.
func Unstage(ctx context.Context, dir string, paths []string) error {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()
	_, err := run(ctx, dir, defaultTimeout, append([]string{"reset", "--", ""}, paths...)...)
	return err
}

// Commit records the staged changes; it returns the short hash.
func Commit(ctx context.Context, dir, message string) (string, error) {
	message = strings.TrimSpace(message)
	if message == "" {
		return "", errors.New("请填写提交说明")
	}
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()
	if _, err := run(ctx, dir, defaultTimeout, "commit", "-m", message); err != nil {
		return "", err
	}
	out, err := run(ctx, dir, defaultTimeout, "rev-parse", "--short", "HEAD")
	if err != nil {
		return "", nil // committed but hash unreadable; not fatal
	}
	return strings.TrimSpace(out), nil
}

// LogEntry is one line of history.
type LogEntry struct {
	Hash    string
	Author  string
	Time    time.Time
	Subject string
}

// Log reads the newest commits, newest first.
func Log(ctx context.Context, dir string, limit int) ([]LogEntry, error) {
	if limit <= 0 {
		limit = 50
	}
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()
	out, err := run(ctx, dir, defaultTimeout,
		"log", "-n", strconv.Itoa(limit), "--pretty=format:%h%x09%an%x09%at%x09%s")
	if err != nil {
		// An empty repository has no commits yet; that is not an error.
		if strings.Contains(err.Error(), "does not have any commits yet") ||
			strings.Contains(err.Error(), "bad revision") {
			return nil, nil
		}
		return nil, err
	}
	var commits []LogEntry
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Split(strings.TrimSpace(line), "\t")
		if len(fields) < 4 {
			continue
		}
		sec, _ := strconv.ParseInt(fields[2], 10, 64)
		commits = append(commits, LogEntry{
			Hash: fields[0], Author: fields[1],
			Time: time.Unix(sec, 0), Subject: fields[3],
		})
	}
	return commits, nil
}

// Tag is one ref under refs/tags.
type Tag struct {
	Name    string
	Hash    string
	Created time.Time
	Subject string
}

// Tags lists the repository's tags, newest first.
func Tags(ctx context.Context, dir string) ([]Tag, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()
	const format = "%(refname:short)%09%(*objectname)%09%(objectname)%09%(creatordate:unix)%09%(contents:subject)"
	out, err := run(ctx, dir, defaultTimeout,
		"for-each-ref", "refs/tags", "--sort=-creatordate", "--format="+format)
	if err != nil {
		return nil, err
	}
	var tags []Tag
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Split(strings.TrimSpace(line), "\t")
		if len(fields) < 4 {
			continue
		}
		hash := fields[1]
		if hash == "" {
			hash = fields[2] // lightweight tag points straight at the commit
		}
		sec, _ := strconv.ParseInt(fields[3], 10, 64)
		subject := ""
		if len(fields) > 4 {
			subject = fields[4]
		}
		tags = append(tags, Tag{Name: fields[0], Hash: hash, Created: time.Unix(sec, 0), Subject: subject})
	}
	return tags, nil
}

// CreateTag makes an annotated tag.
func CreateTag(ctx context.Context, dir, name, message string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("请填写标签名称")
	}
	if message == "" {
		message = name
	}
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()
	_, err := run(ctx, dir, defaultTimeout, "tag", "-a", name, "-m", message)
	return err
}

// Push sends the current branch and the given tag to origin. It needs a
// little longer: a slow link is normal.
func Push(ctx context.Context, dir, tag string) error {
	if !HasRemote(dir) {
		return errors.New("没有配置远端 origin，请先在终端添加")
	}
	ctx, cancel := context.WithTimeout(ctx, pushTimeout)
	defer cancel()
	if _, err := run(ctx, dir, pushTimeout, "push", "-u", "origin", "HEAD"); err != nil {
		return err
	}
	if tag != "" {
		if _, err := run(ctx, dir, pushTimeout, "push", "origin", tag); err != nil {
			return err
		}
	}
	return nil
}

// SortDirsFirst orders directory entries the way a file manager does.
func SortDirsFirst(entries []os.DirEntry) {
	sort.SliceStable(entries, func(i, j int) bool {
		pi, pj := entries[i].IsDir(), entries[j].IsDir()
		if pi != pj {
			return pi
		}
		return strings.ToLower(entries[i].Name()) < strings.ToLower(entries[j].Name())
	})
}
