package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	actionList   = "list"
	actionStatus = "status"
	actionPull   = "pull"

	stateSynced        = "synced"
	stateDifferent     = "different"
	stateMissingSource = "missing-source"
	stateMissingTarget = "missing-target"
	stateMissingBoth   = "missing-both"
)

var currentGOOS = runtime.GOOS

type config struct {
	Root     string
	Mappings []mapping
}

type mapping struct {
	Source string
	Target string
}

type fileStatus struct {
	Mapping mapping
	State   string
	Detail  string
}

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	fs := flag.NewFlagSet("winsync", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	action := fs.String("action", "", "list | status | pull")
	mappingPath := fs.String("mapping", "mapping.txt", "mapping file path")
	selectedPath := fs.String("path", "", "status 可用 source/target，pull 只能用 source")
	fs.Usage = func() {
		printUsage(os.Stderr)
	}

	if err := fs.Parse(args); err != nil {
		return 1
	}
	if fs.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "不接受额外位置参数，请使用 flag")
		printUsage(os.Stderr)
		return 1
	}
	if *action == "" {
		fmt.Fprintln(os.Stderr, "必须指定 -action")
		printUsage(os.Stderr)
		return 1
	}
	if err := requireWindows(); err != nil {
		fmt.Fprintf(os.Stderr, "运行失败: %v\n", err)
		return 1
	}

	cfg, err := loadMappings(*mappingPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载 mapping 失败: %v\n", err)
		return 1
	}

	switch *action {
	case actionList:
		return runList(cfg)
	case actionStatus:
		return runStatus(cfg, *selectedPath)
	case actionPull:
		return runPull(cfg, *selectedPath)
	default:
		fmt.Fprintf(os.Stderr, "未知 action: %s\n", *action)
		printUsage(os.Stderr)
		return 1
	}
}

func runList(cfg config) int {
	if len(cfg.Mappings) == 0 {
		fmt.Println("mapping 为空")
		return 0
	}

	for _, item := range cfg.Mappings {
		fmt.Printf("%s -> %s\n", item.Source, item.Target)
	}
	return 0
}

func runStatus(cfg config, selectedPath string) int {
	statuses, err := collectStatuses(cfg, selectedPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "检查状态失败: %v\n", err)
		return 1
	}
	if len(statuses) == 0 {
		fmt.Println("mapping 为空")
		return 0
	}

	hasChanges := false
	for _, item := range statuses {
		fmt.Printf("%s | %s -> %s | %s\n", item.State, item.Mapping.Source, item.Mapping.Target, item.Detail)
		if item.State != stateSynced {
			hasChanges = true
		}
	}
	if hasChanges {
		return 2
	}
	return 0
}

func runPull(cfg config, selectedPath string) int {
	if selectedPath == "" {
		fmt.Fprintln(os.Stderr, "pull 必须指定 -path，且内容要与 mapping 中某一行的 Windows 源文件路径完全一致")
		return 1
	}

	changed, err := pullOne(cfg, selectedPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "拉取失败: %v\n", err)
		return 1
	}
	if changed {
		fmt.Printf("已更新: %s\n", selectedPath)
		return 0
	}

	fmt.Printf("无需更新: %s\n", selectedPath)
	return 0
}

func requireWindows() error {
	if currentGOOS != "windows" {
		return fmt.Errorf("winsync 只能在 Windows 下运行，当前系统是 %s", currentGOOS)
	}
	return nil
}

func loadMappings(mappingPath string) (config, error) {
	absMappingPath, err := filepath.Abs(mappingPath)
	if err != nil {
		return config{}, err
	}

	content, err := os.ReadFile(absMappingPath)
	if err != nil {
		return config{}, err
	}

	root := filepath.Dir(absMappingPath)
	lines := strings.Split(string(content), "\n")
	seen := map[string]bool{}
	mappings := make([]mapping, 0, len(lines))

	for index, raw := range lines {
		line := strings.TrimSpace(stripComment(raw))
		if line == "" {
			continue
		}

		source, target, err := parseMappingLine(line)
		if err != nil {
			return config{}, fmt.Errorf("mapping.txt 第 %d 行: %w", index+1, err)
		}
		if err := validateSource(source); err != nil {
			return config{}, fmt.Errorf("mapping.txt 第 %d 行: %w", index+1, err)
		}
		if _, err := resolveTarget(root, target); err != nil {
			return config{}, fmt.Errorf("mapping.txt 第 %d 行: %w", index+1, err)
		}
		if seen[source] || seen[target] {
			return config{}, fmt.Errorf("mapping.txt 第 %d 行: source 或 target 存在重复值", index+1)
		}
		seen[source] = true
		seen[target] = true

		mappings = append(mappings, mapping{
			Source: source,
			Target: target,
		})
	}

	return config{
		Root:     root,
		Mappings: mappings,
	}, nil
}

func collectStatuses(cfg config, selectedPath string) ([]fileStatus, error) {
	selected, err := selectMappings(cfg.Mappings, selectedPath)
	if err != nil {
		return nil, err
	}

	statuses := make([]fileStatus, 0, len(selected))
	for _, item := range selected {
		status, err := checkOne(cfg.Root, item)
		if err != nil {
			return nil, err
		}
		statuses = append(statuses, status)
	}
	return statuses, nil
}

