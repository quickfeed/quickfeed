# Score wrapper migration for kit v0.13.0

Courses that still call the removed functions must stay on an earlier kit version in `go.mod` until they migrate.

## Summary of Changes

| Area            | Before                                                                      | After                                                      |
| --------------- | --------------------------------------------------------------------------- | ---------------------------------------------------------- |
| Test handle     | `scores.Max()` returns a `*Score`                                           | `scores.Max(t)` returns a `*score.T`                       |
| Printing        | `defer sc.Print(t)`                                                         | `defer sc.Recover()`                                       |
| Subtests        | `AddSub`, `t.Run`, and `MaxByName` or `MinByName`                           | `Add` for the top-level test and `sc.Run` for each subtest |
| Score methods   | `Score.Error`, `Errorf`, `Fatal`, `Fatalf`, `Fail`, `Print`, `PanicHandler` | removed or unexported; use the `*score.T` methods instead  |
| Multiple choice | `exercise.MultipleChoice(t, sc, ...)`                                       | `exercise.MultipleChoice(sc, ...)` with a `*score.T`       |
| `PrintTestInfo` | takes an optional `sorted` argument and always prints                       | takes no argument and prints only when `SCORE_INIT` is set |
| `TestDetails`   | records the full path of the file                                           | records the base name of the file                          |

## Wrapping a Test

`registry.Max` and `Min` take a `*testing.T` and return a `*score.T`.
Pass `t` to the lookup instead of relying on the caller's function name.

```go
// Before
sc := scores.Max()
defer sc.Print(t)
testAcceptor(t, sc.Dec)

// After
sc := scores.Max(t)
defer sc.Recover()
testAcceptor(sc)
```

- Change shared helpers to accept `*score.T`.
- Call `t.Error` or `t.Errorf` to record `TestDetails`, and `t.Inc` or `t.Dec` for scoring.
- For an ordinary Go test, pass `score.NewUnscored(t)` to the same helper.
- `T.Fail` zeroes the score and fails the test; `T.Fatal` and `T.Fatalf` also zero the score.
- `T.Recover` is the only way to isolate a panic in a graded test.

## Subtests

Subtests are no longer registered one by one.
`MaxByName`, `MinByName`, and `AddSub` are removed.

- Register only the top-level test with `Add`.
- Replace `t.Run` with `sc.Run`; it recovers subtest panics and prints the score after parallel subtests finish.
- All subtests contribute to the one score of the top-level test, so size its `maxScore` for all of them.

```go
// Before
scores.AddSub(TestTable, name, 1, 1) // once per row
// ...
t.Run(name, func(t *testing.T) {
	sc := scores.MaxByName(t)
	defer sc.Recover()
	// ...
})

// After
scores.Add(TestTable, len(rows), 1)
// ...
sc := scores.Max(t)
defer sc.Recover()
sc.Run(name, func(sc *score.T) {
	// ...
})
```

## Test Information

`PrintTestInfo` prints the registered tests, in registration order, only when `SCORE_INIT` is set, and then exits.
Otherwise, it does nothing, and the `sorted` argument is removed.
Test names, maximum scores, and weights come from `tests.json`.
