"use client"

import { CommentSection } from "@/components/comment"
import type {
  Comment,
  PageData,
  SiteConfig,
  UserSummary,
} from "@/lib/api/types"

export function ArticleComments({
  entityId,
  commentCount,
  initialComments,
  allowComments = true,
  categoryPath,
  config,
}: {
  entityId: string | number
  commentCount?: number
  initialComments: PageData<Comment>
  allowComments?: boolean
  categoryPath?: string
  currentUser?: UserSummary | null
  config?: SiteConfig | null
}) {
  return (
    <CommentSection
      entityType="article"
      entityId={String(entityId)}
      commentCount={commentCount}
      initialData={initialComments}
      allowNewComments={allowComments && config?.commentsEnabled !== false}
      contextLabel={categoryPath}
    />
  )
}
