"use client"

import {
  DashboardDataPage,
  type DashboardDataPageConfig,
} from "@/components/dashboard/data"
import * as dashboardData from "@/components/dashboard/data/dashboard-data-route-utils"
import { PERMISSIONS } from "@/lib/auth/permissions.generated"
import { useI18n } from "@/lib/i18n/provider"

export default function DashboardAgentsRoute() {
  const { t } = useI18n()
  const config: DashboardDataPageConfig = {
    title: t("dashboard.pages.agents.title"),
    description: t("dashboard.pages.agents.description"),
    listEndpoint: "/api/admin/agent/list",
    viewPermission: PERMISSIONS.AGENT_ADMIN_VIEW,
    listResult: "array",
    detailEndpoint: (id) => `/api/admin/agent/${id}`,
    filters: [
      { name: "name", label: dashboardData.label(t, "name") },
      { name: "ownerUserId", label: dashboardData.label(t, "userId") },
      {
        name: "status",
        label: dashboardData.label(t, "status"),
        type: "select",
        options: [
          { label: t("dashboard.agentStatus.active"), value: "active" },
          { label: t("dashboard.agentStatus.disabled"), value: "disabled" },
          { label: t("dashboard.agentStatus.revoked"), value: "revoked" },
        ],
      },
    ],
    columns: [
      { key: "id", label: dashboardData.label(t, "id") },
      { key: "name", label: dashboardData.label(t, "name") },
      { key: "ownerUserId", label: dashboardData.label(t, "userId") },
      { key: "owner", label: dashboardData.label(t, "user") },
      { key: "tokenCount", label: t("dashboard.fields.tokenCount") },
      { key: "status", label: dashboardData.label(t, "status") },
      {
        key: "lastActiveTime",
        label: t("dashboard.fields.lastActiveTime"),
        render: (record) => dashboardData.dateCell(record.lastActiveTime),
      },
    ],
    rowActions: [
      {
        label: t("dashboard.actions.disableAgent"),
        endpoint: "/api/admin/agent/disable",
        permission: PERMISSIONS.AGENT_ADMIN_DISABLE,
        method: "POST",
        payload: (record) => ({ id: String(record.id ?? "") }),
        visible: (record) => record.status === "active",
        confirm: t("dashboard.confirmDisableAgent"),
        successMessage: t("dashboard.messages.agentDisabled"),
      },
    ],
    pageSize: 20,
  }

  return <DashboardDataPage config={config} />
}
