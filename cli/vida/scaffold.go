package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/alkemist-17/vida"
)

//go:embed templates
var templatesFS embed.FS

const (
	defaultVersion   = "0.1.0"
	manifestFileName = "vida.toml"
)

type manifest struct {
	Name         string
	Version      string
	Entry        string
	Test         string
	Dependencies map[string]string
}

func scaffold(args []string) {
	if len(args) < 3 {
		handleError(fmt.Errorf("no project name given.\n\tUsage: vida init <name> [--template=app|lib]"))
	}

	projectName := args[2]
	template := "app"
	for _, a := range args[3:] {
		if after, ok := strings.CutPrefix(a, "--template="); ok {
			template = after
		}
	}

	if template != "app" && template != "lib" {
		handleError(fmt.Errorf("unknown template %q. Available templates: app, lib", template))
	}

	targetDir, err := filepath.Abs(projectName)
	handleError(err)

	if info, err := os.Stat(targetDir); err == nil && info.IsDir() {
		entries, err := os.ReadDir(targetDir)
		handleError(err)
		if len(entries) > 0 {
			handleError(fmt.Errorf("directory %q already exists and is not empty", projectName))
		}
	}

	handleError(os.MkdirAll(targetDir, 0o755))

	srcRoot := "templates/" + template
	handleError(fs.WalkDir(templatesFS, srcRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(srcRoot, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		dest := filepath.Join(targetDir, rel)

		if d.IsDir() {
			return os.MkdirAll(dest, 0o755)
		}

		content, err := templatesFS.ReadFile(path)
		if err != nil {
			return err
		}
		content = []byte(strings.ReplaceAll(string(content), "{{.Name}}", projectName))
		return os.WriteFile(dest, content, 0o644)
	}))

	handleError(os.MkdirAll(filepath.Join(targetDir, vida.ModulesDirName, vida.RemoteModulesDir), 0o755))
	handleError(os.WriteFile(filepath.Join(targetDir, ".gitignore"), []byte(vida.ModulesDirName+"/"+vida.RemoteModulesDir+"/\n"), 0o644))

	entry := "main.vida"
	if template == "lib" {
		entry = "src/lib.vida"
	}

	m := manifest{Name: projectName, Version: defaultVersion, Entry: entry, Test: "test"}
	handleError(writeManifest(filepath.Join(targetDir, manifestFileName), m))

	clear()
	printVersion()
	fmt.Printf("\t✅ Created %q (%v template)\n\n", projectName, template)
	fmt.Printf("\tcd %v\n", projectName)
	if template == "app" {
		fmt.Printf("\tvida run\n")
		fmt.Printf("\tvida test\n\n\n\n")
	} else {
		fmt.Printf("\t# import this library from another project with:\n")
		fmt.Printf("\t# import(\"modules/%v/%v\")\n\n\n\n", projectName, entry)
	}
}

func writeManifest(path string, m manifest) error {
	var b strings.Builder
	fmt.Fprintf(&b, "name = %q\n", m.Name)
	fmt.Fprintf(&b, "version = %q\n", m.Version)
	fmt.Fprintf(&b, "entry = %q\n", m.Entry)
	fmt.Fprintf(&b, "test = %q\n", m.Test)
	b.WriteString("\n[dependencies]\n")
	if len(m.Dependencies) == 0 {
		b.WriteString("# populated by 'vida install' once you add entries here, e.g.:\n")
		b.WriteString("# mathlib = \"https://raw.githubusercontent.com/user/repo/main/src/mathlib.vida\"\n")
	} else {
		for name, depURL := range m.Dependencies {
			fmt.Fprintf(&b, "%v = %q\n", name, depURL)
		}
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func readManifest(path string) (manifest, error) {
	var m manifest
	m.Dependencies = make(map[string]string)

	data, err := os.ReadFile(path)
	if err != nil {
		return m, err
	}

	inDependencies := false
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") {
			inDependencies = strings.EqualFold(line, "[dependencies]")
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), `"`)

		if inDependencies {
			m.Dependencies[key] = value
			continue
		}
		switch key {
		case "name":
			m.Name = value
		case "version":
			m.Version = value
		case "entry":
			m.Entry = value
		case "test":
			m.Test = value
		}
	}

	if m.Entry == "" {
		return m, fmt.Errorf("%q has no 'entry' field", path)
	}
	return m, nil
}

// resolveEntryScript figures out what script to run when the user didn't
// give one explicitly: look for vida.toml in the cwd and use its entry.
func resolveEntryScript() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	manifestPath := filepath.Join(cwd, manifestFileName)
	if !fileExistsOnDiskCLI(manifestPath) {
		return "", fmt.Errorf("no script given and no %v found in the current directory", manifestFileName)
	}
	m, err := readManifest(manifestPath)
	if err != nil {
		return "", err
	}
	return filepath.Join(cwd, m.Entry), nil
}

func fileExistsOnDiskCLI(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
