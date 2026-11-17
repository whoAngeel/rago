package eval

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// privateDirName is the git-ignored folder with the owner's real documents;
// the question file that lives there is not a document.
const privateDirName = "private"

// Source is one document selected for ingestion. Name (the base file name) is
// its identity in questions and reports.
type Source struct {
	Name string
	Path string
	Dir  string
	hash [sha256.Size]byte
}

// Private reports whether the document comes from the git-ignored private folder.
func (s Source) Private() bool { return filepath.Base(s.Dir) == privateDirName }

// DefaultDataDirs returns root plus root/manuals and root/private when those
// directories exist; a missing optional directory is not an error.
func DefaultDataDirs(root string) []string {
	dirs := []string{root}
	for _, sub := range []string{"manuals", privateDirName} {
		p := filepath.Join(root, sub)
		if fi, err := os.Stat(p); err == nil && fi.IsDir() {
			dirs = append(dirs, p)
		}
	}
	return dirs
}

// CollectSources lists the regular files directly inside each directory (not
// recursive, dot files and *.json inside a "private" folder ignored). A base
// name found twice is an error. Files with the same content are ingested once:
// the shortest name wins (ties by name) and one warning per skipped file names
// both. The result is sorted by directory order, then name, so document ids
// derived from the position are stable for a given set of files.
func CollectSources(dirs []string) ([]Source, []string, error) {
	var all []Source
	byName := map[string]Source{}
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil, nil, fmt.Errorf("listing %s: %w", dir, err)
		}
		private := filepath.Base(dir) == privateDirName
		for _, e := range entries {
			name := e.Name()
			if !e.Type().IsRegular() || strings.HasPrefix(name, ".") {
				continue
			}
			if private && strings.EqualFold(filepath.Ext(name), ".json") {
				continue
			}
			path := filepath.Join(dir, name)
			if prev, ok := byName[name]; ok {
				return nil, nil, fmt.Errorf("document name %q appears in both %s and %s; names must be unique", name, prev.Dir, dir)
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return nil, nil, err
			}
			s := Source{Name: name, Path: path, Dir: dir, hash: sha256.Sum256(data)}
			byName[name] = s
			all = append(all, s)
		}
	}

	winner := map[[sha256.Size]byte]Source{}
	for _, s := range all {
		w, ok := winner[s.hash]
		if !ok || len(s.Name) < len(w.Name) || (len(s.Name) == len(w.Name) && s.Name < w.Name) {
			winner[s.hash] = s
		}
	}
	var kept []Source
	var warns []string
	for _, s := range all {
		w := winner[s.hash]
		if w.Name != s.Name {
			warns = append(warns, fmt.Sprintf("skipping %s: identical content to %s", s.Name, w.Name))
			continue
		}
		kept = append(kept, s)
	}
	sort.SliceStable(warns, func(i, j int) bool { return warns[i] < warns[j] })
	return kept, warns, nil
}

// SourceNames returns the document names, in order.
func SourceNames(srcs []Source) []string {
	names := make([]string, len(srcs))
	for i, s := range srcs {
		names[i] = s.Name
	}
	return names
}

// ReadSource returns the file content.
func ReadSource(s Source) ([]byte, error) { return os.ReadFile(s.Path) }
