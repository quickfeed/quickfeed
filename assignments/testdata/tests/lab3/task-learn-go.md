# Go Exercises

Before you start working on the assignments below, make sure that your local working copy has all the latest changes from the course [assignments](https://github.com/COURSE_TAG/assignments) repository.
Instructions for fetching the latest changes are [here](https://github.com/COURSE_TAG/info/blob/main/lab-submission.md#update-local-working-copy-from-course-assignments).

1. In the following, we will use `sequence/triangular.go` exercise as an example.
   The file contains the following skeleton code and task description:

    ```golang
    package sequence

    // Task 1: Triangular numbers
    //
    // triangular(n) returns the n-th Triangular number, and is defined by the
    // recurrence relation F_n = n + F_n-1 where F_0=0 and F_1=1
    //
    // Visualization of numbers:
    // n = 1:    n = 2:     n = 3:      n = 4:    etc...
    //   o         o          o           o
    //            o o        o o         o o
    //                      o o o       o o o
    //                                 o o o o
    func triangular(n uint) uint {
        return 0
    }
    ```

2. Implement the function body according to the specification so that all the tests in `sequence/triangular_test.go` passes.
   The test file looks like this:

    ```golang
    package sequence

    import (
        "testing"

        "github.com/google/go-cmp/cmp"
    )

    var triangularTests = []struct {
        in, want uint
    }{
        {0, 0},
        {1, 1},
        {2, 3},
        {3, 6},
        {4, 10},
        {5, 15},
        {6, 21},
        {7, 28},
        {8, 36},
        {9, 45},
        {10, 55},
        {20, 210},
    }

    func TestTriangular(t *testing.T) {
        for _, test := range triangularTests {
            if diff := cmp.Diff(test.want, triangular(test.in)); diff != "" {
                t.Errorf("triangular(%d): (-want +got):\n%s", test.in, diff)
            }
        }
    }
    ```

3. There are several ways to run the tests. If you run:

   ```console
   go test
   ```

   the Go tool will run all tests found in files whose file name ends with `_test.go` (in the current directory).
   Similarly, you can also run a specific test as follows:

   ```console
   go test -run TestTriangular
   ```

4. You should ***not*** edit files or code that are marked with a `// DO NOT EDIT` comment.
   Please make separate `filename_test.go` files if you wish to write and run your own tests.

5. When you have completed a task and sufficiently many local tests pass, you may push your code to GitHub.
   This will trigger QuickFeed which will then run a separate test suite on your code.

   Using `sequence/triangular.go` as an example, use the following procedure to commit and push your changes to GitHub and QuickFeed:

    ```console
    $ git add triangular.go
    $ git commit
    // This will open an editor for you to write a commit message
    // Use for example "Implemented Assignment 2"
    $ git push
    ```

6. QuickFeed will now build and run a test suite on the code you submitted.
   You can check the output by going to the [QuickFeed web interface](https://uis.itest.run).
   The results (build log) is available from the Labs menu.
   Note that the results shows output for all the tests in current lab assignment.
   You will want to focus on the output for the specific test results related to the task you're working on.

7. Follow the same process for the other tasks included in this lab assignment.
   Each task contains a single `.go` template file, along with a task description and a `_test.go` file with tests.

8. Finally, complete the task in `cmd/terminal/main.go`.

   *Note: You must print the current working directory at the start of every prompt.*
   For example, if the working directory is `/home/username`, your prompt should be something like the following:

   ```console
   /home/username>
   ```

   Then we can input commands after the prompt, e.g. the `ls` command:

   ```console
   /home/username> ls
   ```

   You can use the [`os.Getwd` function](https://golang.org/pkg/os/#Getwd) to find the current working directory.
   This could be implemented in a function named `printPrompt()`.

9. When you are done with all assignments and want to submit the final version, please follow these [instructions](https://github.com/COURSE_TAG/info/blob/main/lab-submission.md#final-submission-of-labx).
