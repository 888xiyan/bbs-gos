"use client"

import * as React from "react"
import { useLocation } from "react-router"

import { useAuthChecked, useCurrentUser } from "@/components/app/app-provider"
import { buildSigninHref } from "@/lib/toast"
import { useRouter } from "@/lib/router/navigation"

function isPublicRoute(pathname: string) {
  const path = pathname.replace(/\/+$/, "") || "/"
  if (
    path === "/" ||
    path === "/about" ||
    path === "/links" ||
    path === "/search" ||
    path === "/redirect" ||
    path === "/install" ||
    path === "/user/signin" ||
    path === "/user/signup" ||
    path.startsWith("/user/signin/") ||
    path.startsWith("/user/password/") ||
    path === "/user/email/verify"
  ) {
    return true
  }

  if (
    path === "/topics" ||
    path.startsWith("/topics/") ||
    path === "/articles" ||
    path.startsWith("/articles/") ||
    path === "/topic" ||
    /^\/topic\/[^/]+$/.test(path) ||
    /^\/article\/[^/]+$/.test(path)
  ) {
    return true
  }

  // Public user profiles are addressable by user id; private account routes
  // remain protected below.
  return /^\/user\/(?!profile(?:\/|$)|messages(?:\/|$)|favorites(?:\/|$)|scores(?:\/|$))[^/]+(?:\/(?:articles|badges|fans|followed))?$/.test(
    path
  )
}

export function AuthRouteGuard({ children }: { children: React.ReactNode }) {
  const router = useRouter()
  const location = useLocation()
  const authChecked = useAuthChecked()
  const currentUser = useCurrentUser()
  const currentPath = `${location.pathname}${location.search}`
  const publicRoute = isPublicRoute(location.pathname)

  React.useEffect(() => {
    if (authChecked && !currentUser && !publicRoute) {
      router.replace(buildSigninHref(currentPath))
    }
  }, [authChecked, currentUser, currentPath, publicRoute, router])

  if (!authChecked && !publicRoute) {
    return null
  }
  if (authChecked && !currentUser && !publicRoute) {
    return null
  }

  return children
}
