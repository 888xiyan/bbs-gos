import { Navigate, redirect } from "react-router"
import { noindexRouteMeta } from "@/lib/seo"

export function meta({
  matches,
}: {
  matches: Array<{ data?: unknown; loaderData?: unknown }>
}) {
  return noindexRouteMeta(matches, "Not found", "页面不存在")
}

export function loader(): never {
  throw redirect("/user/signin?redirect=/")
}

export default function NotFoundRoute() {
  return <Navigate to="/user/signin?redirect=/" replace />
}
