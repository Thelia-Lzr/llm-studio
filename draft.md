# LLM Studio 项目启动文档

LLM Studio 是一个基于 <https://github.com/poly-workshop/llm-gateway> 的 LLM 管理和开发平台，
它的主要设计目的是为 LLM Gateway 提供一个易于使用和管理的界面，同时支持用户快速体验 LLM 的强大能力。

LLM Studio 的项目性质是一个 Next.js + Golang 作为 BFF 的单体应用。
其中 BFF 设计上依赖以下项目：

- <https://github.com/poly-workshop/llm-gateway> LLM 相关基本能力
- <https://github.com/poly-workshop/identra> 身份认证和授权能力
- <https://github.com/poly-workshop/go-webmods> 通用 Web 模块，包括数据库、缓存、日志、配置、监控等能力

LLM Studio 的 UI 采用 Shadcn dashboard-01 和 login-03 模板，并在此基础上进行修改和扩展。

Golang 代码需要遵循标准化的代码规范，并使用干净架构（Clean Architecture）进行设计。
