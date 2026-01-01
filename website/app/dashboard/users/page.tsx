"use client"

import * as React from "react"
import { useRouter } from "next/navigation"

import { AppSidebar } from "@/components/app-sidebar"
import { SiteHeader } from "@/components/site-header"
import { Button } from "@/components/ui/button"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"

type Role = "super_admin" | "admin" | "user"

type MeResponse = {
  user_id: string
  role: Role
  email: string
  github_id?: string | null
  nickname: string
}

type AdminUser = {
  ID: string
  Role: Role
  Email: string
  GithubID?: string | null
  PasswordEnabled: boolean
}

export default function UsersPage() {
  const router = useRouter()
  const [me, setMe] = React.useState<MeResponse | null>(null)
  const [users, setUsers] = React.useState<AdminUser[]>([])
  const [loading, setLoading] = React.useState(true)
  const [error, setError] = React.useState<string | null>(null)

  const canView = me?.role === "admin" || me?.role === "super_admin"
  const canEditRoles = me?.role === "super_admin"

  async function load() {
    setLoading(true)
    setError(null)
    try {
      const meRes = await fetch("/api/me", { credentials: "include" })
      if (!meRes.ok) {
        router.replace("/login")
        return
      }
      const meJson = (await meRes.json()) as MeResponse
      setMe(meJson)
      if (meJson.role !== "admin" && meJson.role !== "super_admin") {
        router.replace("/dashboard")
        return
      }

      const res = await fetch("/api/admin/users", { credentials: "include" })
      if (!res.ok) {
        if (res.status === 403) {
          router.replace("/dashboard")
          return
        }
        throw new Error(`failed to load users: ${res.status}`)
      }
      const json = (await res.json()) as { users: AdminUser[] }
      setUsers(json.users ?? [])
    } catch (e) {
      setError(e instanceof Error ? e.message : "unknown error")
    } finally {
      setLoading(false)
    }
  }

  async function updateRole(userId: string, role: Role) {
    const res = await fetch(`/api/admin/users/${encodeURIComponent(userId)}/role`, {
      method: "POST",
      headers: { "content-type": "application/json" },
      credentials: "include",
      body: JSON.stringify({ role }),
    })
    if (!res.ok) {
      throw new Error(`failed to update role: ${res.status}`)
    }
  }

  React.useEffect(() => {
    void load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  if (!canView && me) return null

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
                <div className="text-lg font-semibold">User Management</div>
                <div className="text-muted-foreground text-sm">
                  Only admins and super admins can access this page.
                </div>
              </div>
              <Button variant="outline" onClick={() => void load()} disabled={loading}>
                Refresh
              </Button>
            </div>

            <div className="px-4 pb-6 md:px-6">
              {error && <div className="text-destructive mb-3 text-sm">{error}</div>}

              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>User ID</TableHead>
                    <TableHead>Email</TableHead>
                    <TableHead>GitHub</TableHead>
                    <TableHead>Password</TableHead>
                    <TableHead>Role</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {users.map((u) => (
                    <TableRow key={u.ID}>
                      <TableCell className="max-w-[320px] truncate">{u.ID}</TableCell>
                      <TableCell className="max-w-65 truncate">{u.Email || "-"}</TableCell>
                      <TableCell>{u.GithubID ?? "-"}</TableCell>
                      <TableCell>{u.PasswordEnabled ? "enabled" : "-"}</TableCell>
                      <TableCell>
                        {u.Role === "super_admin" ? (
                          <span className="font-medium">super_admin</span>
                        ) : canEditRoles ? (
                          <Select
                            value={u.Role}
                            onValueChange={async (v) => {
                              const role = v as Role
                              const prev = u.Role
                              setUsers((xs) =>
                                xs.map((x) => (x.ID === u.ID ? { ...x, Role: role } : x))
                              )
                              try {
                                await updateRole(u.ID, role)
                              } catch (e) {
                                setUsers((xs) =>
                                  xs.map((x) => (x.ID === u.ID ? { ...x, Role: prev } : x))
                                )
                                setError(e instanceof Error ? e.message : "failed")
                              }
                            }}
                          >
                            <SelectTrigger className="w-40">
                              <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                              <SelectItem value="admin">admin</SelectItem>
                              <SelectItem value="user">user</SelectItem>
                            </SelectContent>
                          </Select>
                        ) : (
                          <span>{u.Role}</span>
                        )}
                      </TableCell>
                    </TableRow>
                  ))}
                  {!loading && users.length === 0 && (
                    <TableRow>
                      <TableCell colSpan={5} className="text-muted-foreground">
                        No users found.
                      </TableCell>
                    </TableRow>
                  )}
                  {loading && (
                    <TableRow>
                      <TableCell colSpan={5} className="text-muted-foreground">
                        Loading...
                      </TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
            </div>
          </div>
        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}

