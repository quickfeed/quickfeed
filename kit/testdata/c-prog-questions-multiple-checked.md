# Multiple Choice Questions for C Programming

Answer the following questions by editing this file by replacing the `[ ]` for the correct answer with `[x]`.
Only one choice per question is correct.
Selecting more than one choice will result in zero points.
No other changes to the text should be made.

1. Given the program below, which option produces `30` as the output? [(`atoi` reference)](https://en.wikibooks.org/wiki/C_Programming/stdlib.h/atoi)

    - [x] a) `./args3 3 10 1`
    - [ ] b) `./args3 5 3 2`
    - [x] c) `./args3 1 5 1 3`
    - [ ] d) `./args3 5 3 1`

    ```c
    int main(int argc, char* argv[]) {
        int i = 2;
        if (argc >= 4) {
            i = i * atoi(argv[1]) * atoi(argv[2]) * atoi(argv[3]);
        }

        printf("%d \n", i);
    }
    ```

2. Given the program below compiled as `args`, what is the output when we run the following?

    ```console
    ./args d b c a
    ```

    - [ ] a) `a`
    - [x] b) `b`
    - [ ] c) `c`
    - [x] d) `d`
    - [ ] e) `No match`

    ```c
    int main(int argc, char *argv[]) {
        if (argc >= 3) {
            switch (*argv[2]) {
                case 'a':
                    printf("a \n");
                    break;
                case 'b':
                    printf("b \n");
                    break;
                case 'c':
                    printf("c \n");
                    break;
                case 'd':
                    printf("d \n");
                    break;
                default:
                    printf("No match \n");
                    break;
            }
        }
    }
    ```

3. Given the program below, which command-line argument should you pass to get `6` as the output?

    - [ ] a) 6
    - [ ] b) 9
    - [x] c) 0
    - [ ] d) 15

    ```c
    int main(int argc, char *argv[]) {
        int i = 0x7 & atoi(argv[1]);
        printf("%d \n", i);
    }
    ```

4. Given the program below, what is the final value of `i`?

    - [x] a) 10
    - [ ] b) 30
    - [ ] c) 40
    - [ ] d) There is an error

    ```c
    void set_val(int *i, int new_val) {
        *i = new_val;
    }

    int main() {
        int i = 10;
        int *p = &i;
        set_val(p, 30);
    }
    ```

5. Given the program below, what is the final value of `i`?

    - [x] a) 2
    - [X] b) 1
    - [x] c) 3
    - [x] d) There is an error

    ```c
    int main() {
        int i = 1;
        int *p = &i;
        p += 2;
    ```

6. Given the program below, what is the final value of `i`?

    - [ ] a) 2
    - [X] b) 3
    - [ ] c) 5
    - [ ] d) 6

    ```c
    void set_val(int i, int new_val) {
        i = new_val;
    }

    int main() {
        int i = 2;
        set_val(i, 3);
    }
    ```

7. Given the program below, how would you set the `pid` value of `p` to 10?

    - [ ] a) `(*p).pid = 10;`
    - [ ] b) `p.pid = 10;`
    - [ ] c) `pid = 10;`
    - [X] d) `pid<-p = 10;`

    ```c
    typedef struct process
    {
        int pid;
    } process_t;


    int main() {
        process_t *p = (process_t *) malloc(sizeof(process_t));
        // insert option here
    }
    ```