func pullOne(cfg config, selectedPath string) (bool, error) {
	item, err := selectOneBySource(cfg.Mappings, selectedPath)
	if err != nil {
		return false, err
	}

	status, err := checkOne(cfg.Root, item)
	if err != nil {
		return false, err
	}

	switch status.State {
	case stateSynced:
		return false, nil
	case stateDifferent, stateMissingTarget:
		targetPath, err := resolveTarget(cfg.Root, item.Target)
		if err != nil {
			return false, err
		}
		if err := copyFile(item.Source, targetPath); err != nil {
			return false, err
		}
		return true, nil
	case stateMissingSource, stateMissingBoth:
		return false, fmt.Errorf("源文件不存在: %s", item.Source)
	default:
		return false, fmt.Errorf("不支持的状态: %s", status.State)
	}
}

func checkOne(root string, item mapping) (fileStatus, error) {
	targetPath, err := resolveTarget(root, item.Target)
	if err != nil {
		return fileStatus{}, err
	}

	sourceInfo, sourceErr := os.Stat(item.Source)
	targetInfo, targetErr := os.Stat(targetPath)

	switch {
	case sourceErr == nil && sourceInfo.IsDir():
		return fileStatus{Mapping: item, State: stateMissingSource, Detail: "源路径是目录"}, nil
	case targetErr == nil && targetInfo.IsDir():
		return fileStatus{Mapping: item, State: stateMissingTarget, Detail: "目标路径是目录"}, nil
	case sourceErr == nil && targetErr == nil:
		same, err := sameContent(item.Source, targetPath)
		if err != nil {
			return fileStatus{}, err
		}
		if same {
			return fileStatus{Mapping: item, State: stateSynced, Detail: "内容一致"}, nil
		}
		return fileStatus{Mapping: item, State: stateDifferent, Detail: "文件内容不同"}, nil
	case errors.Is(sourceErr, os.ErrNotExist) && errors.Is(targetErr, os.ErrNotExist):
		return fileStatus{Mapping: item, State: stateMissingBoth, Detail: "源文件和目标文件都不存在"}, nil
	case errors.Is(sourceErr, os.ErrNotExist):
		return fileStatus{Mapping: item, State: stateMissingSource, Detail: "源文件不存在"}, nil
	case errors.Is(targetErr, os.ErrNotExist):
		return fileStatus{Mapping: item, State: stateMissingTarget, Detail: "目标文件不存在"}, nil
	case sourceErr != nil:
		return fileStatus{}, sourceErr
	default:
		return fileStatus{}, targetErr
	}
}

func selectMappings(mappings []mapping, selectedPath string) ([]mapping, error) {
	if selectedPath == "" {
		return mappings, nil
	}

	item, err := selectOne(mappings, selectedPath)
	if err != nil {
		return nil, err
	}
	return []mapping{item}, nil
}

func selectOne(mappings []mapping, selectedPath string) (mapping, error) {
	for _, item := range mappings {
		if item.Source == selectedPath || item.Target == selectedPath {
			return item, nil
		}
	}
	return mapping{}, fmt.Errorf("path 未在 mapping 中找到完全一致的条目: %s", selectedPath)
}

func selectOneBySource(mappings []mapping, sourcePath string) (mapping, error) {
	for _, item := range mappings {
		if item.Source == sourcePath {
			return item, nil
		}
	}
	return mapping{}, fmt.Errorf("pull 的 -path 必须是 mapping 中某一行的 Windows 源文件路径: %s", sourcePath)
}

func parseMappingLine(line string) (string, string, error) {
	parts := strings.SplitN(line, "->", 2)
	if len(parts) != 2 {
		return "", "", errors.New("格式必须是 <source> -> <target>")
	}

	source := strings.TrimSpace(parts[0])
	target := strings.TrimSpace(parts[1])
	if source == "" || target == "" {
		return "", "", errors.New("source 和 target 不能为空")
	}
	return source, target, nil
}

func validateSource(source string) error {
	normalized := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(source), "\\", "/"))
	if strings.HasPrefix(normalized, "/mnt/") {
		return fmt.Errorf("source 不支持 WSL 路径，请使用 Windows 路径: %s", source)
	}
	return nil
}

func resolveTarget(root string, target string) (string, error) {
	if filepath.IsAbs(target) {
		return "", fmt.Errorf("target 必须是仓库内相对路径: %s", target)
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	absTarget, err := filepath.Abs(filepath.Join(absRoot, target))
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(absRoot, absTarget)
	if err != nil {
		return "", err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("target 超出仓库根目录: %s", target)
	}
	return absTarget, nil
}

func sameContent(sourcePath string, targetPath string) (bool, error) {
	sourceContent, err := os.ReadFile(sourcePath)
	if err != nil {
		return false, err
	}
	targetContent, err := os.ReadFile(targetPath)
	if err != nil {
		return false, err
	}
	return bytes.Equal(sourceContent, targetContent), nil
}

func copyFile(sourcePath string, targetPath string) error {
	content, err := os.ReadFile(sourcePath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return err
	}

	mode := os.FileMode(0o644)
	if info, err := os.Stat(targetPath); err == nil {
		mode = info.Mode().Perm()
	}
	return os.WriteFile(targetPath, content, mode)
}

func stripComment(line string) string {
	if index := strings.Index(line, "#"); index >= 0 {
		return line[:index]
	}
	return line
}

func printUsage(out *os.File) {
	fmt.Fprintln(out, "用法:")
	fmt.Fprintln(out, "  go run ./cmd/winsync -action list [-mapping mapping.txt]")
	fmt.Fprintln(out, "  go run ./cmd/winsync -action status [-mapping mapping.txt] [-path 精确路径]")
	fmt.Fprintln(out, "  go run ./cmd/winsync -action pull [-mapping mapping.txt] -path Windows源文件路径")
	fmt.Fprintln(out)
	fmt.Fprintln(out, "status 的 -path 可以是 source 或 target；pull 的 -path 只能是 source。")
}
