"use client"

import {
  DashboardDataPage,
  type DashboardDataPageConfig,
} from "@/components/dashboard/data"
import * as dashboardData from "@/components/dashboard/data/dashboard-data-route-utils"
import { PERMISSIONS } from "@/lib/auth/permissions.generated"
import { useI18n } from "@/lib/i18n/provider"

export default function DashboardAgentTokensRoute() {
  const { t } = useI18n()
  const config: DashboardDataPageConfig = {
    title: t("dashboard.pages.agentTokens.title"),
    description: t("dashboard.pages.agentTokens.description"),
    listEndpoint: "/api/admin/agent-token/list",
    viewPermission: PERMISSIONS.AGENT_ADMIN_VIEW,
    listResult: "array",
    filters: [
      { name: "agentId", label: dashboardData.label(t, "agentId") },
      {
        name: "status",
        label: dashboardData.label(t, "status"),
        type: "select",
        options: [
          { label: t("dashboard.agentStatus.active"), value: "active" },
          { label: t("dashboard.agentStatus.revoked"), value: "revoked" },
        ],
      },
    ],
    columns: [
      { key: "id", label: dashboardData.label(t, "id") },
      { key: "agentId", label: dashboardData.label(t, "agentId") },
      { key: "agentName", label: t("dashboard.fields.agentName") },
      { key: "ownerUserId", label: dashboardData.label(t, "userId") },
      {
        key: "tokenPrefix",
        label: t("dashboard.fields.tokenPrefix"),
        render: (record) => `${String(record.tokenPrefix || "")}...`,
      },
      { key: "status", label: dashboardData.label(t, "status") },
      {
        key: "expiredAt",
        label: dashboardData.label(t, "expiresAt"),
        render: (record) => dashboardData.dateCell(record.expiredAt),
      },
      {
        key: "lastUsedTime",
        label: t("dashboard.fields.lastUsedTime"),
        render: (record) => dashboardData.dateCell(record.lastUsedTime),
      },
    ],
    rowActions: [
      {
        label: t("dashboard.actions.revokeAgentToken"),
        endpoint: "/api/admin/agent-token/revoke",
        permission: PERMISSIONS.AGENT_ADMIN_DISABLE,
        method: "POST",
        payload: (record) => ({ id: String(record.id ?? "") }),
        visible: (record) => record.status === "active",
        confirm: t("dashboard.confirmRevokeAgentToken"),
        successMessage: t("dashboard.messages.agentTokenRevoked"),
      },
    ],
    pageSize: 20,
  }

  return <DashboardDataPage config={config} />
}
