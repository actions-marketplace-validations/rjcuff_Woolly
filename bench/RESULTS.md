# Benchmark results

Every number Woolly publishes, with how to get it yourself. All of them run on a laptop with no hosted model and no API key.

## Stale facts

Does Woolly flag a fact when the code under it changes? 50 scripted changes to a small ESP32 drone firmware repo. [How it works](stale/README.md)

```sh
python3 bench/stale/run.py
```

| | Nov 17, 2026 |
| --- | --- |
| Precision | 100% (28 of 28) |
| Recall | 93.3% (28 of 30) |
| Median `woolly check` time | 35 ms |

## LOCOMO

How often Woolly's search returns the messages that answer a question about a long conversation. [How it works](locomo/README.md)

```sh
python3 bench/locomo/run.py --retrieval-only
```

| | Nov 12, 2026, first 3 of 10 conversations, 382 questions |
| --- | --- |
| Answer in the top 20 | 83.2% |
| Answer in the top 5 | 65.7% |
| Median recall time | 19 ms |

## LongMemEval

Not run yet. It's next.
