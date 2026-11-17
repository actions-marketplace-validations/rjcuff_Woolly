# Stale-fact benchmark

Does Woolly flag a fact when the code under it changes, and leave it alone when the code changes in ways that don't matter?

## Run it

```sh
python3 bench/stale/run.py
```

It needs git and Go, and builds `woolly` from this checkout (or pass `--woolly path/to/woolly`). It needs no model and no network, and it never touches your own memories. A full run takes about a minute. Results go to `bench/stale/results/`, which isn't committed. Add `--json` for every case in full.

## How it works

Each case starts from the same small ESP32 drone firmware repo, with pins, config, IMU, motor, PID, radio and battery code in C, plus a Python flashing script. For each case, the script

1. makes a fresh repo and an empty Woolly folder,
2. saves one fact about the code with `woolly remember --source-ref`, the way an agent would,
3. makes one change and commits it,
4. runs `woolly check` and records whether the fact was flagged as stale.

There are 50 cases. In 30 the change makes the fact wrong and it should be flagged.

| Kind | Cases | Example |
| --- | --- | --- |
| Rename | 8 | `imu_read_accel` becomes `imu_read_accel_raw`, and the fact names it |
| Deleted file | 5 | `main/battery.c` is deleted |
| Changed value | 15 | `IMU_CS` moves from GPIO 5 to GPIO 4, `PID_KP` goes from 1.2 to 0.9, a pin in a line range changes |
| Removed code | 2 | `pid_reset` is removed |

In 20 the fact is still true and it should not be flagged.

| Kind | Cases | Example |
| --- | --- | --- |
| Comment | 4 | A comment is added inside the function or after the value |
| Layout | 3 | Re-indenting, blank lines, braces on a new line |
| Unrelated | 4 | Another value or function in the same file changes |
| Moved | 8 | A function the fact doesn't name is renamed, a function moves to another file, lines are added above a line range, a file is renamed |
| Changed | 1 | The function body changes but keeps what the fact says |

Two of the changed-value cases are hard on purpose. The fact says "20 kHz" where the code says `20000`, or spells a number out. Woolly compares the values a fact states with the code, so it can't connect those, and it reports them as changed rather than stale.

## Results

On Nov 17, 2026, on a MacBook:

| | |
| --- | --- |
| Precision | 100% (28 of 28 flagged facts were really stale) |
| Recall | 93.3% (28 of 30 stale facts were flagged) |
| Median `woolly check` time | 35 ms |

The two misses are the two hard cases above. No fact was flagged when it shouldn't have been.

These are cases we wrote, on a repo we wrote, so treat them as a test of the design rather than a measure of every codebase. Pull requests with cases that Woolly gets wrong are very welcome.
