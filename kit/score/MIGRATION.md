# Score wrapper migration for kit v0.13.0

`registry.Max` and `Min` now take `*testing.T` and return `*score.T`.
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

Change shared helpers to accept `*score.T`, call `t.Error` or `t.Errorf` to record `TestDetails`, and call `t.Inc` or `t.Dec` for scoring.
For an ordinary Go test, pass `score.NewUnscored(t)` to the same helper.
`MaxByName`, `MinByName`, and `AddSub` are removed: register only the top-level test with `Add`, and replace `t.Run` with `sc.Run`.
All subtests then contribute to the one score of the top-level test, so size its `maxScore` for all of them.
Use `sc.Run` for subtests that share the parent score; it recovers subtest panics and prints the score after parallel subtests finish.
`T.Recover` is the only way to isolate a panic in a graded test.
`Score.Print`, `Score.PanicHandler`, `Score.Error`, `Score.Errorf`, `Score.Fatal`, `Score.Fatalf`, and `Score.Fail` are removed or unexported.
`T.Fail` zeroes the score and fails the test, and `T.Fatal` and `T.Fatalf` also zero the score.
`exercise.MultipleChoice` now takes a `*score.T` and no longer takes a separate `*testing.T`.
Courses that still call them must stay on an earlier kit version in `go.mod` until they migrate.
