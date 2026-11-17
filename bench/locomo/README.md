# LOCOMO benchmark

How well Woolly remembers long conversations, measured with [LOCOMO](https://github.com/snap-research/locomo), the benchmark mem0 and others report. Everything runs on your computer: Woolly for memory, Ollama for answering and grading. No API keys.

## Run it

```sh
make build
ollama pull qwen2.5:7b
ollama pull nomic-embed-text
python3 bench/locomo/run.py --conversations 1 --questions 30   # a quick try, about 3 minutes
python3 bench/locomo/run.py                                      # all 10 conversations
```

Options: `--retrieval-only` (measure only Woolly's search, no answer model), `--model` (answer model, default `qwen2.5:7b`), `--judge` (grading model, default the answer model), `--top-k` (memories per question, default 20), `--conversations`, `--questions`.

Results go to `bench/locomo/results/`, which isn't committed.

## How it works

1. Downloads `locomo10.json`. The dataset is CC BY-NC 4.0, so it isn't included in this repo.
2. Starts its own Woolly with an empty, temporary memory folder.
3. Saves every message of each conversation as a memory, starting with the date it was said, one project per conversation.
4. For each question, asks Woolly for the 20 closest memories, has the answer model answer from only those, and has the judge compare the answer with the correct one.
5. Reports accuracy by category, how often Woolly's results included a message that holds the answer, and how long recall took.

Adversarial questions (category 5) are left out, as in mem0's paper. The judge is generous in the same way published LOCOMO results are: an answer that says the same thing in other words counts.

## Results

### Search

Woolly's own job is search. Given a question, are the messages that answer it in what it returns? `--retrieval-only` measures just that, with no answer model.

```sh
python3 bench/locomo/run.py --retrieval-only
```

On the first three conversations, 382 questions, searched on a laptop with `nomic-embed-text`:

| | Top 5 | Top 10 | Top 20 | Every message needed, top 20 |
|---|---|---|---|---|
| Multi-hop | 67.1% | 76.7% | 83.6% | 35.6% |
| Dates and time | 83.3% | 92.2% | 95.6% | 93.3% |
| Open-domain | 26.3% | 47.4% | 63.2% | 47.4% |
| Single fact | 61.0% | 70.5% | 79.5% | 79.0% |
| **Overall** | **65.7%** | **75.7%** | **83.2%** | **72.5%** |

Median recall time was 19 ms. Before Woolly gave the embedding model its search and document prefixes, the overall top 20 score was 75.9%.

### Answer accuracy

On the first 30 questions of conversation 1, with `qwen2.5:7b` answering and grading:

| | Score |
|---|---|
| Overall | 56.7% |
| Woolly found the message with the answer | 86.7% |

Accuracy also depends on the answer model. A 7B model on a laptop misses more dates and multi-step questions than the hosted models most published numbers use, so these aren't directly comparable to mem0's.

## Reading the numbers

Two things are measured here, and they're worth keeping apart:

- **Retrieval** is Woolly's job: did the memories it returned include the message with the answer?
- **Accuracy** also depends on the answer model. A 7B model on a laptop gets dates and multi-step questions wrong more often than GPT-4o, which is what most published numbers use.

So these numbers aren't directly comparable to mem0's, which use hosted models for both answering and grading. Pass `--model` and `--judge` to try other models.
