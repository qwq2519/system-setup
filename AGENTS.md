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

## Security & Configuration Tips
- 切勿提交个人凭据、SSH 密钥、API token 等；检查新加入的配置文件是否含有用户路径、主机名或私有仓库地址。
- `install-go.sh` 会执行 `sudo rm -rf /usr/local/go` 覆盖旧版本；提醒用户先确认路径与权限，必要时备份旧安装。
