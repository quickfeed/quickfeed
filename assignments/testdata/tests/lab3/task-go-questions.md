# Multiple Choice Questions about Go Programming

Answer the following questions by editing this file by replacing the `[ ]` for the correct answer with `[x]`.
Only one choice per question is correct.
Selecting more than one choice will result in zero points.
No other changes to the text should be made.

1. What is the zero value of an integer (`int`)?

    - [ ] a) `1`
    - [ ] b) `nil`
    - [ ] c) `0`
    - [ ] d) `NaN`

2. What is the zero value of a slice?

    - [ ] a) `[_]`
    - [ ] b) `[0]`
    - [ ] c) `nil`
    - [ ] d) `null`

3. Which of these is a correct way of declaring a slice?

    - [ ] a) `slice := [4]int{4, 5, 2, 1}`
    - [ ] b) `slice := {4, 5, 2, 1}int`
    - [ ] c) `slice := [4, 5, 2, 1]int`
    - [ ] d) `slice := []int{4, 5, 2, 1}`

4. Given the `Person` struct below, which of these is the correct way of creating an instance of `Person`?

    - [ ] a) `var p Person{firstName: "Johnny", shortName: "Bravo", age: 43}`
    - [ ] b) `p := new Person(firstName: "Johnny", shortName: "Bravo", age: 43)`
    - [ ] c) `p := Person{firstName: "Johnny", shortName: "Bravo", age: 43}`
    - [ ] d) `p Person := {firstName: "Johnny", shortName: "Bravo", age: 43}`

    ```go
    type Person struct {
        firstName, shortName string
        age int
    }
    ```

5. Which is the correct way to create a map `m` with the initial values `{"a": 1, "b": 2, "c": 3}`?

    - [ ] a) `m := make(map[string]int{"a": 1, "b": 2, "c": 3})`
    - [ ] b) `m := map[string]int{"a": 1, "b": 2, "c": 3}`
    - [ ] c) `m := make(map[string]int, []string{"a", "b", "c"}, []int{1, 2, 3})`
    - [ ] d) `m := {"a": 1, "b": 2, "c": 3}`

6. How would you range over a slice?

    - [ ] a) `for a, b in range slice {}`
    - [ ] b) `for each a : slice {}`
    - [ ] c) `for a, b := range slice {}`
    - [ ] d) `for a := range slice[a] {}`

7. Given a slice of integers, named `sli`. How would you append a number to this slice?

    - [ ] a) `sli.append(2)`
    - [ ] b) `sli = append(sli, 2)`
    - [ ] c) `sli[len(sli)] = 2`
    - [ ] d) `append(sli, 2)`
    - [ ] e) `sli += 2`

8. Which condition below checks if the map `m` contains the key `b`?

    - [ ] a) `if _, hasKey := m["b"]; hasKey {`
    - [ ] b) `if "b" in m {`
    - [ ] c) `if m.hasKey("b") {`
    - [ ] d) `if m["b"] {`

9. Given the `ChessPiece` interface and `Bishop` struct below.
   How would you implement the `ChessPiece` interface on the piece `Bishop`?

    - [ ] a) add the phrase `implements ChessPiece` after `struct`, then add the interface methods below the struct
    - [ ] b) write methods with the same names as defined in the `ChessPiece` interface with `Bishop` as the receiver
    - [ ] c) write the required functions from the `ChessPiece` interface directly inside the `Bishop` struct
    - [ ] d) none of the above

    ```go
    type ChessPiece interface {
        Move(x, y, int)
        GetPos() (x, y int)
    }

    type Bishop struct {
        posX, posY int
    }
    ```

10. What does this program print?

    ```go
    var s = "¡¡¡Hello, Gophers!!!"
    s = strings.TrimSuffix(s, ", Gophers!!!")
    s = strings.TrimPrefix(s, "¡¡¡")
    fmt.Print(s)
    ```

    - [ ] a) `Hello`
    - [ ] b) `Hello, ¡¡¡`
    - [ ] c) `¡¡¡Hello, Gophers!!!`
    - [ ] d) `, Gophers!!!Hello¡¡¡`

11. What is the type of `pi` in `pi := fmt.Sprintf("%.2f", 3.1415)`?

    - [ ] a) `int`
    - [ ] b) `float`
    - [ ] c) `float64`
    - [ ] d) `string`
