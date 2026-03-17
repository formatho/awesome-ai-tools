module github.com/formatho/agent-orchestrator/packages/agent-runner

go 1.25.0

require (
	github.com/formatho/agent-orchestrator/packages/agent-skills v0.0.0
	github.com/formatho/agent-orchestrator/packages/goagent v0.0.0
	github.com/formatho/agent-orchestrator/packages/llm-client v0.0.0
)

replace github.com/formatho/agent-orchestrator/packages/agent-skills => ../agent-skills
replace github.com/formatho/agent-orchestrator/packages/goagent => ../goagent
replace github.com/formatho/agent-orchestrator/packages/llm-client => ../llm-client
