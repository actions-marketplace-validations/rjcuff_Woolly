"""Python client for Woolly, the local memory for your AI tools.

Woolly runs on your own computer. Start it with ``woolly serve``, then:

    import woolly

    woolly.remember("I prefer short answers")
    print(woolly.recall("how should replies be written?"))

    # Memory for the agents you build, with mem0's API
    woolly.add([{"role": "user", "content": "I moved to Austin"}], user_id="alice")
    woolly.search("where does she live?", user_id="alice")

    # Add what Woolly knows to any AI prompt
    system_prompt = "You are a helpful assistant.\\n\\n" + woolly.context(user_message)

Everything stays on your computer. See https://usewoolly.com/docs/python.
"""

from .client import (
    Memory,
    Profile,
    RememberResult,
    Woolly,
    WoollyError,
    WoollyNotRunning,
)

__all__ = [
    "Memory",
    "Profile",
    "RememberResult",
    "Woolly",
    "WoollyError",
    "WoollyNotRunning",
    "add",
    "context",
    "forget",
    "learn",
    "profile",
    "recall",
    "remember",
    "search",
]

__version__ = "0.4.0"

_default = None


def _client() -> Woolly:
    global _default
    if _default is None:
        _default = Woolly()
    return _default


def remember(content, key=None, tags=None, source="user"):
    """Save a memory. See :meth:`Woolly.remember`."""
    return _client().remember(content, key=key, tags=tags, source=source)


def recall(query, limit=5):
    """Find memories relevant to ``query``. See :meth:`Woolly.recall`."""
    return _client().recall(query, limit=limit)


def forget(memory_id):
    """Move a memory to the trash. See :meth:`Woolly.forget`."""
    return _client().forget(memory_id)


def profile():
    """Get a summary of what Woolly knows. See :meth:`Woolly.profile`."""
    return _client().profile()


def context(query, limit=5):
    """Get a note of relevant memories to add to a prompt. See :meth:`Woolly.context`."""
    return _client().context(query, limit=limit)


def learn(text, project=None, dry_run=False):
    """Pick out facts from a conversation with a local model and save them."""
    return _client().learn(text, project=project, dry_run=dry_run)


def add(messages, user_id=None, agent_id=None, run_id=None, project=None, metadata=None, infer=True):
    """Add a conversation to memory. See :meth:`Woolly.add`."""
    return _client().add(messages, user_id=user_id, agent_id=agent_id, run_id=run_id, project=project, metadata=metadata, infer=infer)


def search(query, user_id=None, agent_id=None, run_id=None, project=None, top_k=5):
    """Search a scope's memories. See :meth:`Woolly.search`."""
    return _client().search(query, user_id=user_id, agent_id=agent_id, run_id=run_id, project=project, top_k=top_k)
