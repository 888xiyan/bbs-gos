"use client"

import { CopyIcon } from "lucide-react"

import {
  DashboardDataPage,
  type DashboardDataPageConfig,
} from "@/components/dashboard/data"
import * as dashboardData from "@/components/dashboard/data/dashboard-data-route-utils"
import { Button } from "@/components/ui/button"
import { PERMISSIONS } from "@/lib/auth/permissions.generated"
import type { TFunction } from "@/lib/i18n"
import { useI18n } from "@/lib/i18n/provider"
import { msgError, msgSuccess } from "@/lib/toast"

function inviteCodeStatusCell(t: TFunction, value: unknown) {
  const status = String(value || "")
  if (status === "unused") return t("dashboard.inviteCodeStatus.unused")
  if (status === "used") return t("dashboard.inviteCodeStatus.used")
  if (status === "disabled") return t("dashboard.inviteCodeStatus.disabled")
  if (status === "expired") return t("dashboard.inviteCodeStatus.expired")
  return status || "-"
}

function copyInviteCode(
  code: unknown,
  successMessage: string,
  errorMessage: string
) {
  const value = String(code || "")
  if (!value) return
  void window.navigator.clipboard
    .writeText(value)
    .then(() => msgSuccess(successMessage))
    .catch(() => msgError(errorMessage))
}

export default function DashboardInviteCodesRoute() {
  const { t } = useI18n()
  const config: DashboardDataPageConfig = {
    title: dashboardData.title(t, "inviteCodes"),
    description: dashboardData.desc(t, "inviteCodes"),
    listEndpoint: "/api/admin/invite-code/list",
    viewPermission: PERMISSIONS.INVITE_CODE_VIEW,
    createEndpoint: "/api/admin/invite-code/create",
    createPermission: PERMISSIONS.INVITE_CODE_CREATE,
    detailEndpoint: (id) => `/api/admin/invite-code/${id}`,
    filters: [
      { name: "code", label: dashboardData.label(t, "inviteCode") },
      {
        name: "status",
        label: dashboardData.label(t, "status"),
        type: "select",
        options: [
          { label: t("dashboard.inviteCodeStatus.unused"), value: "unused" },
          { label: t("dashboard.inviteCodeStatus.used"), value: "used" },
          {
            label: t("dashboard.inviteCodeStatus.disabled"),
            value: "disabled",
          },
          {
            label: t("dashboard.inviteCodeStatus.expired"),
            value: "expired",
          },
        ],
      },
      {
        name: "createdByUserId",
        label: dashboardData.label(t, "createdByUserId"),
      },
      {
        name: "usedByUserId",
        label: dashboardData.label(t, "usedByUserId"),
      },
      {
        name: "usedStartTime",
        label: dashboardData.label(t, "usedStartTime"),
        type: "date",
      },
      {
        name: "usedEndTime",
        label: dashboardData.label(t, "usedEndTime"),
        type: "date",
      },
    ],
    columns: [
      { key: "id", label: dashboardData.label(t, "id") },
      {
        key: "code",
        label: dashboardData.label(t, "inviteCode"),
        render: (record) => (
          <div className="flex items-center gap-1">
            <span className="font-mono text-base tracking-[0.22em]">
              {String(record.code || "")}
            </span>
            <Button
              type="button"
              variant="ghost"
              size="icon-sm"
              onClick={() =>
                copyInviteCode(
                  record.code,
                  t("dashboard.messages.inviteCodeCopied"),
                  t("dashboard.errors.copyFailed")
                )
              }
            >
              <CopyIcon />
              <span className="sr-only">
                {t("dashboard.actions.copyInviteCode")}
              </span>
            </Button>
          </div>
        ),
      },
      {
        key: "status",
        label: dashboardData.label(t, "status"),
        render: (record) => inviteCodeStatusCell(t, record.status),
      },
      {
        key: "createdBy",
        label: dashboardData.label(t, "createdBy"),
        render: (record) => String(record.createdBy || "-"),
      },
      {
        key: "usedBy",
        label: dashboardData.label(t, "usedBy"),
        render: (record) => String(record.usedBy || "-"),
      },
      {
        key: "note",
        label: dashboardData.label(t, "note"),
        className: "min-w-40",
        render: (record) => String(record.note || "-"),
      },
      {
        key: "expiresAt",
        label: dashboardData.label(t, "expiresAt"),
        render: (record) => dashboardData.dateCell(record.expiresAt),
      },
      {
        key: "usedAt",
        label: dashboardData.label(t, "usedAt"),
        render: (record) =>
          Number(record.usedAt) > 0
            ? dashboardData.dateCell(record.usedAt)
            : "-",
      },
      {
        key: "createTime",
        label: dashboardData.label(t, "createTime"),
        render: (record) => dashboardData.dateCell(record.createTime),
      },
    ],
    formFields: [
      {
        name: "count",
        label: dashboardData.label(t, "count"),
        type: "number",
        required: true,
        min: 1,
        max: 100,
        defaultValue: 10,
      },
      {
        name: "note",
        label: dashboardData.label(t, "note"),
        type: "textarea",
      },
    ],
    detailFields: [
      { key: "id", label: dashboardData.label(t, "id") },
      {
        key: "code",
        label: dashboardData.label(t, "inviteCode"),
        render: (record) => (
          <div className="flex items-center gap-1">
            <span className="font-mono text-base tracking-[0.22em]">
              {String(record.code || "")}
            </span>
            <Button
              type="button"
              variant="ghost"
              size="icon-sm"
              onClick={() =>
                copyInviteCode(
                  record.code,
                  t("dashboard.messages.inviteCodeCopied"),
                  t("dashboard.errors.copyFailed")
                )
              }
            >
              <CopyIcon />
              <span className="sr-only">
                {t("dashboard.actions.copyInviteCode")}
              </span>
            </Button>
          </div>
        ),
      },
      {
        key: "status",
        label: dashboardData.label(t, "status"),
        render: (record) => inviteCodeStatusCell(t, record.status),
      },
      { key: "createdBy", label: dashboardData.label(t, "createdBy") },
      { key: "usedBy", label: dashboardData.label(t, "usedBy") },
      { key: "note", label: dashboardData.label(t, "note") },
      {
        key: "expiresAt",
        label: dashboardData.label(t, "expiresAt"),
        render: (record) => dashboardData.dateCell(record.expiresAt),
      },
      {
        key: "usedAt",
        label: dashboardData.label(t, "usedAt"),
        render: (record) =>
          Number(record.usedAt) > 0
            ? dashboardData.dateCell(record.usedAt)
            : "-",
      },
    ],
    rowActions: [
      {
        label: t("dashboard.actions.disableInviteCode"),
        endpoint: "/api/admin/invite-code/disable",
        permission: PERMISSIONS.INVITE_CODE_DISABLE,
        method: "POST",
        payload: (record) => ({ id: String(record.id ?? "") }),
        visible: (record) => record.status === "unused",
        confirm: t("dashboard.confirmDisableInviteCode"),
        successMessage: t("dashboard.messages.inviteCodeDisabled"),
      },
    ],
    pageSize: 20,
  }

  return <DashboardDataPage config={config} />
}
