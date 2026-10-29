# bbs-go 本地 Agent MCP Python 跨平台施工方案

## 1. 实施目标

本地 Agent 通过 Python MCP 服务访问 bbs-go，使用本地保存的 `bbsgo_token` Cookie 发布和维护文章、查询栏目、发布文章评论或回复。MCP 使用 stdio，不监听公网端口，不包含浏览器自动化、模型编排和管理员后台能力。

固定决策：

- Python 3.11+，包名 `bbs_go_mcp`，命令名 `bbs-go-mcp`。
- MCP 传输方式为 stdio。
- 写入工具默认 `confirm=false`，只有 `confirm=true` 才调用写接口。
- MCP 只封装文章和文章评论，不暴露 Topic、QA、Tweet。
- MCP 预检查只改善提示，认证、权限、栏目、评论开关、频控和审核由 bbs-go 最终校验。

## 2. 项目结构

```text
local-agent-mcp/
├── pyproject.toml
├── README.md
├── src/bbs_go_mcp/
│   ├── __init__.py
│   ├── __main__.py
│   ├── cli.py
│   ├── server.py
│   ├── config.py
│   ├── auth.py
│   ├── credential_store.py
│   ├── api_client.py
│   ├── category_resolver.py
│   ├── tool_service.py
│   ├── schemas.py
│   ├── errors.py
│   ├── logging.py
│   └── security.py
├── tests/
│   ├── test_config.py
│   ├── test_credentials.py
│   ├── test_api_client.py
│   ├── test_category_resolver.py
│   ├── test_tool_service.py
│   └── test_mcp_server.py
└── docs/
    ├── mcp-client-config.md
    └── security.md
```

## 3. 运行与配置

```bash
pipx install .
bbs-go-mcp config init
bbs-go-mcp auth import-cookie
bbs-go-mcp auth status
bbs-go-mcp run --stdio
```

配置文件不保存敏感信息：

```json
{
  "baseUrl": "http://127.0.0.1:8082",
  "timeoutSeconds": 30,
  "defaultConfirm": false,
  "allowedArticleUrlHosts": ["127.0.0.1:8082"],
  "maxContentLength": 200000
}
```

## 4. 凭据设计

- `auth import-cookie` 使用隐藏输入读取 Cookie，不接受命令行参数。
- 保存前调用 `/api/user/current` 验证。
- macOS 使用 Keychain，Windows 使用 Credential Manager，Linux 使用 Secret Service。
- 系统钥匙串不可用时使用 AES-256-GCM 加密文件。
- 主密钥和密文分离，Unix 文件权限为 `0600`，Windows 使用当前用户 ACL。
- 工具输出、错误和日志禁止包含 Cookie、Token、密码、完整正文或加密密钥。
- `bbs_logout` 调用 `/api/login/signout` 后删除本地凭据。

## 5. MCP 工具

| 工具 | 类型 | 说明 |
| --- | --- | --- |
| `bbs_status` | 只读 | 站点配置、登录状态和当前用户 |
| `bbs_auth_status` | 只读 | 只返回认证状态，不返回凭据 |
| `bbs_list_categories` | 只读 | 返回栏目树、完整路径、`categoryId`、`urlPath`、`allowPost` |
| `bbs_get_article` | 只读 | 支持 `articleId` 或同源 `articleUrl` |
| `bbs_list_my_articles` | 只读 | 查询当前用户文章 |
| `bbs_list_comments` | 只读 | 查询文章评论和回复 |
| `bbs_publish_article` | 写入 | 发布文章 |
| `bbs_update_article` | 写入 | 修改自己的文章 |
| `bbs_set_comment_policy` | 写入 | 设置文章 `allowComments` |
| `bbs_publish_comment` | 写入 | 发布文章评论或回复 |
| `bbs_delete_comment` | 破坏性 | 删除自己的评论 |
| `bbs_logout` | 破坏性 | 撤销服务端登录并清理本地凭据 |

