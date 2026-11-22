# woolly-ai

Python client for [Woolly](https://usewoolly.com), the local memory for your AI tools.

Woolly keeps memories about the user on their own computer, keeps facts current, and shares them with every AI tool. This package lets your Python app use that memory.

## Install

```sh
pip install woolly-ai
```

You also need Woolly running on the same computer:

```sh
curl -fsSL https://usewoolly.com/install.sh | sh
woolly serve
```

## Use it

```python
import woolly

woolly.remember("Moved to Denver in October", key="user.location")
woolly.remember("I prefer short answers")

for m in woolly.recall("where do I live?"):
    print(m.content)
```

## Add memory to any AI app

`woolly.context()` returns a short note of relevant memories, ready to add to a prompt:

```python
import woolly
from openai import OpenAI

client = OpenAI()
question = "Any good hikes near me this weekend?"

response = client.chat.completions.create(
    model="gpt-4.1-mini",
    messages=[
        {"role": "system", "content": "You are a helpful assistant.\n\n" + woolly.context(question)},
        {"role": "user", "content": question},
    ],
)
```

It works the same with Anthropic, Ollama or any other model.

## Memory for the agents you build

The same API as mem0, on your own machine. `add` turns a conversation into memories for one user, and a local model decides whether each fact is new, updates an older one, or is no longer true. `search` finds them before your agent answers.

```python
woolly.add(
    [
        {"role": "user", "content": "I just moved to Austin"},
        {"role": "assistant", "content": "Nice!"},
    ],
    user_id="alice",
)
# {"results": [{"memory": "Lives in Austin", "event": "UPDATE", "previous_memory": "Lives in Denver", ...}]}

memories = woolly.search("where does she live?", user_id="alice")["results"]
```

Scope with `user_id`, `agent_id`, `run_id` or `project`. Also `get_all`, `history` (every version of a memory), `delete_all`, `link` (recall one memory, get a related one too), `feedback` (say whether a memory helped), `scoped_conflicts` and `resolve_scoped`, and `dream`, which merges duplicates and settles facts that changed. Pass `infer=False` to `add` to save messages as they are, without a model, and `memory_type="procedure"` (or `preference`, `episode`) to say what kind they are; `search` and `recall` take `type=[...]`. See [Memory API for agents](https://usewoolly.com/docs/memory-api).

## More

```python
from woolly import Woolly

w = Woolly()                      # finds the server and token automatically
r = w.remember("Prefers spaces", key="code.indent", source="agent")
if r.saved_as_conflict:           # it disagrees with something the user said
    print("Needs review:", r.conflicts_with.content)

w.profile().markdown              # a summary of everything known
w.forget(memory_id)               # moves it to the trash
w.restore(memory_id)              # brings it back
```

Woolly refuses passwords, API keys and instruction-like text that doesn't come from the user. Those calls raise `woolly.WoollyError`.

Docs: [usewoolly.com/docs/python](https://usewoolly.com/docs/python)
