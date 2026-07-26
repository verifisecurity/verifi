# Use verifi in your editor

verifi runs as a local MCP server your editor starts on demand. It reads the workspace on your
machine; nothing is uploaded, and there is no daemon or account. This is the IDE-first way to use
verifi: your coding agent calls its tools while you work.

## Install

```
verifi mcp install vscode
verifi mcp install cursor
verifi mcp install claude
```

This writes a project-scoped MCP config (`.vscode/mcp.json`, `.cursor/mcp.json`, or `.mcp.json`)
pointing the editor at `verifi mcp`, keeping any servers already in the file. Restart the editor to
pick it up.

To wire it by hand instead, add one server that runs `verifi mcp` over stdio:

```json
{
  "mcpServers": {
    "verifi": { "command": "verifi", "args": ["mcp"] }
  }
}
```

VS Code uses `"servers"` and an explicit `"type": "stdio"`.

## Tools your agent gets

<!-- BEGIN CAPABILITIES -->
```
scan_workspace     Scan the project for vulnerable dependencies and their fixes.
list_dependencies  Resolve the full dependency tree, as inventory or SBOM.
propose_fix        Propose a gated fix for each vulnerable dependency, without applying it.
apply_fix          Apply a confirmed, gate-approved fix to the workspace.
```
<!-- END CAPABILITIES -->

The editor discovers these live each session, so they stay current as verifi grows. No reinstall
when a new tool lands.

## Keeping OSV fresh

`scan_workspace` matches against a local OSV database. Run `verifi update` once, and on a schedule
(a weekly task), to keep it current. That download is the only external step; everything else is
local.
