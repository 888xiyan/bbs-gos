import { Navigate, useSearchParams } from "react-router"

import { SignupForm } from "@/components/auth/signup-form"
import { WidgetCard } from "@/components/common/widget-card"
import { useI18n } from "@/lib/i18n/provider"
import { noindexRouteMeta } from "@/lib/seo"
import { useDocumentTitle } from "@/lib/use-document-title"
import { useAppState } from "@/components/app/app-provider"
import { safeRedirect } from "@/lib/site"

export function meta({
  matches,
}: {
  matches: Array<{ data?: unknown; loaderData?: unknown }>
}) {
  return noindexRouteMeta(matches, "Sign up", "注册")
}

export default function SignupRoute() {
  const [searchParams] = useSearchParams()
  const { t } = useI18n()
  const { currentUser } = useAppState()
  useDocumentTitle(t("user.signup.title"))
  if (currentUser) {
    return (
      <Navigate
        to={safeRedirect(searchParams.get("redirect") || "/", "/")}
        replace
      />
    )
  }
  return (
    <section className="main">
      <div className="container mx-auto">
        <WidgetCard
          className="mx-auto max-w-[600px]"
          title={t("user.signup.title")}
          bodyClassName="pt-0"
        >
          <SignupForm redirect={searchParams.get("redirect") || undefined} />
        </WidgetCard>
      </div>
    </section>
  )
}
