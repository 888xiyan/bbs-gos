import { useLoaderData } from "react-router"

import { ArticleCategoryClientPage } from "@/components/article/article-category-client-page"
import { categoryMeta, rootDataFromMatches } from "@/lib/seo"

import {
  loadArticleCategoryRouteData,
  type ArticleCategoryRouteData,
} from "../route-helpers/loaders"

type RouteArgs = {
  request: Request
  params: { id?: string }
}

export async function loader({ request, params }: RouteArgs) {
  return loadArticleCategoryRouteData({ request, id: params.id })
}

export async function clientLoader({ request, params }: RouteArgs) {
  return loadArticleCategoryRouteData({ request, id: params.id })
}

export function meta({
  data,
  location,
  matches,
}: {
  data: ArticleCategoryRouteData
  location: { pathname: string }
  matches: Array<{ data?: unknown; loaderData?: unknown }>
}) {
  return categoryMeta(
    rootDataFromMatches(matches)?.config,
    data.category,
    location.pathname
  )
}

export default function ArticleCategoryRoute() {
  const data = useLoaderData<typeof loader>()
  return <ArticleCategoryClientPage initialData={data} />
}
