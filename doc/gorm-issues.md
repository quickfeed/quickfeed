# Note about GORM database issues

When using GORM, we must be careful when using false or zero values.
This is because GORM will ignore false or zero values, and will not produce the expected SQL query.
Here are some examples of different ways that may produce unexpected SQL queries:

```go
a.Where(&qf.Assignment{IsGroupLab: false}) // Bad: the false value will simply be ignored
a.Where("is_group_lab = ?", "false")       // Bad: the value will be a string, not a bool
a.Where("is_group_lab = ?", false)         // Good: the value will be a bool
a.Where("is_group_lab = false")            // Good: the value will be a bool (same as above)
a.Where(&qf.Assignment{IsGroupLab: true})  // Good: the true value will be used
```

The following SQL queries are produced by the above examples:

```sql
SELECT `id` FROM `assignments` WHERE `assignments`.`course_id` = 1 ORDER BY 'order'
SELECT `id` FROM `assignments` WHERE `assignments`.`course_id` = 1 AND is_group_lab = \"false\" ORDER BY 'order'
SELECT `id` FROM `assignments` WHERE `assignments`.`course_id` = 1 AND is_group_lab = false ORDER BY 'order'
SELECT `id` FROM `assignments` WHERE `assignments`.`course_id` = 1 AND is_group_lab = false ORDER BY 'order'
SELECT `id` FROM `assignments` WHERE `assignments`.`course_id` = 1 AND `assignments`.`is_group_lab` = true ORDER BY 'order'
```

## Debugging GORM issues

To debug a particular database issue, you can use the following environment variables to enable GORM logging:

```bash
LOG=1 LOGDB=4 go test -v -run TestGetSubmissionsByCourse
```

This will generate a lot of output, but you should be able to find the particular SQL query that is being generated for a particular GORM query.
