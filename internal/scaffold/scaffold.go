package scaffold

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"micromanager/internal/config"
	"micromanager/internal/lang"
)

// InitOptions customizes repository initialization.
type InitOptions struct {
	BackendLang    string
	FrontendLang   string
	FrontendServer string
	FrontendClient string
	DatabaseEngine string
	PackageManager string
}

// NewServiceOptions configures service scaffolding.
type NewServiceOptions struct {
	FrontendOnly bool
	Empty        bool
	Defaults     config.Defaults
	Dependencies []string
}

// InitRepo creates .mm structure, services directory, build directory, and defaults.
func InitRepo(ctx context.Context, root string, opts InitOptions) (config.Defaults, error) {
	_ = ctx
	defaults := config.DefaultDefaults()
	if opts.BackendLang != "" {
		defaults.Backend.Lang = opts.BackendLang
	}
	if opts.FrontendLang != "" {
		defaults.Frontend.Lang = opts.FrontendLang
	}
	if opts.FrontendServer != "" {
		defaults.Frontend.Server = opts.FrontendServer
	}
	if opts.FrontendClient != "" {
		defaults.Frontend.Client = opts.FrontendClient
	}
	if opts.PackageManager != "" {
		defaults.Frontend.PackageManager = opts.PackageManager
	}
	if opts.DatabaseEngine != "" {
		defaults.Database.Engine = opts.DatabaseEngine
	}

	requiredDirs := []string{
		filepath.Join(root, ".mm"),
		filepath.Join(root, ".mm", "instructions"),
		filepath.Join(root, "services"),
		filepath.Join(root, "build"),
	}
	for _, dir := range requiredDirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return config.Defaults{}, err
		}
	}

	if err := config.SaveDefaults(root, defaults); err != nil {
		return config.Defaults{}, err
	}

	if err := ensureBuildIgnored(root); err != nil {
		return config.Defaults{}, err
	}
	return defaults, nil
}

// NewService scaffolds a service directory according to options and defaults.
func NewService(root, name string, opts NewServiceOptions) (config.ServiceConfig, error) {
	servicePath := filepath.Join(root, "services", name)
	if err := os.MkdirAll(servicePath, 0o755); err != nil {
		return config.ServiceConfig{}, err
	}

	svcCfg := config.ServiceConfig{}

	if opts.Empty {
		svcCfg.General.Lang = opts.Defaults.Backend.Lang
		svcCfg.General.External = true
	} else if opts.FrontendOnly {
		svcCfg.General.Lang = opts.Defaults.Frontend.Lang
	} else {
		svcCfg.General.Lang = opts.Defaults.Backend.Lang
		if opts.Defaults.Database.Engine != "" {
			svcCfg.General.Database = opts.Defaults.Database.Engine
		}
	}

	if len(opts.Dependencies) > 0 {
		svcCfg.Dependencies.Services = append([]string{}, opts.Dependencies...)
	}

	// Persist config before creating files, so downstream logic can read it if needed.
	if err := config.SaveServiceConfig(root, name, svcCfg); err != nil {
		return config.ServiceConfig{}, err
	}

	if err := scaffoldServiceFiles(root, name, svcCfg, opts); err != nil {
		return config.ServiceConfig{}, err
	}

	// If this service has dependencies, ensure those services have stub/mock/client
	if svcCfg.HasDependencies() {
		for _, depName := range svcCfg.Dependencies.Services {
			if err := ensureDependencyStubs(root, depName); err != nil {
				return config.ServiceConfig{}, err
			}
		}
	}

	if err := addDocInstruction(root, name); err != nil {
		return config.ServiceConfig{}, err
	}

	return svcCfg, nil
}

func scaffoldServiceFiles(root, name string, cfg config.ServiceConfig, opts NewServiceOptions) error {
	servicePath := filepath.Join(root, "services", name)

	// Attempt to use a language pack for this service language if available
	if !opts.Empty {
		if p, err := lang.FindByLang(root, cfg.General.Lang); err != nil {
			return err
		} else if p != nil {
			if err := lang.ApplyService(root, *p, name); err != nil {
				return err
			}
			// Pack applied successfully; do not run default scaffolding
			return nil
		}
	}

	if opts.Empty {
		if err := writeFile(filepath.Join(servicePath, "Dockerfile"), defaultDockerfile(cfg.General.Lang)); err != nil {
			return err
		}
		if err := writeFile(filepath.Join(servicePath, "README.md"), serviceReadme(name)); err != nil {
			return err
		}
		return nil
	}

	if opts.FrontendOnly {
		return scaffoldFrontend(servicePath, cfg, opts)
	}

	return scaffoldBackend(servicePath, cfg, opts)
}

func scaffoldBackend(servicePath string, cfg config.ServiceConfig, opts NewServiceOptions) error {
	requiredDirs := []string{
		filepath.Join(servicePath, "api"),
		filepath.Join(servicePath, "core"),
		filepath.Join(servicePath, "server"),
	}

	if cfg.HasDatabase() {
		requiredDirs = append(requiredDirs, filepath.Join(servicePath, "database"))
	}

	// Create stub and client only for internal services (not external)
	if !cfg.General.External {
		requiredDirs = append(requiredDirs,
			filepath.Join(servicePath, "stub"),
			filepath.Join(servicePath, "client"),
		)
	}

	for _, dir := range requiredDirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	if err := writeFile(filepath.Join(servicePath, "Dockerfile"), defaultDockerfile(cfg.General.Lang)); err != nil {
		return err
	}
	if err := writeFile(filepath.Join(servicePath, "README.md"), serviceReadme(filepath.Base(servicePath))); err != nil {
		return err
	}
	if err := scaffoldBackendAPI(servicePath, cfg.General.Lang); err != nil {
		return err
	}

	// Create mock directory for testing this service itself
	mockDir := filepath.Join(servicePath, "mock")
	if err := os.MkdirAll(mockDir, 0o755); err != nil {
		return err
	}

	return nil
}

