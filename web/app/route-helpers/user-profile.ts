import { apiFetch } from "@/lib/api/client"
import type { UserSummary } from "@/lib/api/types"
import { redirect } from "react-router"

export type UserProfileRouteData = {
  user: UserSummary
}

export async function loadUserProfileRouteData({
  request,
  userId,
}: {
  request?: Request
  userId: string
}): Promise<UserProfileRouteData> {
  let user: UserSummary
  try {
    user = await apiFetch<UserSummary>(`/api/user/${userId}`, {
      ...(request ? { request } : {}),
    })
  } catch {
    throw redirect("/user/signin?redirect=/")
  }

  return { user }
}
