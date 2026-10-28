import { Navigate, redirect } from "react-router"
import { apiFetch } from "@/lib/api/client"
import type { Category } from "@/lib/api/types"
import { noindexRouteMeta } from "@/lib/seo"

export function meta({
  matches,
}: {
  matches: Array<{ data?: unknown; loaderData?: unknown }>
}) {
  return noindexRouteMeta(matches, "Not found", "页面不存在")
}

export async function loader({
  request,
}: {
  request: Request
}): Promise<never> {
  const url = new URL(request.url)
  const category = await apiFetch<Category>("/api/category/by_path", {
    request,
    params: { urlPath: url.pathname },
  }).catch(() => null)
  if (category?.id) {
    if (category.urlPath && category.urlPath !== url.pathname) {
      throw redirect(category.urlPath)
    }
    throw redirect(`/articles/category/${category.id}`)
  }
  throw redirect("/user/signin?redirect=/")
}

export async function clientLoader(): Promise<never> {
  const category = await apiFetch<Category>("/api/category/by_path", {
    params: { urlPath: window.location.pathname },
  }).catch(() => null)
  if (category?.id) {
    if (category.urlPath && category.urlPath !== window.location.pathname) {
      throw redirect(category.urlPath)
    }
    throw redirect(`/articles/category/${category.id}`)
  }
  throw redirect("/user/signin?redirect=/")
}

export default function NotFoundRoute() {
  return <Navigate to="/user/signin?redirect=/" replace />
}