写入工具统一只接受 `confirm=true` 执行。`confirm=false` 返回 `CONFIRMATION_REQUIRED` 和预检查结果。

## 6. 栏目与 URL 解析

1. 优先使用 `categoryId`。
2. 没有 ID 时按 `categoryPath` 精确匹配完整路径。
3. 匹配多个栏目返回 `CATEGORY_AMBIGUOUS`。
4. 栏目不存在返回 `CATEGORY_NOT_FOUND`。
5. 栏目停用或 `allowPost=false` 返回 `CATEGORY_NOT_POSTABLE`。
6. `articleUrl` 只允许解析配置中的 bbs-go 域名。
7. 禁止任意 URL、内网跳转、非 HTTP 协议和重定向。

## 7. API 映射

| MCP 能力 | bbs-go 接口 |
| --- | --- |
| 状态 | `GET /api/config/configs`、`GET /api/user/current` |
| 栏目树 | `GET /api/category/tree` |
| 文章创建 | `POST /api/article/create` |
| 文章编辑 | `POST /api/article/edit/:id` |
| 文章详情 | `GET /api/article/:id` |
| 我的文章 | `GET /api/article/user_articles` |
| 文章评论 | `GET /api/comment/comments` |
| 发表评论 | `POST /api/comment/create` |
| 删除评论 | `POST /api/comment/delete/:id` |
| 退出登录 | `GET /api/login/signout` |

请求使用：

```text
Cookie: bbsgo_token=<本地保存的值>
Content-Type: application/json
```

## 8. 文章与评论规则

- 文章必须选择有效栏目。
- 文章作者可以设置 `allowComments`。
- 评论只能挂在文章或文章内的回复下。
- 评论不接收栏目参数，栏目由所属文章继承。
- 文章关闭评论后，已有评论仍可读取，但不能新增评论或回复。
- 服务端负责所有权、权限、禁言、频控、违禁词和审核。

## 9. 错误模型

```json
{
  "success": false,
  "error": "COMMENTS_DISABLED",
  "message": "该文章作者已关闭评论",
  "retryable": false,
  "requestId": "..."
}
```

错误码：`AUTH_REQUIRED`、`PERMISSION_DENIED`、`CATEGORY_NOT_FOUND`、`CATEGORY_AMBIGUOUS`、`CATEGORY_NOT_POSTABLE`、`COMMENTS_DISABLED`、`ARTICLE_NOT_FOUND`、`VALIDATION_ERROR`、`RATE_LIMITED`、`CONFIRMATION_REQUIRED`、`NETWORK_ERROR`。

## 10. 跨平台要求

- 使用 `platformdirs` 计算 Linux、Windows、macOS 的配置目录。
- 文件写入使用临时文件和原子替换。
- macOS 使用 Keychain，Windows 使用 Credential Manager 和 ACL，Linux 使用 Secret Service。
- CI 至少覆盖 `ubuntu-latest`、`windows-latest`、`macos-latest`，Python 3.11、3.12、3.13。
- `doctor` 命令检查 Python 版本、配置、存储可用性、目标站点和权限。

## 11. 测试计划

- 单元测试：配置、凭据、错误转换、栏目解析、URL 白名单、confirm 机制。
- API Mock：Cookie 认证、401、403、404、429、超时、JSON 错误。
- 集成测试：真实 bbs-go 的文章创建、编辑、评论和评论开关。
- MCP 冒烟测试：`initialize`、`tools/list`、`tools/call`。
- 安全测试：扫描工具输出和日志，确认无 Cookie、Token、密码和完整正文。

## 12. 交付顺序

1. 建立 Python 包、配置、错误模型和 API 客户端。
2. 完成系统钥匙串和加密文件凭据存储。
3. 接入 MCP stdio 与只读工具。
4. 接入文章工具和 confirm 预检。
5. 接入评论工具和服务端错误映射。
6. 完成安全审计、三平台测试、安装文档和发布包。

预计工作量：`8 至 11 人日`。
