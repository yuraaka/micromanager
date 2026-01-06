package lang

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"
)

type goTemplateData struct {
	ModulePath         string
	ServiceName        string
	ServiceDir         string
	ServiceSlug        string
	ServiceBinaryName  string
	ServiceRoutePrefix string
	ServiceImportPath  string
	CommonImportPath   string
}

func findLocalPack(root, lang string) (*Pack, error) {
	candidate := filepath.Join(root, "pack", "lang", strings.ToLower(lang))
	info, err := os.Stat(candidate)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	if !info.IsDir() {
		return nil, nil
	}

	meta := Metadata{
		ID:      "local-" + strings.ToLower(lang),
		Name:    capitalize(lang) + " pack",
		Lang:    lang,
		Version: "0.0.0",
	}
	return &Pack{Meta: meta, BaseDir: candidate}, nil
}

func isGoServicePack(p Pack) bool {
	if strings.EqualFold(p.Meta.Lang, "go") {
		return true
	}

	markers := []string{"api", "core", "server", "client"}
	for _, marker := range markers {
		if pathExists(filepath.Join(p.BaseDir, marker)) {
			return true
		}
	}
	return false
}

func applyGoServicePack(root string, p Pack, serviceName string) error {
	data := buildGoTemplateData(root, serviceName)

	if err := ensureGoMod(root, filepath.Join(p.BaseDir, "root", "go.mod"), data); err != nil {
		return err
	}

	serviceRoot := filepath.Join(root, "services", serviceName)
	if err := os.MkdirAll(serviceRoot, 0o755); err != nil {
		return err
	}

	for _, dir := range []string{"api", "core", "server", "client"} {
		src := filepath.Join(p.BaseDir, dir)
		if !pathExists(src) {
			continue
		}
		dst := filepath.Join(serviceRoot, dir)
		if err := copyTemplateTree(src, dst, data, true); err != nil {
			return err
		}
	}

	dockerfile := filepath.Join(p.BaseDir, "Dockerfile")
	if pathExists(dockerfile) {
		if err := renderTemplateFile(dockerfile, filepath.Join(serviceRoot, "Dockerfile"), data); err != nil {
			return err
		}
	}

	commonSrc := filepath.Join(p.BaseDir, "common", "std")
	if pathExists(commonSrc) {
		commonDst := filepath.Join(root, "services", "common", "std")
		if err := os.RemoveAll(commonDst); err != nil {
			return err
		}
		if err := copyTemplateTree(commonSrc, commonDst, data, true); err != nil {
			return err
		}
	}

	if err := runGoModTidy(root); err != nil {
		return err
	}

	return nil
}

func buildGoTemplateData(root, serviceName string) goTemplateData {
	modulePath := detectModulePath(root)
	slug := toKebab(serviceName)
	if slug == "" {
		slug = strings.ToLower(serviceName)
	}

	return goTemplateData{
		ModulePath:         modulePath,
		ServiceName:        serviceName,
		ServiceDir:         serviceName,
		ServiceSlug:        slug,
		ServiceBinaryName:  slug,
		ServiceRoutePrefix: "/" + slug + "/v1",
		ServiceImportPath:  path.Join(modulePath, "services", serviceName),
		CommonImportPath:   path.Join(modulePath, "services", "common", "std"),
	}
}

func detectModulePath(root string) string {
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
	if base == "." || base == string(filepath.Separator) {
		base = "service"
	}
	return base
}

func ensureGoMod(root, templatePath string, data goTemplateData) error {
	dest := filepath.Join(root, "go.mod")
	if pathExists(dest) {
		return nil
	}

	if pathExists(templatePath) {
		return renderTemplateFile(templatePath, dest, data)
	}

	minimal := fmt.Sprintf("module %s\n\ngo 1.21\n", data.ModulePath)
	return os.WriteFile(dest, []byte(minimal), 0o644)
}

func copyTemplateTree(src, dst string, data goTemplateData, overwrite bool) error {
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
		return renderTemplateFile(path, outPath, data)
	})
}

func renderTemplateFile(src, dst string, data goTemplateData) error {
	raw, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	tpl, err := template.New(filepath.Base(src)).Parse(string(raw))
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return err
	}

	return os.WriteFile(dst, buf.Bytes(), 0o644)
}

func runGoModTidy(root string) error {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func toKebab(s string) string {
	parts := splitWords(s)
	return strings.Join(parts, "-")
}

func toSnake(s string) string {
	parts := splitWords(s)
	return strings.Join(parts, "_")
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

func capitalize(s string) string {
	if s == "" {
		return s
	}
	lower := strings.ToLower(s)
	return strings.ToUpper(lower[:1]) + lower[1:]
}
