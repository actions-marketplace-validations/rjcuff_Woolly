/** The questions people ask, shown on the homepage and in its Markdown for agents. */
export type Faq = { question: string; answer: string; keywords?: string[] };

export const FAQ: Faq[] = [
  {
    question: 'Does Woolly send my data anywhere?',
    answer:
      'No. There are no servers, no accounts and no telemetry. Your memories live in ~/.woolly on your laptop, and that is the whole story.',
    keywords: ['privacy', 'cloud', 'server', 'tracking', 'data', 'offline'],
  },
  {
    question: 'Which apps work with Woolly?',
    answer:
      'Anything that speaks MCP. Claude Code, Claude Desktop, Cursor, VS Code, Windsurf, Codex, Gemini CLI, LM Studio and opencode are one woolly connect away. Ollama and other local models get memory through woolly proxy.',
    keywords: ['claude', 'cursor', 'vscode', 'windsurf', 'codex', 'gemini', 'mcp', 'apps', 'tools'],
  },
  {
    question: 'Do I need Ollama?',
    answer:
      'No. Woolly works on its own. With Ollama running, it also searches by meaning, learns facts from your chats and saves decisions when a session ends.',
    keywords: ['ollama', 'local', 'model', 'llm'],
  },
  {
    question: 'How is this different from ChatGPT or Claude memory?',
    answer:
      "Those memories are locked inside one app, on that company's servers. Claude never hears what you told Cursor. Woolly is one memory every tool shares, and it lives on your laptop in files you can read.",
    keywords: ['chatgpt', 'openai', 'mem0', 'different', 'compare'],
  },
  {
    question: 'What happens when a fact changes?',
    answer:
      "A new memory with the same key replaces the old one. If an AI saves something that disagrees with what you said, your version stays and the AI's is saved as a conflict for you to review.",
    keywords: ['update', 'conflict', 'replace', 'wrong', 'old'],
  },
  {
    question: 'Can my team share memory?',
    answer:
      "Yes. Run woolly init in a repo and commit .woolly/. Every teammate's agent reads and updates the same project memory, and facts about you never go in the repo.",
    keywords: ['team', 'git', 'repo', 'project', 'share'],
  },
  {
    question: 'Can I use Woolly as the memory for agents I build?',
    answer:
      'Yes. The memory API uses mem0’s routes and field names, with JavaScript and Python packages. Add a conversation for a user, then search before your agent replies. It runs on your own machine and only answers that machine, so it fits agents you and your team run, not a hosted service for thousands of users.',
    keywords: ['api', 'sdk', 'mem0', 'app', 'product', 'build', 'agent'],
  },
  {
    question: 'How is Woolly different from mem0 or supermemory?',
    answer:
      'Woolly is one open-source program that runs on your own machine, with your memories in Markdown files you can read. Its API is shaped like mem0’s, so switching is mostly a new address. It keeps every old version of a fact, never lets a model overrule what you said, and dreams to turn your sessions into lessons.',
    keywords: ['mem0', 'supermemory', 'zep', 'letta', 'different', 'compare', 'alternative'],
  },
  {
    question: 'Is Woolly free?',
    answer: 'Yes. It is MIT licensed and built in public on GitHub. Issues and pull requests are welcome.',
    keywords: ['price', 'cost', 'license', 'mit', 'open', 'pay'],
  },
];
