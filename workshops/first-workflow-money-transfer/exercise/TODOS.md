# TODOs

One TODO. Challenge 1 changes no code; Challenge 2 fixes a bug while a
transfer is waiting on it.

| TODO | Challenge | File | What to do |
|---|---|---|---|
| 1 | 2 | `activities.go`, `Deposit` | Give the large-deposit audit note's slice a length of 2, `make([]string, 2)`, so a deposit of $1,000 or more stops panicking with `index out of range [0] with length 0`. |

The TODO sits between a `// TODO 1 start` and a `// TODO 1 end` comment. Change
only the lines between them, then restart the Worker: it does not reload on save.
