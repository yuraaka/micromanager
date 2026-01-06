package lang

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"

	toml "github.com/pelletier/go-toml/v2"
)

// Metadata describes a language pack as defined in language.toml.
type Metadata struct {
	ID          string `toml:"id"`
	Name        string `toml:"name"`
	Lang        string `toml:"lang"`
	Version     string `toml:"version"`
	Description string `toml:"description,omitempty"`
}

// Pack represents a loaded language pack.
type Pack struct {
	Meta    Metadata
	BaseDir string // directory containing language.toml and templates
}

// TemplateData is passed into templates during rendering.
type TemplateData struct {
	ProjectName string
	ServiceName string
}

// PacksDir returns the default packs directory under the repo root.
func PacksDir(root string) string {
	return filepath.Join(root, ".mm", "packs")
}

// LoadAll scans the packs directory for language.toml files and loads packs.
func LoadAll(root string) ([]Pack, error) {
	base := PacksDir(root)
	entries, err := os.ReadDir(base)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var packs []Pack
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		packDir := filepath.Join(base, e.Name())
		metaPath := filepath.Join(packDir, "language.toml")
		data, err := os.ReadFile(metaPath)
		if err != nil {
			// skip invalid pack folder
			continue
		}
		var m Metadata
		if err := toml.Unmarshal(data, &m); err != nil {
			continue
		}
		p := Pack{Meta: m, BaseDir: packDir}
		if err := Validate(p); err != nil {
			// skip invalid packs
			continue
		}
		packs = append(packs, p)
	}
	return packs, nil
}

// Validate ensures required metadata and template paths exist.
func Validate(p Pack) error {
	if strings.TrimSpace(p.Meta.ID) == "" {
		return fmt.Errorf("pack missing id")
	}
	if strings.TrimSpace(p.Meta.Lang) == "" {
		return fmt.Errorf("pack missing lang")
	}
	// Require templates/service directory for service scaffolding
	serviceTpl := filepath.Join(p.BaseDir, "templates", "service")
	info, err := os.Stat(serviceTpl)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("pack %s missing templates/service", p.Meta.ID)
	}
	return nil
}

// FindByLang returns the first pack matching a given language.
func FindByLang(root, lang string) (*Pack, error) {
	packs, err := LoadAll(root)
	if err != nil {
		return nil, err
	}
	for _, p := range packs {
		if strings.EqualFold(p.Meta.Lang, lang) {
			pp := p
			return &pp, nil
		}
	}

	// Fallback to local pack directory (pack/lang/<lang>) if .mm/packs has none.
	if p, err := findLocalPack(root, lang); err != nil {
		return nil, err
	} else if p != nil {
		return p, nil
	}
	return nil, nil
}

// ApplyService copies pack templates into the repository using generic rules:
// - templates/service/* => services/<serviceName>/
// - templates/common/*  => services/common/ (full overwrite)
// - templates/root/*    => repo root
// After copy, go mod tidy is executed to produce go.sum.
func ApplyService(root string, p Pack, serviceName string) error {
	base := filepath.Join(p.BaseDir, "templates")
	vars := TemplateData{
		ProjectName: detectProjectName(root),
		ServiceName: serviceName,
	}

	serviceSrc := filepath.Join(base, "service")
	if !pathExists(serviceSrc) {
		return fmt.Errorf("pack %s missing templates/service", p.Meta.ID)
	}
	serviceDst := filepath.Join(root, "services", serviceName)
	if err := copyTemplateTree(serviceSrc, serviceDst, vars, true); err != nil {
		return err
	}

	commonSrc := filepath.Join(base, "common")
	if pathExists(commonSrc) {
		commonDst := filepath.Join(root, "services", "common")
		if err := os.RemoveAll(commonDst); err != nil {
			return err
		}
		if err := copyTemplateTree(commonSrc, commonDst, vars, true); err != nil {
			return err
		}
	}

	rootSrc := filepath.Join(base, "root")
	if pathExists(rootSrc) {
		if err := copyTemplateTree(rootSrc, root, vars, true); err != nil {
			return err
		}
	}

	if err := runGoModTidy(root); err != nil {
		return err
	}

	return nil
}

