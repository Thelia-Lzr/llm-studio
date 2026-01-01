"use client"

import * as React from "react"
import { useRouter } from "next/navigation"
import { toast } from "sonner"

import { AppSidebar } from "@/components/app-sidebar"
import { SiteHeader } from "@/components/site-header"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"

type Role = "super_admin" | "admin" | "user"

type MeResponse = {
  user_id: string
  role: Role
  email: string
  github_id?: string | null
  nickname: string
}

type ProviderType = "dashscope" | "openrouter"

type ProviderConfigView = {
  provider: ProviderType
  base_url: string
  timeout_seconds: number
  api_key_present: boolean
}

type ModelSpec = {
  id: string
  provider: ProviderType
  upstream_model: string
  capabilities: string[]
}

const ALL_CAPABILITIES = [
  "text",
  "images",
  "audio",
  "video",
  "tools",
  "prompt_cache",
  "streaming",
  "reasoning",
] as const

type Capability = (typeof ALL_CAPABILITIES)[number]

export default function LLMAdminPage() {
  const router = useRouter()

  const [me, setMe] = React.useState<MeResponse | null>(null)

  const [providers, setProviders] = React.useState<ProviderConfigView[]>([])
  const [models, setModels] = React.useState<ModelSpec[]>([])

  const [loading, setLoading] = React.useState(true)

  const [providerType, setProviderType] = React.useState<ProviderType>("dashscope")
  const [providerBaseURL, setProviderBaseURL] = React.useState("")
  const [providerAPIKey, setProviderAPIKey] = React.useState("")
  const [providerTimeoutSeconds, setProviderTimeoutSeconds] = React.useState<string>("")

  const [modelProvider, setModelProvider] = React.useState<ProviderType | "">("")
  const [modelUpstream, setModelUpstream] = React.useState("")
  const [modelCaps, setModelCaps] = React.useState<Set<Capability>>(new Set(["text"]))

  const canView = me?.role === "admin" || me?.role === "super_admin"

  const configuredProviders = React.useMemo(() => {
    // If a provider is not present in the list, it is NOT configured and must not be used.
    return providers.map((p) => p.provider)
  }, [providers])

  const hasConfiguredProviders = configuredProviders.length > 0

  React.useEffect(() => {
    // Keep model provider selection valid.
    if (!hasConfiguredProviders) {
      setModelProvider("")
      return
    }
    if (!modelProvider || !configuredProviders.includes(modelProvider)) {
      setModelProvider(configuredProviders[0] ?? "")
    }
  }, [configuredProviders, hasConfiguredProviders, modelProvider])

  async function load() {
    setLoading(true)
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

      const [provRes, modelRes] = await Promise.all([
        fetch("/api/admin/llm/providers", { credentials: "include" }),
        fetch("/api/admin/llm/models", { credentials: "include" }),
      ])

      if (!provRes.ok) {
        if (provRes.status === 403) {
          router.replace("/dashboard")
          return
        }
        throw new Error(`failed to load providers: ${provRes.status}`)
      }
      if (!modelRes.ok) {
        if (modelRes.status === 403) {
          router.replace("/dashboard")
          return
        }
        throw new Error(`failed to load models: ${modelRes.status}`)
      }

      const provJson = (await provRes.json()) as { configs: ProviderConfigView[] }
      const modelJson = (await modelRes.json()) as { models: ModelSpec[] }

      setProviders(provJson.configs ?? [])
      setModels(modelJson.models ?? [])
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Failed to load LLM admin")
    } finally {
      setLoading(false)
    }
  }

  React.useEffect(() => {
    void load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  async function upsertProvider() {
    try {
      const timeout = providerTimeoutSeconds.trim()
      const body: Record<string, unknown> = {
        base_url: providerBaseURL,
        api_key: providerAPIKey,
      }
      if (timeout !== "") {
        const v = Number(timeout)
        if (!Number.isFinite(v) || v < 0) {
          throw new Error("timeout_seconds must be a non-negative number")
        }
        body.timeout_seconds = v
      }

      const res = await fetch(`/api/admin/llm/providers/${providerType}`, {
        method: "PUT",
        headers: { "content-type": "application/json" },
        credentials: "include",
        body: JSON.stringify(body),
      })

      if (!res.ok) {
        if (res.status === 401) {
          router.replace("/login")
          return
        }
        if (res.status === 403) {
          router.replace("/dashboard")
          return
        }
        throw new Error(`failed to save provider: ${res.status}`)
      }

      toast.success("Provider config saved.")
      await load()
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Failed to save provider")
    }
  }

  async function deleteProvider(provider: ProviderType) {
    try {
      const res = await fetch(`/api/admin/llm/providers/${provider}`, {
        method: "DELETE",
        credentials: "include",
      })
      if (!res.ok) {
        if (res.status === 401) {
          router.replace("/login")
          return
        }
        if (res.status === 403) {
          router.replace("/dashboard")
          return
        }
        throw new Error(`failed to delete provider: ${res.status}`)
      }

      toast.success("Provider deleted.")
      await load()
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Failed to delete provider")
    }
  }

  async function upsertModel() {
    try {
      if (!modelProvider) throw new Error("Please configure a provider first")
      if (!configuredProviders.includes(modelProvider)) {
        throw new Error("Selected provider is not configured")
      }

      const upstream = modelUpstream.trim()
      if (!upstream) throw new Error("upstream_model is required")

      const res = await fetch("/api/admin/llm/models", {
        method: "POST",
        headers: { "content-type": "application/json" },
        credentials: "include",
        body: JSON.stringify({
          provider: modelProvider,
          upstream_model: upstream,
          capabilities: Array.from(modelCaps.values()),
        }),
      })

      if (!res.ok) {
        if (res.status === 401) {
          router.replace("/login")
          return
        }
        if (res.status === 403) {
          router.replace("/dashboard")
          return
        }
        throw new Error(`failed to save model: ${res.status}`)
      }

      const json = (await res.json()) as { id: string }
      toast.success(`Model saved: ${json.id}`)
      await load()
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Failed to save model")
    }
  }

  async function deleteModel(id: string) {
    try {
      const res = await fetch(`/api/admin/llm/models?id=${encodeURIComponent(id)}`, {
        method: "DELETE",
        credentials: "include",
      })
      if (!res.ok) {
        if (res.status === 401) {
          router.replace("/login")
          return
        }
        if (res.status === 403) {
          router.replace("/dashboard")
          return
        }
        throw new Error(`failed to delete model: ${res.status}`)
      }

      toast.success("Model deleted.")
      await load()
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Failed to delete model")
    }
  }

  function toggleCap(cap: Capability, checked: boolean) {
    setModelCaps((prev) => {
      const next = new Set(prev)
      if (checked) next.add(cap)
      else next.delete(cap)
      return next
    })
  }

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
                <div className="text-lg font-semibold">LLM Management</div>
                <div className="text-muted-foreground text-sm">
                  Only admins and super admins can access this page.
                </div>
              </div>
              <Button variant="outline" onClick={() => void load()} disabled={loading}>
                Refresh
              </Button>
            </div>

            <div className="px-4 pb-6 md:px-6">
              <Tabs defaultValue="providers">
                <TabsList>
                  <TabsTrigger value="providers">Providers</TabsTrigger>
                  <TabsTrigger value="models">Models</TabsTrigger>
                </TabsList>

                <TabsContent value="providers" className="mt-4 space-y-6">
                  <div className="grid gap-4 md:grid-cols-2">
                    <div className="space-y-4 rounded-lg border p-4">
                      <div className="font-medium">Upsert Provider</div>

                      <div className="space-y-2">
                        <Label htmlFor="provider-type">Provider</Label>
                        <select
                          id="provider-type"
                          className="border-input bg-background flex h-9 w-full rounded-md border px-3 py-1 text-sm shadow-sm"
                          value={providerType}
                          onChange={(e) => setProviderType(e.target.value as ProviderType)}
                          disabled={loading}
                        >
                          <option value="dashscope">dashscope</option>
                          <option value="openrouter">openrouter</option>
                        </select>
                      </div>

                      <div className="space-y-2">
                        <Label htmlFor="provider-base-url">Base URL (optional)</Label>
                        <Input
                          id="provider-base-url"
                          value={providerBaseURL}
                          onChange={(e) => setProviderBaseURL(e.target.value)}
                          placeholder="Leave blank to use llm-gateway default"
                          disabled={loading}
                        />
                        <div className="text-muted-foreground text-xs">
                          If empty, llm-gateway will use its preset base URL.
                        </div>
                      </div>

                      <div className="space-y-2">
                        <Label htmlFor="provider-api-key">API Key</Label>
                        <Input
                          id="provider-api-key"
                          value={providerAPIKey}
                          onChange={(e) => setProviderAPIKey(e.target.value)}
                          placeholder="sk-..."
                          disabled={loading}
                        />
                      </div>

                      <div className="space-y-2">
                        <Label htmlFor="provider-timeout">Timeout (seconds)</Label>
                        <Input
                          id="provider-timeout"
                          value={providerTimeoutSeconds}
                          onChange={(e) => setProviderTimeoutSeconds(e.target.value)}
                          placeholder="(optional)"
                          disabled={loading}
                        />
                      </div>

                      <div className="flex items-center gap-2">
                        <Button onClick={() => void upsertProvider()} disabled={loading}>
                          Save
                        </Button>
                        <Button
                          variant="outline"
                          onClick={() => {
                            setProviderBaseURL("")
                            setProviderAPIKey("")
                            setProviderTimeoutSeconds("")
                          }}
                          disabled={loading}
                        >
                          Reset
                        </Button>
                      </div>
                    </div>

                    <div className="space-y-3 rounded-lg border p-4">
                      <div className="font-medium">Current Providers</div>
                      <Table>
                        <TableHeader>
                          <TableRow>
                            <TableHead>Provider</TableHead>
                            <TableHead>Base URL</TableHead>
                            <TableHead>Timeout</TableHead>
                            <TableHead>API Key</TableHead>
                            <TableHead className="text-right">Actions</TableHead>
                          </TableRow>
                        </TableHeader>
                        <TableBody>
                          {providers.map((p) => (
                            <TableRow key={p.provider}>
                              <TableCell className="font-medium">{p.provider}</TableCell>
                              <TableCell className="max-w-70 truncate">
                                {p.base_url || "-"}
                              </TableCell>
                              <TableCell>{p.timeout_seconds || "-"}</TableCell>
                              <TableCell>{p.api_key_present ? "present" : "-"}</TableCell>
                              <TableCell className="text-right">
                                <Button
                                  variant="destructive"
                                  size="sm"
                                  onClick={() => void deleteProvider(p.provider)}
                                  disabled={loading}
                                >
                                  Delete
                                </Button>
                              </TableCell>
                            </TableRow>
                          ))}
                          {!loading && providers.length === 0 && (
                            <TableRow>
                              <TableCell colSpan={5} className="text-muted-foreground">
                                No providers configured.
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
                </TabsContent>

                <TabsContent value="models" className="mt-4 space-y-6">
                  <div className="grid gap-4 md:grid-cols-2">
                    <div className="space-y-4 rounded-lg border p-4">
                      <div className="font-medium">Upsert Model</div>

                      {!hasConfiguredProviders && (
                        <div className="text-muted-foreground text-sm">
                          Configure a provider first (Providers tab) before creating models.
                        </div>
                      )}

                      <div className="space-y-2">
                        <Label htmlFor="model-provider">Provider</Label>
                        <select
                          id="model-provider"
                          className="border-input bg-background flex h-9 w-full rounded-md border px-3 py-1 text-sm shadow-sm"
                          value={modelProvider}
                          onChange={(e) => setModelProvider(e.target.value as ProviderType)}
                          disabled={loading || !hasConfiguredProviders}
                        >
                          {!hasConfiguredProviders && <option value="">(not configured)</option>}
                          {configuredProviders.map((p) => (
                            <option key={p} value={p}>
                              {p}
                            </option>
                          ))}
                        </select>
                      </div>

                      <div className="space-y-2">
                        <Label htmlFor="model-upstream">Upstream model</Label>
                        <Input
                          id="model-upstream"
                          value={modelUpstream}
                          onChange={(e) => setModelUpstream(e.target.value)}
                          placeholder="gpt-4o-mini / qwen-plus / ..."
                          disabled={loading}
                        />
                      </div>

                      <div className="space-y-2">
                        <Label>Capabilities</Label>
                        <div className="grid grid-cols-2 gap-2 md:grid-cols-3">
                          {ALL_CAPABILITIES.map((cap) => (
                            <label
                              key={cap}
                              className="flex items-center gap-2 rounded-md border px-3 py-2 text-sm"
                            >
                              <Checkbox
                                checked={modelCaps.has(cap)}
                                onCheckedChange={(v) => toggleCap(cap, Boolean(v))}
                                disabled={loading}
                              />
                              <span>{cap}</span>
                            </label>
                          ))}
                        </div>
                      </div>

                      <div className="flex items-center gap-2">
                        <Button
                          onClick={() => void upsertModel()}
                          disabled={loading || !hasConfiguredProviders}
                        >
                          Save
                        </Button>
                        <Button
                          variant="outline"
                          onClick={() => {
                            setModelUpstream("")
                            setModelCaps(new Set(["text"]))
                          }}
                          disabled={loading || !hasConfiguredProviders}
                        >
                          Reset
                        </Button>
                      </div>
                    </div>

                    <div className="space-y-3 rounded-lg border p-4">
                      <div className="font-medium">Current Models</div>
                      <Table>
                        <TableHeader>
                          <TableRow>
                            <TableHead>ID</TableHead>
                            <TableHead>Provider</TableHead>
                            <TableHead>Upstream</TableHead>
                            <TableHead>Capabilities</TableHead>
                            <TableHead className="text-right">Actions</TableHead>
                          </TableRow>
                        </TableHeader>
                        <TableBody>
                          {models.map((m) => (
                            <TableRow key={m.id}>
                              <TableCell className="max-w-55 truncate font-mono text-xs">
                                {m.id}
                              </TableCell>
                              <TableCell>{m.provider}</TableCell>
                              <TableCell className="max-w-55 truncate">
                                {m.upstream_model}
                              </TableCell>
                              <TableCell className="max-w-65 truncate">
                                {(m.capabilities ?? []).join(", ") || "-"}
                              </TableCell>
                              <TableCell className="text-right">
                                <Button
                                  variant="destructive"
                                  size="sm"
                                  onClick={() => void deleteModel(m.id)}
                                  disabled={loading}
                                >
                                  Delete
                                </Button>
                              </TableCell>
                            </TableRow>
                          ))}
                          {!loading && models.length === 0 && (
                            <TableRow>
                              <TableCell colSpan={5} className="text-muted-foreground">
                                No models configured.
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
                </TabsContent>
              </Tabs>
            </div>
          </div>
        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}
