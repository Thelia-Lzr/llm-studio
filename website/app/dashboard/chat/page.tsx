"use client"

import * as React from "react"
import { useRouter } from "next/navigation"
import { toast } from "sonner"

import { streamText, type CoreMessage } from "ai"
import { createOpenAI } from "@ai-sdk/openai"

import { AppSidebar } from "@/components/app-sidebar"
import { SiteHeader } from "@/components/site-header"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"

type Role = "super_admin" | "admin" | "user"

type MeResponse = {
  user_id: string
  role: Role
  email: string
  github_id?: string | null
  nickname: string
}

type GatewayModel = {
  id: string
  name?: string
  provider?: string
  capabilities?: string[]
}

type ListModelsResponse = {
  data: GatewayModel[]
}

type ChatMessage = {
  id: string
  role: "system" | "user" | "assistant"
  content: string
}

async function issueToken(router: ReturnType<typeof useRouter>) {
  const res = await fetch("/api/llm/token", { method: "POST", credentials: "include" })
  if (!res.ok) {
    if (res.status === 401) {
      router.replace("/login")
      return null
    }
    throw new Error(`failed to issue token: ${res.status}`)
  }

  // BFF sets an HttpOnly cookie for llm-gateway auth; frontend only needs expiry.
  const json = (await res.json()) as { expires_at_unix: number }
  const exp = Math.floor(json.expires_at_unix)
  return { exp }
}

function tokenNeedsRefresh(expUnix: number, skewSeconds = 300) {
  const now = Math.floor(Date.now() / 1000)
  return expUnix <= now + skewSeconds
}

