# Woolly plugin for the Claude plugin directory

A smaller copy of the plugin in `../plugin`, with only the `woolly-check` skill. The directory accepts local MCP servers and command hooks only from Verified organizations, so this one leaves them out.

For the full plugin (MCP server and session hooks), use `claude plugin marketplace add rjcuff/woolly`.

Keep `skills/woolly-check/SKILL.md` and the version in `.claude-plugin/plugin.json` the same as in `../plugin`. `internal/connect/plugin_test.go` checks both.
