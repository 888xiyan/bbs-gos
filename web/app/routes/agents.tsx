"use client"

import * as React from "react"

import { RequireUser } from "@/components/auth/require-user"
import { WidgetCard } from "@/components/common/widget-card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { useCurrentUser } from "@/components/app/app-provider"
import { apiFetch } from "@/lib/api/client"
import { useI18n } from "@/lib/i18n/provider"
import { msgError, msgSuccess } from "@/lib/toast"
import { useDocumentTitle } from "@/lib/use-document-title"
import { requireUser, requireUserClient } from "../route-helpers/auth"

type Agent = {
  id: number
  name: string
  description?: string
  status: "active" | "disabled" | "revoked" | string
  lastActiveTime?: number
}

type AgentToken = {
  id: number
  tokenPrefix: string
  status: string
  expiredAt?: number
  lastUsedTime?: number
}

export async function loader(args: { request: Request }) {
  await requireUser(args)
  return null
}

export async function clientLoader(args: { request: Request }) {
  await requireUserClient(args)
  return null
}

export default function AgentsRoute() {
  const user = useCurrentUser()
  return (
    <RequireUser initialUser={user} redirectPath="/agents">
      <AgentsPage />
    </RequireUser>
  )
}

function AgentsPage() {
  const { t } = useI18n()
  useDocumentTitle(t("pages.agents.title"))
  const [agents, setAgents] = React.useState<Agent[]>([])
  const [name, setName] = React.useState("")
  const [description, setDescription] = React.useState("")
  const [tokens, setTokens] = React.useState<Record<number, AgentToken[]>>({})
  const [issuedToken, setIssuedToken] = React.useState("")
  const [loading, setLoading] = React.useState(true)

  const load = React.useCallback(async () => {
    setLoading(true)
    try {
      setAgents(await apiFetch<Agent[]>("/api/agents"))
    } catch (error) {
      msgError(
        error instanceof Error
          ? error.message
          : t("dashboard.errors.loadFailed")
      )
    } finally {
      setLoading(false)
    }
  }, [t])

  React.useEffect(() => {
    void load()
  }, [load])

  async function createAgent() {
    try {
      await apiFetch<Agent>("/api/agents", {
        method: "POST",
        body: { name: name.trim(), description: description.trim() },
      })
      setName("")
      setDescription("")
      await load()
      msgSuccess(t("pages.agents.created"))
    } catch (error) {
      msgError(
        error instanceof Error
          ? error.message
          : t("dashboard.errors.saveFailed")
      )
    }
  }

  async function loadTokens(agentId: number) {
    try {
      setTokens((current) => ({ ...current, [agentId]: [] }))
      const values = await apiFetch<AgentToken[]>(
        `/api/agents/${agentId}/tokens`
      )
      setTokens((current) => ({ ...current, [agentId]: values }))
    } catch (error) {
      msgError(
        error instanceof Error
          ? error.message
          : t("dashboard.errors.loadFailed")
      )
    }
  }

  async function issueToken(agentId: number) {
    try {
      const result = await apiFetch<AgentToken & { token: string }>(
        `/api/agents/${agentId}/tokens`,
        { method: "POST" }
      )
      setIssuedToken(result.token)
      await loadTokens(agentId)
      msgSuccess(t("pages.agents.tokenIssued"))
    } catch (error) {
      msgError(
        error instanceof Error
          ? error.message
          : t("dashboard.errors.saveFailed")
      )
    }
  }

  async function revokeToken(agentId: number, tokenId: number) {
    try {
      await apiFetch<void>(`/api/agents/${agentId}/tokens/${tokenId}/revoke`, {
        method: "POST",
      })
      await loadTokens(agentId)
      msgSuccess(t("pages.agents.tokenRevoked"))
    } catch (error) {
      msgError(
        error instanceof Error
          ? error.message
          : t("dashboard.errors.actionFailed")
      )
    }
  }

  async function disableAgent(agentId: number) {
    try {
      await apiFetch<void>(`/api/agents/${agentId}`, { method: "DELETE" })
      await load()
      msgSuccess(t("pages.agents.disabled"))
    } catch (error) {
      msgError(
        error instanceof Error
          ? error.message
          : t("dashboard.errors.actionFailed")
      )
    }
  }

  return (
    <main className="main">
      <div className="container space-y-4">
        <WidgetCard title={t("pages.agents.title")}>
          <div className="grid gap-3">
            <Label htmlFor="agent-name">{t("pages.agents.name")}</Label>
            <Input
              id="agent-name"
              value={name}
              onChange={(event) => setName(event.target.value)}
            />
            <Label htmlFor="agent-description">
              {t("pages.agents.description")}
            </Label>
            <Input
              id="agent-description"
              value={description}
              onChange={(event) => setDescription(event.target.value)}
            />
            <Button onClick={() => void createAgent()} disabled={!name.trim()}>
              {t("pages.agents.create")}
            </Button>
          </div>
        </WidgetCard>

        {issuedToken ? (
          <WidgetCard title={t("pages.agents.tokenOnce")}>
            <code className="block rounded-md bg-muted p-3 text-sm break-all">
              {issuedToken}
            </code>
          </WidgetCard>
        ) : null}

        {loading
          ? null
          : agents.map((agent) => (
              <WidgetCard key={agent.id} title={agent.name}>
                <div className="flex flex-wrap items-center gap-2">
                  <span>{agent.status}</span>
                  <Button
                    variant="outline"
                    onClick={() => void loadTokens(agent.id)}
                  >
                    {t("pages.agents.tokens")}
                  </Button>
                  <Button
                    variant="outline"
                    onClick={() => void issueToken(agent.id)}
                  >
                    {t("pages.agents.issueToken")}
                  </Button>
                  {agent.status === "active" ? (
                    <Button
                      variant="destructive"
                      onClick={() => void disableAgent(agent.id)}
                    >
                      {t("pages.agents.disable")}
                    </Button>
                  ) : null}
                </div>
                {tokens[agent.id]?.length ? (
                  <ul className="mt-3 divide-y">
                    {tokens[agent.id].map((token) => (
                      <li
                        key={token.id}
                        className="flex items-center justify-between py-2"
                      >
                        <span className="font-mono">
                          {token.tokenPrefix}...
                        </span>
                        <span>{token.status}</span>
                        {token.status === "active" ? (
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => void revokeToken(agent.id, token.id)}
                          >
                            {t("pages.agents.revokeToken")}
                          </Button>
                        ) : null}
                      </li>
                    ))}
                  </ul>
                ) : null}
              </WidgetCard>
            ))}
      </div>
    </main>
  )
}