func scaffoldFrontend(servicePath string, cfg config.ServiceConfig, opts NewServiceOptions) error {
	requiredDirs := []string{
		filepath.Join(servicePath, "app"),
		filepath.Join(servicePath, "components"),
		filepath.Join(servicePath, "public"),
	}
	for _, dir := range requiredDirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	if err := writeFile(filepath.Join(servicePath, "Dockerfile"), frontendDockerfile(cfg.General.Lang)); err != nil {
		return err
	}
	if err := writeFile(filepath.Join(servicePath, "README.md"), serviceReadme(filepath.Base(servicePath))); err != nil {
		return err
	}
	if err := writeFile(filepath.Join(servicePath, "package.json"), frontendPackageJSON(cfg, opts)); err != nil {
		return err
	}
	if err := writeFile(filepath.Join(servicePath, packageLockName(opts.Defaults.Frontend.PackageManager)), "{}"); err != nil {
		return err
	}
	// Minimal Next.js app directory content.
	if err := writeFile(filepath.Join(servicePath, "app", "page.tsx"), "export default function Page() { return <main>hello</main> }\n"); err != nil {
		return err
	}
	if err := writeFile(filepath.Join(servicePath, "app", "layout.tsx"), "export default function RootLayout({ children }: { children: React.ReactNode }) { return <html><body>{children}</body></html> }\n"); err != nil {
		return err
	}

	return nil
}

func packageLockName(manager string) string {
	switch strings.ToLower(manager) {
	case "pnpm":
		return "pnpm-lock.yaml"
	case "yarn":
		return "yarn.lock"
	default:
		return "package-lock.json"
	}
}

func serviceReadme(name string) string {
	return fmt.Sprintf("# %s\n\nService scaffold generated by mm.\n", name)
}

func writeFile(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func defaultDockerfile(lang string) string {
	return fmt.Sprintf("# Dockerfile for %s service\nFROM alpine\nCMD [\"echo\", \"stub\"]\n", lang)
}

func frontendDockerfile(lang string) string {
	_ = lang
	return "# Dockerfile for frontend service\nFROM node:18-alpine\nCMD [\"node\", \"server.js\"]\n"
}

func frontendPackageJSON(cfg config.ServiceConfig, opts NewServiceOptions) string {
	name := filepath.Base(opts.Defaults.Frontend.Server)
	_ = cfg
	return fmt.Sprintf(`{
  "name": "%s",
  "private": true,
  "scripts": {
    "dev": "next dev",
    "build": "next build",
    "start": "next start"
  },
  "dependencies": {
    "next": "latest",
    "react": "latest",
    "react-dom": "latest"
  }
}
`, name)
}

func scaffoldBackendAPI(servicePath, lang string) error {
	var apiContent string
	switch strings.ToLower(lang) {
	case "go":
		apiContent = goAPITemplate(filepath.Base(servicePath))
	default:
		apiContent = genericAPIComment()
	}

	return writeFile(filepath.Join(servicePath, "api", "types.go"), apiContent)
}

func goAPITemplate(serviceName string) string {
	return fmt.Sprintf(`package api

// Service defines the interface for %s.
type Service interface {
	// Add your service methods here
}

// Request and response types for your service
type Request struct {
}

type Response struct {
}
`, serviceName)
}

func genericAPIComment() string {
	return `// api package contains service interfaces and data structures.
// Implement these in core/, consume in client/, serve in server/,
// and stub/mock for testing.
`
}

func addDocInstruction(root, serviceName string) error {
	instructionsDir := filepath.Join(root, ".mm", "instructions")
	if err := os.MkdirAll(instructionsDir, 0o755); err != nil {
		return err
	}
	stamp := time.Now().UTC().Format("20060102T150405")
	file := filepath.Join(instructionsDir, fmt.Sprintf("%s.md", stamp))
	body := fmt.Sprintf(`# Update Instructions - %s

## Service: %s

Please update README.md and related docs for the new service. After completing, delete this file.
`, stamp, serviceName)
	return os.WriteFile(file, []byte(body), 0o644)
}

func ensureDependencyStubs(root, depName string) error {
	depPath := filepath.Join(root, "services", depName)
	depInfo, err := os.Stat(depPath)
	if err != nil {
		// Dependency service doesn't exist yet, skip
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if !depInfo.IsDir() {
		return fmt.Errorf("dependency %s is not a directory", depName)
	}

	// Dependency exists, ensure it has stub, mock, client directories
	for _, dir := range []string{"stub", "mock", "client"} {
		dirPath := filepath.Join(depPath, dir)
		if err := os.MkdirAll(dirPath, 0o755); err != nil {
			return err
		}
	}
	return nil
}

func ensureBuildIgnored(root string) error {
	path := filepath.Join(root, ".gitignore")
	content, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if strings.Contains(string(content), "build/") {
		return nil
	}
	updated := string(content)
	if len(updated) > 0 && !strings.HasSuffix(updated, "\n") {
		updated += "\n"
	}
	updated += "build/\n"
	return os.WriteFile(path, []byte(updated), 0o644)
}
