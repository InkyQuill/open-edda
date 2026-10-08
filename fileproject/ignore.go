package fileproject

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type ignoreRule struct {
	pattern                      string
	include, directory, anchored bool
}

func readIgnoreRules(root *os.Root) ([]ignoreRule, error) {
	rules := []ignoreRule{{pattern: ".remember"}, {pattern: ".creative-writing"}, {pattern: ".agents/skills"}}
	info, err := root.Lstat(".eddaignore")
	if os.IsNotExist(err) {
		return rules, nil
	}
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		external, err := externalInventoryLink(root.Name(), ".eddaignore")
		if err != nil {
			return nil, err
		}
		if external {
			return rules, nil
		}
	}
	file, err := openInventoryFile(root, ".eddaignore")
	if err != nil {
		return nil, fmt.Errorf("read .eddaignore: %w", err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(io.LimitReader(file, (1<<20)+1))
	line, size := 0, 0
	for scanner.Scan() {
		line++
		size += len(scanner.Bytes()) + 1
		if size > 1<<20 {
			return nil, fmt.Errorf(".eddaignore exceeds 1 MiB")
		}
		text := strings.TrimSpace(scanner.Text())
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		rule := ignoreRule{}
		if strings.HasPrefix(text, "!") {
			rule.include = true
			text = text[1:]
		}
		rule.directory = strings.HasSuffix(text, "/")
		rule.anchored = strings.HasPrefix(text, "/")
		text = strings.TrimSuffix(strings.TrimPrefix(text, "/"), "/")
		if text == "" || strings.Contains(text, "\\") {
			return nil, fmt.Errorf(".eddaignore line %d: invalid pattern", line)
		}
		for _, part := range strings.Split(text, "/") {
			if part == "" || part == "." || part == ".." {
				return nil, fmt.Errorf(".eddaignore line %d: invalid path", line)
			}
			if _, err := path.Match(part, ""); err != nil {
				return nil, fmt.Errorf(".eddaignore line %d: %w", line, err)
			}
		}
		rule.pattern = text
		rules = append(rules, rule)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read .eddaignore: %w", err)
	}
	return rules, nil
}

func ignoreGlob(pattern, name string) bool {
	parts, names := strings.Split(pattern, "/"), strings.Split(name, "/")
	// Dynamic programming bounds matching work even for repeated ** patterns.
	previous := make([]bool, len(names)+1)
	previous[0] = true
	for _, part := range parts {
		next := make([]bool, len(names)+1)
		if part == "**" {
			next[0] = previous[0]
		}
		for j, name := range names {
			if part == "**" {
				next[j+1] = previous[j+1] || next[j]
			} else {
				match, _ := path.Match(part, name)
				next[j+1] = previous[j] && match
			}
		}
		previous = next
	}
	return previous[len(names)]
}

func ignoredByRules(name string, directory bool, rules []ignoreRule) bool {
	if name == ".eddaignore" {
		return false
	}
	parts := strings.Split(name, "/")
	// An excluded parent must be re-included before its children, as in gitignore.
	for i := range parts {
		candidate := strings.Join(parts[:i+1], "/")
		isDir := i < len(parts)-1 || directory
		ignored := false
		for _, rule := range rules {
			if rule.directory && !isDir {
				continue
			}
			target := candidate
			if !rule.anchored && !strings.Contains(rule.pattern, "/") {
				target = parts[i]
			}
			if ignoreGlob(rule.pattern, target) {
				ignored = !rule.include
			}
		}
		if ignored {
			return true
		}
	}
	return false
}

// IgnoredPaths applies the same local rules to incoming paths before an update.
func IgnoredPaths(directory string, paths map[string]bool) ([]string, error) {
	root, err := os.OpenRoot(directory)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	rules, err := readIgnoreRules(root)
	if err != nil {
		return nil, err
	}
	var ignored []string
	for name, isDir := range paths {
		if ignoredByRules(name, isDir, rules) {
			ignored = append(ignored, name)
		}
	}
	return ignored, nil
}

// External links are local references, never portable project entries. Resolving
// their target only classifies the link; no target contents are read or copied.
func externalInventoryLink(root, name string) (bool, error) {
	target, err := os.Readlink(filepath.Join(root, filepath.FromSlash(name)))
	if err != nil {
		return false, err
	}
	canonicalRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return false, err
	}
	if !filepath.IsAbs(target) {
		target = filepath.Join(canonicalRoot, filepath.FromSlash(path.Dir(name)), target)
	}
	if resolved, err := filepath.EvalSymlinks(target); err == nil {
		target = resolved
	}
	relative, err := filepath.Rel(canonicalRoot, target)
	if err != nil {
		return false, err
	}
	return !filepath.IsLocal(relative), nil
}
