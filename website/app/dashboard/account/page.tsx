"use client"

import * as React from "react"
import { useRouter } from "next/navigation"
import { toast } from "sonner"

import { AppSidebar } from "@/components/app-sidebar"
import { useMe, type Me } from "@/components/me-provider"
import { SiteHeader } from "@/components/site-header"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar"

export default function AccountPage() {
  const router = useRouter()
  const { me, setMe, refresh } = useMe()
  const [nickname, setNickname] = React.useState("")
  const [dirty, setDirty] = React.useState(false)
  const [loading, setLoading] = React.useState(true)
  const [saving, setSaving] = React.useState(false)

  async function load() {
    setLoading(true)
    try {
      const next = await refresh()
      if (!next) {
        router.replace("/login")
        return
      }
    } catch {
      toast.error("Failed to load account.")
    } finally {
      setLoading(false)
    }
  }

  async function save() {
    try {
      setSaving(true)
      const res = await fetch("/api/me", {
        method: "PATCH",
        headers: { "content-type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ nickname }),
      })
      if (!res.ok) {
        if (res.status === 401) {
          router.replace("/login")
          return
        }
        const msg =
          res.status === 400
            ? "Invalid nickname (max 32 chars)."
            : `Failed to update account (${res.status}).`
        throw new Error(msg)
      }
      const json = (await res.json()) as Me
      setMe(json)
      setNickname(json.nickname ?? "")
      setDirty(false)
      toast.success("Account updated.")
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Failed to update account.")
    } finally {
      setSaving(false)
    }
  }

  React.useEffect(() => {
    void load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  React.useEffect(() => {
    if (!dirty) setNickname(me?.nickname ?? "")
  }, [me?.nickname, dirty])

  return (
    <SidebarProvider
      style={
        {
          "--sidebar-width": "calc(var(--spacing) * 72)",
          "--header-height": "calc(var(--spacing) * 12)",
        } as React.CSSProperties
      }
    >
      <AppSidebar variant="inset" />
      <SidebarInset>
        <SiteHeader />
        <div className="flex flex-1 flex-col">
          <div className="@container/main flex flex-1 flex-col gap-2">
            <div className="flex items-center justify-between gap-4 px-4 py-4 md:px-6">
              <div>
                <div className="text-lg font-semibold">Account</div>
                <div className="text-muted-foreground text-sm">
                  Manage your profile details.
                </div>
              </div>
              <Button
                variant="outline"
                onClick={() => void load()}
                disabled={loading || saving}
              >
                Refresh
              </Button>
            </div>

            <div className="px-4 pb-6 md:px-6">
              <div className="max-w-xl space-y-6">
                <div className="space-y-2">
                  <Label htmlFor="nickname">Nickname</Label>
                  <Input
                    id="nickname"
                    value={nickname}
                  onChange={(e) => {
                    setNickname(e.target.value)
                    setDirty(true)
                  }}
                    placeholder="Enter a nickname"
                    maxLength={32}
                    disabled={loading || saving}
                  />
                  <p className="text-muted-foreground text-xs">
                    Shown across the dashboard. Max 32 characters.
                  </p>
                </div>

                <div className="space-y-2">
                  <Label htmlFor="email">Email</Label>
                  <Input
                    id="email"
                    value={me?.email ?? ""}
                    disabled
                    placeholder={loading ? "Loading..." : ""}
                  />
                </div>

                <div className="flex items-center gap-2">
                  <Button onClick={() => void save()} disabled={loading || saving}>
                    Save changes
                  </Button>
                  <Button
                    variant="outline"
                    onClick={() => {
                      setNickname(me?.nickname ?? "")
                      setDirty(false)
                    }}
                    disabled={loading || saving}
                  >
                    Reset
                  </Button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}

