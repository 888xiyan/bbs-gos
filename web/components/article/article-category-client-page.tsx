"use client"

import { ArticleList } from "@/components/article/article-list"
import { EmptyState } from "@/components/common/empty-state"
import { LoadMore } from "@/components/common/load-more"
import { HomeAside } from "@/components/layout/home-aside"
import { MainShell } from "@/components/layout/main-shell"
import { apiFetch } from "@/lib/api/client"
import type { Article, PageData } from "@/lib/api/types"
import { useI18n } from "@/lib/i18n/provider"
import { useDocumentTitle } from "@/lib/use-document-title"

import type { ArticleCategoryRouteData } from "@/app/route-helpers/loaders"

export function ArticleCategoryClientPage({
  initialData,
}: {
  initialData: ArticleCategoryRouteData
}) {
  const { t } = useI18n()
  useDocumentTitle(initialData.category.path || initialData.category.name)

  return (
    <MainShell aside={<HomeAside />}>
      <div className="rounded-lg bg-background">
        <div className="border-b px-4 py-3 text-base font-semibold">
          {initialData.category.path || initialData.category.name}
        </div>
        <LoadMore<Article>
          initialItems={initialData.results}
          initialCursor={initialData.cursor || "0"}
          initialHasMore={initialData.hasMore}
          initialLoad={false}
          autoLoadOnScroll
          resetKey={`article-category:${initialData.category.id}`}
          labels={{
            loadMore: t("common.loadMore.loadMore"),
            noMore: t("common.loadMore.noMore"),
          }}
          loadPage={({ cursor }) =>
            apiFetch<PageData<Article>>("/api/article/articles", {
              params: { categoryId: initialData.category.id, cursor },
            })
          }
          renderItems={(items) => <ArticleList articles={items} t={t} />}
          renderEmpty={() => <EmptyState title={t("common.noData")} />}
        />
      </div>
    </MainShell>
  )
}
