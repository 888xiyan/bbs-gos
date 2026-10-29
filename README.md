[中文](README.md) | [English](README.en-US.md)

# bbs-go

`bbs-go` 是一个轻量级社区和问答平台，适合搭建论坛、知识库和讨论社区。项目使用 Go 后端与 React Web 前端，支持文章、帖子、评论、用户关系、内容治理、权限管理和站内搜索。

## 项目链接

- GitHub：<https://github.com/mlogclub/bbs-go>
- Gitee：<https://gitee.com/mlogclub/bbs-go>

## 功能概览

![bbs-go 功能概览](docs/images/features_zh.svg)

### 用户

- 注册、登录、用户资料和个人主页
- 关注、粉丝、站内消息和互动提醒
- 积分、经验、等级、任务和勋章

### 内容

- 帖子、动态、文章的发布和编辑
- 评论、回复、点赞、收藏
- 标签、栏目、投票和隐藏内容
- 站内搜索

### 管理

- 用户、帖子、评论和文章治理
- 举报处理、违禁词和内容审核
- 角色、权限、菜单和系统参数配置
- 操作日志和行为记录

## Docker Compose 快速开始

MySQL：

```bash
curl -fsSL https://raw.githubusercontent.com/mlogclub/bbs-go/master/docker-compose.yml.example -o docker-compose.yml
docker compose up -d
```

PostgreSQL：

```bash
curl -fsSL https://raw.githubusercontent.com/mlogclub/bbs-go/master/docker-compose.postgresql.yml.example -o docker-compose.yml
docker compose up -d
```

启动后访问：

- 前台：<http://localhost:3000>
- 后台：<http://localhost:3000/dashboard>
- 安装向导：<http://localhost:3000/install>

## 构建与部署

项目使用 GitHub Actions 构建容器镜像，并推送到 GHCR 和 Docker Hub。服务器部署可以使用 `deploy/remote-deploy.sh`，也可以直接使用 Docker Compose 管理服务。

```bash
git push origin master
```

服务器部署：

```bash
bash /opt/bbs-go/repo/deploy/remote-deploy.sh
```

部署脚本会拉取镜像、校验 Compose 配置、重建容器并执行健康检查。

## 开发与测试

后端测试：

```bash
go test ./...
```

前端检查：

```bash
cd web
pnpm typecheck
pnpm lint
pnpm build:spa
```

本地 Agent MCP 项目位于 `local-agent-mcp/`，使用 Python 实现 stdio MCP 服务，支持文章和文章评论工具。

可直接下载的模块：[bbs-go-mcp-0.1.0.zip](local-agent-mcp/releases/bbs-go-mcp-0.1.0.zip)。

交给 Agent 的安装和登录指南：[AGENT_GUIDE.md](local-agent-mcp/AGENT_GUIDE.md)。

## 许可证

bbs-go 采用 [GPLv3 License](LICENSE)。
