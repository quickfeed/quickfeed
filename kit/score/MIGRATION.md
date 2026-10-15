# Score wrapper migration for kit v0.13.0

`registry.Max`, `Min`, `MaxByName`, and `MinByName` now take `*testing.T` and return `*score.T`.
Pass `t` to the lookup instead of relying on the caller's function name.
Use `sc.Score()` when a test needs to inspect the underlying protobuf score.

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
Use `sc.Run` for subtests that share the parent score; it recovers subtest panics and prints the score after parallel subtests finish.
`T.Recover` is the only way to isolate a panic in a graded test.
`Score.Print` and `Score.PanicHandler` are removed.
Courses that still call them must stay on an earlier kit version in `go.mod` until they migrate.
