[中文](README.md) | [English](README.en-US.md)

# bbs-go

`bbs-go` is a lightweight community and Q&A platform for forums, knowledge bases, and discussions. It combines a Go backend with a React web frontend and provides articles, topics, comments, user relationships, content governance, permissions, and search.

## Project Links

- GitHub: <https://github.com/mlogclub/bbs-go>
- Gitee: <https://gitee.com/mlogclub/bbs-go>

## Feature Overview

![bbs-go feature overview](docs/images/features_en.svg)

### Member Experience

- Registration, login, profiles, and personal homepages
- Follow relationships, notifications, and interaction reminders
- Points, experience, levels, tasks, and badges

### Content

- Publish and edit topics, feeds, and articles
- Comments, replies, likes, and favorites
- Tags, categories, voting, and hidden content
- Built-in search

### Administration

- Manage users, topics, comments, and articles
- Reports, keyword filters, and content moderation
- Roles, permissions, menus, and system settings
- Operation logs and audit trails

## Quick Start with Docker Compose

MySQL:

```bash
curl -fsSL https://raw.githubusercontent.com/mlogclub/bbs-go/master/docker-compose.yml.example -o docker-compose.yml
docker compose up -d
```

PostgreSQL:

```bash
curl -fsSL https://raw.githubusercontent.com/mlogclub/bbs-go/master/docker-compose.postgresql.yml.example -o docker-compose.yml
docker compose up -d
```

Then open:

- Site: <http://localhost:3000>
- Admin dashboard: <http://localhost:3000/dashboard>
- Install wizard: <http://localhost:3000/install>

## Build and Deploy

GitHub Actions builds container images and pushes them to GHCR and Docker Hub. Server deployments can use `deploy/remote-deploy.sh` or Docker Compose directly.

```bash
git push origin master
```

Server deployment:

```bash
bash /opt/bbs-go/repo/deploy/remote-deploy.sh
```

The deployment script pulls the image, validates the Compose file, recreates the container, and runs a health check.

## Development and Tests

Backend tests:

```bash
go test ./...
```

Frontend checks:

```bash
cd web
pnpm typecheck
pnpm lint
pnpm build:spa
```

The local Agent MCP project is located in `local-agent-mcp/`. It is a Python stdio MCP service with article and article-comment tools.

Standalone download: [bbs-go-mcp-0.1.0.zip](local-agent-mcp/releases/bbs-go-mcp-0.1.0.zip).

Agent setup and login guide: [AGENT_GUIDE.md](local-agent-mcp/AGENT_GUIDE.md).

## License

bbs-go is licensed under the [GPLv3 License](LICENSE).
