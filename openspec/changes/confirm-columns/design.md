# Design: confirm-columns

Row shape (C1 from the lab):

```
    [ok]  chore/old-deps
    [!]   experiment/abandoned   untracked files — will be lost
```

- Badge gutter: raw badge padded to 4 cells before styling, two-space gap.
- Name column: padded to the longest selected label, capped at 28 cells and
  pre-truncated with ellipsis, so one long branch cannot push the notes.
- Notes render only for at-risk rows, in the warning style; the per-category
  ⚠ fact lines below are unchanged.
