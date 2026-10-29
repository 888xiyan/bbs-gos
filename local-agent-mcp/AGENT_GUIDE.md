# bbs-go Agent MCP 使用指南

本指南面向本地 Agent。用户会把本链接转交给 Agent，Agent 按以下步骤下载模块并指导用户完成登录。

## 1. 获取模块

下载发布包：

```text
https://raw.githubusercontent.com/888xiyan/bbs-gos/master/local-agent-mcp/releases/bbs-go-mcp-0.1.0.zip
```

发布包包含 `bbs_go_mcp-0.1.0-py3-none-any.whl`、源码包、安装说明、配置示例和校验文件。

推荐下载后先校验 SHA-256。校验值文件位于：

```text
https://raw.githubusercontent.com/888xiyan/bbs-gos/master/local-agent-mcp/releases/SHA256SUMS
```

## 2. 安装 Python 模块

要求 Python 3.11 或更高版本。

Linux 和 macOS：

```bash
python3 -m venv .venv
. .venv/bin/activate
python -m pip install bbs_go_mcp-0.1.0-py3-none-any.whl
```

Windows PowerShell：

```powershell
py -3 -m venv .venv
.venv\Scripts\Activate.ps1
python -m pip install bbs_go_mcp-0.1.0-py3-none-any.whl
```

初始化配置：

```bash
bbs-go-mcp config init
```

默认目标站点是 `http://127.0.0.1:8082`。连接其他 bbs-go 站点时，编辑 `config.json` 的 `baseUrl`。

## 3. 指导用户登录

Agent 不进行浏览器自动化。请按以下方式指导用户完成登录：

1. 让用户打开目标 bbs-go 站点并完成登录。
2. 登录成功后，让用户从浏览器开发者工具的 Cookie 中复制 `bbsgo_token` 的值。
3. 执行本地导入命令，按提示粘贴 Cookie。终端输入不会显示：

```bash
bbs-go-mcp auth import-cookie
```

4. 验证登录状态：

```bash
bbs-go-mcp auth status
```

不要要求用户把 Cookie 发送给 Agent，也不要把 Cookie 写入聊天、日志或脚本。用户只需要在本地 `bbs-go-mcp` 命令中输入。

## 4. 启动 MCP

```bash
bbs-go-mcp run --stdio
```

MCP 客户端配置示例：

```json
{
  "mcpServers": {
    "bbs-go": {
      "command": "bbs-go-mcp",
      "args": ["run", "--stdio"]
    }
  }
}
```

## 5. Agent 可用工具

- `bbs_status`
- `bbs_auth_status`
- `bbs_list_categories`
- `bbs_get_article`
- `bbs_list_my_articles`
- `bbs_list_comments`
- `bbs_publish_article`
- `bbs_update_article`
- `bbs_set_comment_policy`
- `bbs_publish_comment`
- `bbs_delete_comment`
- `bbs_logout`

写入工具必须传入 `confirm=true`。`confirm=false` 只返回预检查结果，不会写入 bbs-go。

## 6. 登录引导文案

Agent 可以向用户展示下面的文案：

```text
请先登录 bbs-go 网站。登录成功后，在浏览器开发者工具的 Cookie 中复制 bbsgo_token 的值，然后在本机终端执行：

bbs-go-mcp auth import-cookie

请只在终端提示处粘贴 Cookie，不要把 Cookie 发到聊天窗口。完成后执行 bbs-go-mcp auth status 确认登录成功。
```

## 7. 错误处理

- `AUTH_REQUIRED`：用户未登录或 Cookie 已过期，重新执行 `auth import-cookie`。
- `CATEGORY_NOT_FOUND`：先调用 `bbs_list_categories`。
- `CATEGORY_NOT_POSTABLE`：换一个允许发布的栏目。
- `COMMENTS_DISABLED`：文章已关闭评论，停止评论操作。
- `CONFIRMATION_REQUIRED`：写入前必须明确传入 `confirm=true`。
- `NETWORK_ERROR`：检查 `baseUrl` 和本机网络。

完整设计和安全约束见 `docs/security.md`。
