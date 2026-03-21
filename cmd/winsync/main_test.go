package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadMappingsIgnoresComments(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mappingPath := filepath.Join(root, "mapping.txt")
	content := "# comment only\n\nC:\\Users\\alice\\settings.json -> config/windows/vscode/settings.json # inline comment\n"
	if err := os.WriteFile(mappingPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write mapping: %v", err)
	}

	cfg, err := loadMappings(mappingPath)
	if err != nil {
		t.Fatalf("load mappings: %v", err)
	}

	if len(cfg.Mappings) != 1 {
		t.Fatalf("expected 1 mapping, got %d", len(cfg.Mappings))
	}
	if cfg.Mappings[0].Source != `C:\Users\alice\settings.json` {
		t.Fatalf("unexpected source: %s", cfg.Mappings[0].Source)
	}
}

func TestLoadMappingsRejectsWSLPath(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mappingPath := filepath.Join(root, "mapping.txt")
	content := "/mnt/c/Users/alice/settings.json -> config/windows/vscode/settings.json\n"
	if err := os.WriteFile(mappingPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write mapping: %v", err)
	}

	_, err := loadMappings(mappingPath)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "不支持 WSL 路径") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStatusAndPull(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	sourcePath := filepath.Join(root, "windows-source.json")
	targetPath := filepath.Join(root, "config", "windows", "vscode", "settings.json")
	mappingPath := filepath.Join(root, "mapping.txt")

	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		t.Fatalf("mkdir target: %v", err)
	}
	if err := os.WriteFile(sourcePath, []byte("{\"fontSize\":14}\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	if err := os.WriteFile(targetPath, []byte("{\"fontSize\":12}\n"), 0o644); err != nil {
		t.Fatalf("write target: %v", err)
	}
	line := sourcePath + " -> config/windows/vscode/settings.json\n"
	if err := os.WriteFile(mappingPath, []byte(line), 0o644); err != nil {
		t.Fatalf("write mapping: %v", err)
	}

	cfg, err := loadMappings(mappingPath)
	if err != nil {
		t.Fatalf("load mappings: %v", err)
	}

	statuses, err := collectStatuses(cfg, "")
	if err != nil {
		t.Fatalf("collect statuses: %v", err)
	}
	if len(statuses) != 1 {
		t.Fatalf("expected 1 status, got %d", len(statuses))
	}
	if statuses[0].State != stateDifferent {
		t.Fatalf("expected %s, got %s", stateDifferent, statuses[0].State)
	}

	changed, err := pullOne(cfg, sourcePath)
	if err != nil {
		t.Fatalf("pull: %v", err)
	}
	if !changed {
		t.Fatal("expected changed")
	}

	content, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if string(content) != "{\"fontSize\":14}\n" {
		t.Fatalf("unexpected content: %s", string(content))
	}
}

func TestPullRejectsTargetPath(t *testing.T) {
	t.Parallel()

	cfg := config{
		Mappings: []mapping{
			{
				Source: `C:\Users\alice\AppData\Roaming\Code\User\settings.json`,
				Target: "config/windows/vscode/settings.json",
			},
		},
	}

	_, err := pullOne(cfg, "config/windows/vscode/settings.json")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "Windows 源文件路径") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRequireWindows(t *testing.T) {
	previous := currentGOOS
	currentGOOS = "linux"
	t.Cleanup(func() {
		currentGOOS = previous
	})

	if err := requireWindows(); err == nil {
		t.Fatal("expected windows-only error")
	}
}
