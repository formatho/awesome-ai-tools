module github.com/formatho/agent-orchestrator/backend

go 1.25.0

require (
	github.com/fasthttp/websocket v1.5.12
	github.com/formatho/agent-orchestrator/packages/agent-config v0.0.0
	github.com/formatho/agent-orchestrator/packages/agent-pool v0.0.0
	github.com/formatho/agent-orchestrator/packages/cron-agents v0.0.0
	github.com/formatho/agent-orchestrator/packages/llm-client v0.0.0
	github.com/formatho/agent-orchestrator/packages/todo-queue v0.0.0
	github.com/gofiber/fiber/v2 v2.52.0
	github.com/google/uuid v1.6.0
	github.com/mattn/go-sqlite3 v1.14.22
	github.com/stretchr/testify v1.11.1
	github.com/valyala/fasthttp v1.58.0
)

require (
	cloud.google.com/go v0.115.0 // indirect
	cloud.google.com/go/ai v0.8.0 // indirect
	cloud.google.com/go/auth v0.17.0 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.8 // indirect
	cloud.google.com/go/compute/metadata v0.9.0 // indirect
	cloud.google.com/go/longrunning v0.5.7 // indirect
	github.com/BurntSushi/toml v1.3.2 // indirect
	github.com/Protocol-Lattice/go-agent v1.11.3 // indirect
	github.com/andybalholm/brotli v1.1.1 // indirect
	github.com/anthropics/anthropic-sdk-go v1.13.0 // indirect
	github.com/bahlo/generic-list-go v0.2.0 // indirect
	github.com/buger/jsonparser v1.1.1 // indirect
	github.com/caarlos0/env/v11 v11.3.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dlclark/regexp2 v1.11.5 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.13 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.30.1 // indirect
	github.com/google/generative-ai-go v0.20.1 // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.6 // indirect
	github.com/googleapis/gax-go/v2 v2.15.0 // indirect
	github.com/invopop/jsonschema v0.13.0 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/mailru/easyjson v0.9.1 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/ollama/ollama v0.12.5 // indirect
	github.com/pkoukk/tiktoken-go v0.1.8 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rivo/uniseg v0.4.6 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/sashabaranov/go-openai v1.41.2 // indirect
	github.com/savsgio/gotils v0.0.0-20240704082632-aef3928b8a38 // indirect
	github.com/stretchr/objx v0.5.3 // indirect
	github.com/teilomillet/gollm v0.1.11 // indirect
	github.com/tidwall/gjson v1.18.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/sjson v1.2.5 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/tcplisten v1.0.0 // indirect
	github.com/wk8/go-ordered-map/v2 v2.1.8 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.61.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.61.0 // indirect
	go.opentelemetry.io/otel v1.37.0 // indirect
	go.opentelemetry.io/otel/metric v1.37.0 // indirect
	go.opentelemetry.io/otel/trace v1.37.0 // indirect
	golang.org/x/crypto v0.47.0 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/oauth2 v0.31.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	golang.org/x/time v0.14.0 // indirect
	google.golang.org/api v0.252.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250707201910-8d1bb00bc6a7 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251002232023-7c0ddcbb5797 // indirect
	google.golang.org/grpc v1.75.1 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/formatho/agent-orchestrator/packages/agent-config => ../packages/agent-config
	github.com/formatho/agent-orchestrator/packages/agent-pool => ../packages/agent-pool
	github.com/formatho/agent-orchestrator/packages/agent-skills => ../packages/agent-skills
	github.com/formatho/agent-orchestrator/packages/cron-agents => ../packages/cron-agents
	github.com/formatho/agent-orchestrator/packages/llm-client => ../packages/llm-client
	github.com/formatho/agent-orchestrator/packages/todo-queue => ../packages/todo-queue
)