func copyTemplateTree(src, dst string, vars TemplateData, overwrite bool) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		outPath := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(outPath, 0o755)
		}
		if strings.HasSuffix(outPath, ".tmpl") {
			outPath = strings.TrimSuffix(outPath, ".tmpl")
		}
		if !overwrite && pathExists(outPath) {
			return nil
		}
		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			return err
		}

		if isLikelyText(path) || strings.HasSuffix(path, ".tmpl") {
			return renderTemplateFile(path, outPath, vars)
		}
		return copyFile(path, outPath)
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return nil
}

func isLikelyText(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".md", ".txt", ".go", ".js", ".ts", ".tsx", ".json", ".yaml", ".yml", ".toml", ".css", ".html":
		return true
	}
	// Handle Dockerfile and files without extension
	base := filepath.Base(path)
	if base == "Dockerfile" || base == "dockerfile" {
		return true
	}
	return false
}

func renderTemplateFile(src, dst string, vars TemplateData) error {
	raw, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	tpl, err := template.New(filepath.Base(src)).Funcs(templateFuncMap()).Parse(string(raw))
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, vars); err != nil {
		return err
	}

	return os.WriteFile(dst, buf.Bytes(), 0o644)
}

func templateFuncMap() template.FuncMap {
	return template.FuncMap{
		"snake":    snake,
		"kebab":    kebab,
		"camel":    camel,
		"title":    strings.Title,
		"upper":    strings.ToUpper,
		"lower":    strings.ToLower,
		"joinPath": path.Join,
	}
}

func detectProjectName(root string) string {
	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "module ") {
				module := strings.TrimSpace(strings.TrimPrefix(trimmed, "module"))
				if module != "" {
					return module
				}
			}
		}
	}

	base := filepath.Base(root)
	if base == "." || base == string(filepath.Separator) || base == "" {
		base = "service"
	}
	return base
}

func snake(s string) string {
	return strings.Join(splitWords(s), "_")
}

func kebab(s string) string {
	return strings.Join(splitWords(s), "-")
}

func camel(s string) string {
	parts := splitWords(s)
	for i := range parts {
		parts[i] = strings.Title(parts[i])
	}
	return strings.Join(parts, "")
}

func splitWords(s string) []string {
	var parts []string
	var current []rune

	for _, r := range s {
		if r == '-' || r == '_' || r == ' ' {
			if len(current) > 0 {
				parts = append(parts, string(current))
				current = current[:0]
			}
			continue
		}

		if len(current) > 0 {
			last := current[len(current)-1]
			if unicode.IsUpper(r) && (unicode.IsLower(last) || unicode.IsDigit(last)) {
				parts = append(parts, string(current))
				current = current[:0]
			}
		}
		current = append(current, unicode.ToLower(r))
	}

	if len(current) > 0 {
		parts = append(parts, string(current))
	}

	return parts
}

func pathExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func runGoModTidy(root string) error {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func findLocalPack(root, lang string) (*Pack, error) {
	// Search current and parent directories for pack/lang/<lang>
	langLower := strings.ToLower(lang)
	maxAscend := 5
	cur := root
	for i := 0; i <= maxAscend; i++ {
		candidate := filepath.Join(cur, "pack", "lang", langLower)
		info, err := os.Stat(candidate)
		if err == nil && info.IsDir() {
			// Optional safety: ensure templates/service exists
			if pathExists(filepath.Join(candidate, "templates", "service")) {
				meta := Metadata{
					ID:      "local-" + langLower,
					Name:    strings.Title(langLower) + " pack",
					Lang:    lang,
					Version: "0.0.0",
				}
				return &Pack{Meta: meta, BaseDir: candidate}, nil
			}
		}
		// Move up one directory
		parent := filepath.Dir(cur)
		if parent == cur { // reached filesystem root
			break
		}
		cur = parent
	}
	return nil, nil
}