export default function ChatPage() {
  const router = useRouter()

  const gatewayBase =
    process.env.NEXT_PUBLIC_LLM_GATEWAY_HTTP_BASE_URL ?? "http://localhost:8081"

  const [me, setMe] = React.useState<MeResponse | null>(null)
  const [loading, setLoading] = React.useState(true)

  const [models, setModels] = React.useState<GatewayModel[]>([])
  const [model, setModel] = React.useState<string>("")

  const [messages, setMessages] = React.useState<ChatMessage[]>([])
  const [input, setInput] = React.useState("")
  const [sending, setSending] = React.useState(false)

  const tokenExpRef = React.useRef<number>(0)

  const canView = Boolean(me)

  function makeId() {
    if (typeof crypto !== "undefined" && "randomUUID" in crypto) return crypto.randomUUID()
    return `${Date.now()}-${Math.random().toString(16).slice(2)}`
  }

  async function ensureToken(): Promise<void> {
    const exp = tokenExpRef.current
    if (!exp || !Number.isFinite(exp) || tokenNeedsRefresh(exp)) {
      const issued = await issueToken(router)
      if (!issued) throw new Error("unauthenticated")
      tokenExpRef.current = issued.exp
    }
  }

  async function gatewayFetch(path: string, init?: RequestInit, retry = true) {
    const res = await fetch(`${gatewayBase}${path}`, {
      ...init,
      credentials: "include",
      headers: {
        ...(init?.headers ?? {}),
      },
    })

    if (res.status === 401 && retry) {
      // Cookie token missing/expired; re-issue and retry once.
      tokenExpRef.current = 0
      await ensureToken()
      return gatewayFetch(path, init, false)
    }

    return res
  }

  async function streamChatCompletion(
    nextMessages: Array<{ role: "system" | "user" | "assistant"; content: string }>,
    assistantId: string,
    retry = true
  ) {
    const applyDelta = (delta: string) => {
      if (!delta) return
      setMessages((prev) =>
        prev.map((m) => (m.id === assistantId ? { ...m, content: m.content + delta } : m))
      )
    }

    const gatewayOpenAI = createOpenAI({
      baseURL: `${gatewayBase}/v1`,
      // Gateway auth is done via HttpOnly cookie. We provide a placeholder key and strip the
      // Authorization header in the custom fetch below.
      apiKey: "cookie",
      fetch: async (input, init) => {
        const headers = new Headers(init?.headers)
        headers.delete("authorization")

        const run = () =>
          fetch(input, {
            ...init,
            headers,
            credentials: "include",
          })

        let res = await run()
        if (res.status === 401 && retry) {
          tokenExpRef.current = 0
          await ensureToken()
          res = await run()
        }
        return res
      },
    })

    const coreMessages: CoreMessage[] = nextMessages.map((m) => ({
      role: m.role,
      content: m.content,
    }))

    try {
      const { textStream } = streamText({
        model: gatewayOpenAI.chat(model),
        messages: coreMessages,
      })

      for await (const part of textStream) {
        applyDelta(part)
      }
    } catch (e) {
      const message = e instanceof Error ? e.message : String(e)
      throw new Error(message || "chat failed")
    }
  }

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

      await ensureToken()

      const modelsRes = await gatewayFetch("/v1/models")
      if (!modelsRes.ok) {
        throw new Error(`failed to load models: ${modelsRes.status}`)
      }
      const json = (await modelsRes.json()) as ListModelsResponse
      const items = json.data ?? []
      setModels(items)
      setModel((prev) => prev || items[0]?.id || "")

      if (items.length === 0) {
        toast.message("No models available", {
          description: "Ask an admin to configure providers/models in /dashboard/llm.",
        })
      }
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Failed to load")
    } finally {
      setLoading(false)
    }
  }

  React.useEffect(() => {
    void load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  React.useEffect(() => {
    const id = window.setInterval(() => {
      const exp = tokenExpRef.current
      if (exp && Number.isFinite(exp) && tokenNeedsRefresh(exp, 600)) {
        void issueToken(router)
          .then((issued) => {
            if (issued) tokenExpRef.current = issued.exp
          })
          .catch(() => {})
      }
    }, 60_000)
    return () => window.clearInterval(id)
  }, [router])

  async function send() {
    const text = input.trim()
    if (!text) return
    if (!model) {
      toast.error("No model selected")
      return
    }

    if (sending) return
    setInput("")
    setSending(true)

    const userId = makeId()
    const assistantId = makeId()

    const nextMessages = [...messages, { id: userId, role: "user", content: text } as const]
    setMessages([...nextMessages, { id: assistantId, role: "assistant", content: "" }])

    try {
      await ensureToken()
      await streamChatCompletion(
        nextMessages.map((m) => ({ role: m.role, content: m.content })),
        assistantId
      )
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Chat failed")
    } finally {
      setSending(false)
    }
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
                <div className="text-lg font-semibold">Chat</div>
                <div className="text-muted-foreground text-sm">
                  Streams from the browser using Vercel AI SDK to llm-gateway-http (OpenAI-compatible /v1).
                </div>
              </div>
              <Button variant="outline" onClick={() => void load()} disabled={loading}>
                Refresh
              </Button>
            </div>

            <div className="px-4 pb-6 md:px-6">
              <Tabs defaultValue="chat">
                <TabsList>
                  <TabsTrigger value="chat">Chat</TabsTrigger>
                  <TabsTrigger value="models">Models</TabsTrigger>
                </TabsList>

                <TabsContent value="chat" className="mt-4">
                  <div className="space-y-4 rounded-lg border p-4">
                    <div className="grid gap-3 md:grid-cols-2">
                      <div className="space-y-2">
                        <Label htmlFor="chat-model">Model</Label>
                        <select
                          id="chat-model"
                          className="border-input bg-background flex h-9 w-full rounded-md border px-3 py-1 text-sm shadow-sm"
                          value={model}
                          onChange={(e) => setModel(e.target.value)}
                          disabled={loading || sending}
                        >
                          {models.map((m) => (
                            <option key={m.id} value={m.id}>
                              {m.id}
                            </option>
                          ))}
                          {!loading && models.length === 0 && <option value="">(no models)</option>}
                        </select>
                      </div>
                      <div className="space-y-2">
                        <Label>Gateway</Label>
                        <Input value={gatewayBase} readOnly />
                      </div>
                    </div>

                    <div className="h-105 overflow-auto rounded-md border p-3 text-sm">
                      {messages.length === 0 && (
                        <div className="text-muted-foreground">Send a message to start.</div>
                      )}
                      <div className="space-y-3">
                        {messages.map((m) => (
                          <div key={m.id} className="space-y-1">
                            <div className="text-muted-foreground text-xs">{m.role}</div>
                            <div className="whitespace-pre-wrap">{m.content || (m.role === "assistant" ? "…" : "")}</div>
                          </div>
                        ))}
                      </div>
                    </div>

                    <div className="flex gap-2">
                      <Input
                        value={input}
                        onChange={(e) => setInput(e.target.value)}
                        placeholder="Type a message..."
                        disabled={loading || sending || models.length === 0}
                        onKeyDown={(e) => {
                          if (e.key === "Enter" && !e.shiftKey) {
                            e.preventDefault()
                            void send()
                          }
                        }}
                      />
                      <Button onClick={() => void send()} disabled={loading || sending || !input.trim()}>
                        Send
                      </Button>
                    </div>
                  </div>
                </TabsContent>

                <TabsContent value="models" className="mt-4">
                  <div className="space-y-2 rounded-lg border p-4 text-sm">
                    <div className="text-muted-foreground">Available models from llm-gateway-http:</div>
                    <ul className="list-disc space-y-1 pl-5">
                      {models.map((m) => (
                        <li key={m.id}>
                          <span className="font-medium">{m.id}</span>
                          {m.provider ? ` (${m.provider})` : ""}
                        </li>
                      ))}
                      {!loading && models.length === 0 && <li>(none)</li>}
                    </ul>
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
