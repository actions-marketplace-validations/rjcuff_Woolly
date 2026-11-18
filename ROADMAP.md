# Roadmap

Woolly's goal: the memory for developers and their teams. Open source and local first, with a paid service for teams that want their project memory kept in sync for them.

## How Woolly is put together

- **Woolly (free, MIT):** the command line, the Mac and Windows apps, MCP for every AI tool, project memory in git, and sync between your own devices.
- **SDKs (free, MIT):** `woolly-ai` for JavaScript and Python, plus plugins for the agent frameworks developers already use.
- **Woolly Teams (paid):** a hosted relay that keeps a team's project memory and devices in step. Encrypted end to end, so the relay can't read anything. Adds a team view of what everyone's AI tools learned, with approval before anything is saved to the project.

## Shipped

- One memory for Claude Code, Claude Desktop, Cursor, VS Code, Windsurf, Codex, Gemini CLI, LM Studio and opencode.
- Project memory in git with `woolly init`, and decisions saved when a Claude Code session ends.
- `woolly sync` between your own devices, encrypted, with no cloud.
- Mac and Windows apps.
- JavaScript and Python packages.

## Now: get Woolly in front of developers

- Show HN post.
- List Woolly in the MCP directories: the MCP Registry, Smithery, Glama, mcp.so and PulseMCP.
- A comparison page: Woolly, Mem0, OpenMemory and CLAUDE.md files.
- Run the full LOCOMO benchmark and publish the scores next to the ones cloud memory services report.
- Short demo videos: Claude recalls something Cursor learned, a teammate's agent knows the project's decisions.

## Next: the SDKs app builders need

- Separate memories per user in the API and SDKs, so an app can give each of its users their own memory.
- Plugins for LangChain, LlamaIndex, CrewAI, the OpenAI Agents SDK and Mastra.
- A `woolly serve` mode for servers: a Docker image and an API key per app.

## Then: Woolly Teams

- A hosted relay for team sync, encrypted end to end.
- A team view: what the team's AI tools learned this week, with approve and reject.
- Sign in, seats and billing.

## Later

- Find devices on the network automatically instead of typing an address.
- Use the model to merge near duplicates and summarize long histories.
