/** The MCP server entry every app gets, the same one `woolly connect` writes. */
export interface McpEntry {
  command: string;
  args: string[];
}

export const WOOLLY_ENTRY: McpEntry = { command: 'woolly', args: ['mcp'] };

function encode(entry: McpEntry): string {
  return btoa(JSON.stringify(entry));
}

/** Opens Cursor and asks to add Woolly as an MCP server. */
export function cursorInstallUrl(entry: McpEntry = WOOLLY_ENTRY): string {
  const q = new URLSearchParams({ name: 'woolly', config: encode(entry) });
  return `cursor://anysphere.cursor-deeplink/mcp/install?${q.toString()}`;
}

/** Opens LM Studio and asks to add Woolly as an MCP server. */
export function lmStudioInstallUrl(entry: McpEntry = WOOLLY_ENTRY): string {
  const q = new URLSearchParams({ name: 'woolly', config: encode(entry) });
  return `lmstudio://add_mcp?${q.toString()}`;
}
