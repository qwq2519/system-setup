# 简介

个人开发环境配置与脚本集合，主要针对 WSL/Ubuntu 22.04 + zsh，方便快速搭建或恢复环境。Windows 端仅安装基础工具（如 scoop），开发依赖集中在 WSL。

## 目录结构
- `scripts/linux/`: 安装脚本，当前含 `install-go.sh`（交互式获取并安装最新 Go）。
- `config/linux/`: Linux/WSL 配置（zsh/Powerlevel10k，LazyVim Neovim 配置）。
- `config/windows/`: Windows 相关配置占位目录。
- `docs/`: 预留文档目录。

## 环境要求
- 推荐发行版：Ubuntu 22.04（WSL 或原生）。
- Shell：zsh（依赖 Oh My Zsh + Powerlevel10k）。
- 需要 `sudo` 权限执行部分安装步骤（脚本内部已处理）。

## 快速使用
1) 安装 Go  
```bash
bash scripts/linux/install-go.sh
```  
脚本会提示确认版本和覆盖安装；注意会 `sudo rm -rf /usr/local/go` 以替换旧版本。

2) 使用配置（软链接）  
- zsh：  
```bash
ln -sf $PWD/config/linux/zsh/.zshrc ~/.zshrc
ln -sf $PWD/config/linux/zsh/.p10k.zsh ~/.p10k.zsh
```
- Neovim：  
```bash
ln -sf $PWD/config/linux/nvim ~/.config/nvim
```
重新打开终端或 `source ~/.zshrc` 使配置生效。

## 配置说明
- zsh：插件包含 `git`、`zsh-autosuggestions`、`zsh-syntax-highlighting`、`history-substring-search`、`z`、`extract`、`copypath`；主题为 Powerlevel10k。
- Neovim：基于 LazyVim 模板，入口 `config/linux/nvim/init.lua`；`lua/config` 为可扩展点，`lua/plugins/example.lua` 默认返回空表，可复制改写以添加插件。格式化可参考 `stylua.toml`。

## 注意事项
- 不要整体 `sudo bash ...` 运行脚本，避免权限污染；脚本内部需要的 `sudo` 已显式调用。
- 提交前检查是否包含个人凭据、主机名或私钥；
