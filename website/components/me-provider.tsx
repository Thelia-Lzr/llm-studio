"use client"

import * as React from "react"
import { usePathname } from "next/navigation"

export type Role = "super_admin" | "admin" | "user"

export type Me = {
  user_id: string
  role: Role
  email: string
  github_id?: string | null
  nickname: string
}

type MeContextValue = {
  me: Me | null
  setMe: (me: Me | null) => void
  refresh: () => Promise<Me | null>
}

const MeContext = React.createContext<MeContextValue | null>(null)

export function MeProvider({ children }: { children: React.ReactNode }) {
  const pathname = usePathname()
  const [me, setMe] = React.useState<Me | null>(null)

  const refresh = React.useCallback(async () => {
    try {
      const res = await fetch("/api/me", { credentials: "include" })
      if (!res.ok) {
        setMe(null)
        return null
      }
      const json = (await res.json()) as Me
      setMe(json)
      return json
    } catch {
      // ignore
      return null
    }
  }, [])

  React.useEffect(() => {
    void refresh()
  }, [pathname, refresh])

  const value = React.useMemo(() => ({ me, setMe, refresh }), [me, refresh])

  return <MeContext.Provider value={value}>{children}</MeContext.Provider>
}

export function useMe() {
  const ctx = React.useContext(MeContext)
  if (!ctx) throw new Error("useMe must be used within MeProvider")
  return ctx
}

