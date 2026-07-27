package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/alkemist-17/vida"
)

const lockFileName = "vida.lock"

type lockEntry struct {
	URL    string
	SHA256 string
}

func install(args []string) {
	clear()
	printVersion()

	cwd, err := os.Getwd()
	handleError(err)

	manifestPath := filepath.Join(cwd, manifestFileName)
	if !fileExistsOnDiskCLI(manifestPath) {
		handleError(fmt.Errorf("no %v found in the current directory", manifestFileName))
	}

	m, err := readManifest(manifestPath)
	handleError(err)

	existingLock, _ := readLockFile(filepath.Join(cwd, lockFileName))

	updateAll, updateNames, rest := parseUpdateFlag(args)

	if len(rest) > 2 {
		installOne(cwd, manifestPath, &m, rest, existingLock, updateAll || contains(updateNames, rest[2]))
		return
	}

	installAll(cwd, m, existingLock, updateAll, updateNames)
}

func parseUpdateFlag(args []string) (updateAll bool, names []string, rest []string) {
	rest = make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "--update":
			updateAll = true
		case strings.HasPrefix(args[i], "--update="):
			names = append(names, strings.TrimPrefix(args[i], "--update="))
		default:
			rest = append(rest, args[i])
		}
	}
	return
}

func contains(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}

// fetchAndVerify downloads depURL to a temp file, hashes it, and only
// moves it into destPath if either (a) there's no prior lock entry for
// this name, (b) allowUpdate is set, or (c) the hash matches what's
// already locked. On mismatch without allowUpdate, it leaves the existing
// modules/ file untouched and returns an error — install fails closed.
func fetchAndVerify(name, depURL, destPath string, existingLock map[string]lockEntry, allowUpdate bool) (lockEntry, error) {
	tmp, err := os.CreateTemp(vida.EmptyString, "vida-install-*")
	if err != nil {
		return lockEntry{}, err
	}
	tmpPath := tmp.Name()
	tmp.Close()
	defer os.Remove(tmpPath)

	if err := vida.DownloadCellTo(depURL, tmpPath); err != nil {
		return lockEntry{}, err
	}

	sum, err := sha256File(tmpPath)
	if err != nil {
		return lockEntry{}, err
	}

	if prev, locked := existingLock[name]; locked && !allowUpdate {
		if prev.URL == depURL && prev.SHA256 != sum {
			return lockEntry{}, fmt.Errorf(
				"%v: content at %v has changed since it was locked\n"+
					"\t    locked sha256: %v\n"+
					"\t    fetched sha256: %v\n"+
					"\trun 'vida install --update=%v' if this change is expected",
				name, depURL, prev.SHA256, sum, name,
			)
		}
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return lockEntry{}, err
	}

	if err := os.Rename(tmpPath, destPath); err != nil {
		return lockEntry{}, err
	}

	return lockEntry{URL: depURL, SHA256: sum}, nil
}

func installOne(cwd, manifestPath string, m *manifest, args []string, existingLock map[string]lockEntry, allowUpdate bool) {
	if len(args) < 4 {
		handleError(fmt.Errorf("usage: vida install <name> <url> [--update]"))
	}
	name, depURL := args[2], args[3]

	cellsDir := filepath.Join(cwd, vida.ProjectCellsDir, name+vida.VidaFileExtension)
	fmt.Printf("\t⬇  %-13v %v\n", name, depURL)

	entry, err := fetchAndVerify(name, depURL, cellsDir, existingLock, allowUpdate)
	handleError(err)

	if m.Dependencies == nil {
		m.Dependencies = make(map[string]string)
	}
	m.Dependencies[name] = depURL
	handleError(writeManifest(manifestPath, *m))

	if existingLock == nil {
		existingLock = make(map[string]lockEntry)
	}

	existingLock[name] = entry
	handleError(writeLockFile(filepath.Join(cwd, lockFileName), existingLock))

	fmt.Printf("\n\n\t✅ installed %v and updated %v\n\n\n\n", name, manifestFileName)
}

func installAll(cwd string, m manifest, existingLock map[string]lockEntry, updateAll bool, updateNames []string) {
	if len(m.Dependencies) == 0 {
		fmt.Printf("\tNo dependencies listed in %v.\n\n\n\n", manifestFileName)
		return
	}

	names := make([]string, 0, len(m.Dependencies))
	for name := range m.Dependencies {
		names = append(names, name)
	}
	sort.Strings(names)

	locked := make(map[string]lockEntry, len(names))
	failed := 0

	for _, name := range names {
		depURL := m.Dependencies[name]
		cellsDir := filepath.Join(cwd, vida.ProjectCellsDir, name+vida.VidaFileExtension)
		allow := updateAll || contains(updateNames, name)
		fmt.Printf("\t⬇  %-13v %v\n", name, depURL)

		entry, err := fetchAndVerify(name, depURL, cellsDir, existingLock, allow)
		if err != nil {
			fmt.Printf("\t❌ %v\n", err)
			failed++
			if prev, ok := existingLock[name]; ok {
				locked[name] = prev // keep the last-known-good lock entry, don't drop it
			}
			continue
		}
		locked[name] = entry
	}

	handleError(writeLockFile(filepath.Join(cwd, lockFileName), locked))

	fmt.Printf("\n\n\t%v installed, %v failed\n\n\n\n", len(locked)-failed, failed)
	if failed > 0 {
		os.Exit(1)
	}
}

func sha256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return vida.EmptyString, err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return vida.EmptyString, err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func writeLockFile(path string, locked map[string]lockEntry) error {
	names := make([]string, 0, len(locked))
	for name := range locked {
		names = append(names, name)
	}
	sort.Strings(names)

	var b strings.Builder
	b.WriteString("# generated by 'vida install' — do not edit by hand\n\n")
	for _, name := range names {
		entry := locked[name]
		fmt.Fprintf(&b, "[%v]\n", name)
		fmt.Fprintf(&b, "url = %q\n", entry.URL)
		fmt.Fprintf(&b, "sha256 = %q\n\n", entry.SHA256)
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func readLockFile(path string) (map[string]lockEntry, error) {
	locked := make(map[string]lockEntry)
	data, err := os.ReadFile(path)
	if err != nil {
		return locked, err // caller treats "missing file" as "no prior lock", not fatal
	}

	var current string
	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == vida.EmptyString || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			current = strings.TrimSuffix(strings.TrimPrefix(line, "["), "]")
			continue
		}
		if current == vida.EmptyString {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), `"`)

		entry := locked[current]
		switch key {
		case "url":
			entry.URL = value
		case "sha256":
			entry.SHA256 = value
		}
		locked[current] = entry
	}
	return locked, nil
}
