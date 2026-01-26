# Repository Guidelines

## Project Structure & Module Organization
- `README.md`: 使用说明与环境搭建提示。
- `scripts/linux/`: 安装脚本；当前包含 `install-go.sh`（交互式下载并安装 Go）。
- `config/linux/`: Linux/WSL 配置（zsh + Powerlevel10k、LazyVim 模板）。Neovim入口 `config/linux/nvim/init.lua`，LazyVim 配置位于 `config/linux/nvim/lua/`.
- `config/windows/`: 预留的 Windows 配置目录（目前空）。
- `docs/`: 预留文档目录（目前空）。

## Build, Test, and Development Commands
- 目前无构建流程或测试套件。主要用途是同步 dotfiles 与脚本。
- 运行 Go 安装脚本（需要交互确认）：`bash scripts/linux/install-go.sh`
- 查看或调试 Neovim 配置：将 `config/linux/nvim` 链接到 `$HOME/.config/nvim` 后运行 `nvim`.

## Coding Style & Naming Conventions
- Shell 脚本：`bash` + `set -euo pipefail`，保持交互提示简洁，避免在命令前直接加 `sudo`（脚本内部已有）。
- Lua（Neovim）：遵循 LazyVim 默认风格，缩进 2 空格；格式化可用 `stylua`（参见 `config/linux/nvim/stylua.toml`）。
- 目录/文件命名：与平台对应，如 `config/linux/...`、`scripts/linux/...`，按功能分组。

## Testing Guidelines
- 当前无自动化测试。变更脚本时至少本地手动演练关键分支（如下载失败、用户取消、替换旧版本）。
- 如添加新脚本，建议附最小化干跑模式（如 `--dry-run`）或清晰的确认提示。

## Commit & Pull Request Guidelines
- 提交信息：使用简洁动词+范围，例如 `add go installer prompt`、`tweak zsh path order`。如涉及安全/破坏性操作，请在信息中点明。
- PR 要求（若有协作）：说明变更目的、受影响的目录（例如 `scripts/linux` 或 `config/linux/zsh`）、测试方式（手动步骤）、以及任何需要的后置操作（如重新 source `.zshrc`）。

## 安全检查要点
- 提交前用 `rg -n "SECRET|token|password|passwd|key"` 自查敏感字符串；关注 `~/.zshrc`、`.p10k.zsh` 等是否包含主机名/私服地址。
- 避免提交个人凭据、SSH 密钥、API token；路径中如含用户名/内网地址，请用占位符或删除。
- 下载来源优先官方 https，标注版本；覆盖性操作写清楚目标路径（如 `/usr/local/go`）并在脚本里确认。
- 涉及 `sudo` 或系统目录写操作时，提示风险与备份建议；必要时提供回滚/删除旧版本的步骤。
- 文档中说明后置动作（如 `source ~/.zshrc`、重启终端），减少误用。

## 脚本编写规范
- 统一使用 `#!/usr/bin/env bash` 与 `set -euo pipefail`；交互提示保持简洁，不在入口整体 `sudo`。
- 开始处检测依赖：`command -v curl wget tar sudo >/dev/null` 等，缺失则提示安装。
- 临时文件/目录用 `tmpdir=$(mktemp -d)`，并 `trap 'rm -rf \"$tmpdir\"' EXIT` 做清理；避免在仓库根或 `$HOME` 直接写入。
- 下载使用 `curl -fsSL` 或 `wget --quiet --show-progress`，必要时附官方 `sha256sum` 校验；版本号设为变量，便于复用。
- 破坏性操作（如 `rm -rf`、覆盖安装）前必须 `read -r -p "确认... [y/N]"`，默认拒绝；打印将要操作的路径。
- 需要提权的单条命令前加 `sudo`，而不是整脚本 `sudo bash ...`；在 WSL 与原生 Linux 保持路径显式。
- 结束时给出后置步骤和验证命令（如 `source ~/.zshrc`、`go version`），并使用合适的退出码。
