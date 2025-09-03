# Recommended Editors for Writing Go Code

## Visual Studio Code (VSCode)

For writing Go code (and other languages too) we highly recommend using the cross-platform editor **[Visual Studio Code](https://code.visualstudio.com/)**.
The teaching staff mainly use VSCode for Go programming.

### Installing and Setting Up VSCode

Simply follow the instructions to install the program for your desired system.

For Go language support, use the **[Go extension](https://code.visualstudio.com/docs/languages/go)**, which gives you things such as intellisense (autocompletion, etc.) and linter (check code for errors).
You install the Go extension from the marketplace within VSCode.

Another useful VSCode extension is [Code Runner](https://marketplace.visualstudio.com/items?itemName=formulahendry.code-runner), which allows to run code using a keyboard shortcut or right-clicking a file instead of using ``go run``.
Runs code by default in a read-only editor.

### Configuring VSCode

To configure VSCode, open the settings by pressing `Ctrl+,` (or `Cmd+,` on macOS).
You can also open the settings by clicking the gear icon in the lower left corner of the window.

Please use the following settings:

```json
  "go.useLanguageServer": true,
  "gopls": {
    "formatting.gofumpt": true,
    "ui.semanticTokens": true,
    "build.directoryFilters": ["-public", "-dev", "-doc"],
    // Add parameter placeholders when completing a function.
    "usePlaceholders": false,
    // If true, enable additional analyses with staticcheck.
    // Warning: This will significantly increase memory usage.
    "staticcheck": true
  },
```

### Developing in WSL with VSCode

If you are developing with WSL on Windows you can use VSCode for interacting with the WSL environment.
The VSCode documentation has [detailed instructions](https://code.visualstudio.com/docs/remote/wsl) for this use case.

## GoLand

[GoLand](https://www.jetbrains.com/go/) is a commercial IDE specially designed for the Go language.
As a student you can create a [free student user account](https://www.jetbrains.com/community/education/?fromMenu), and thus use GoLand for free.

Some of GoLand's features include:

* Excellent refactoring support.
* On-the-fly error detection and suggestion for fixes.
* Navigation & Search.
* Run & Debug code without extra work.

## Other Editors

If you prefer some other editor there exists Go support for many editors, such as Atom, Emacs, and vim.
The Go wiki maintains a [comprehensive list](https://go.dev/wiki/IDEsAndTextEditorPlugins) of several IDEs and text editors that can be used for Go programming.

Whichever editor you choose, it is highly recommended that you configure it to use the [`gofumpt`](https://github.com/mvdan/gofumpt) tool.
This will reformat your code to follow the Go style, and make sure that all the necessary import statements are inserted (so you don’t need to write the import statement when you start using a new package.)
The `gofumpt` tool is compatible with most editors, but may require some configuration.
Using the Go plugin for VSCode should automatically configure `gofumpt`.

Note that editors may also be able to run your code within the editor itself, but it may require some configuration.
However, using the go tool from the command line is often times preferred.

### Basic vi/vim Usage

Sometimes it may be convenient to edit files using terminal-based editors, e.g. if you need to edit a file on a server via a SSH connection.
For these cases, we recommend vi or vim.
vi is typically preinstalled on most Unix systems, such as embedded systems (e.g. routers), set-top-boxes, etc.
vim (Vi IMproved) is a more powerful version of vi, which has many useful features and can be customized via configuration files and adding extensions.
vi/vim have a bit of learning curve compared to other editors, but many software developers find these to be very productive once you have gotten used to them.
If you intend to write most code in a remote environment or in a terminal, we recommend learning to use vim.

There are many detailed tutorials on vi/vim online, e.g. [here](http://www.washington.edu/computing/unix/vi.html).
Below is a short primer on vim.

To modify/write the file `file.go` with vim:

```console
vim file.go
```

To exit vim without saving, use the command `:q!`.
To exit vim and save any changes, use the command: `:wq`.
To save any changes and continue editing, use the command: `:w`.

vim has two modes of operation.
When you open vim you start in *command mode*.
In command mode key presses are interpreted as shortcuts.
For example,

* `j` moves the cursor one line down,
* `k` moves it one line up,
* `h` moves one character to the left, and
* `l` moves one character to the right.

You could also use the arrow keys.
If you press `i` from command mode you enter the *insert mode*.
In this mode you can type, remove text with the `Backspace` and `Delete` keys, add new lines with the `Enter` key, and navigate with the arrow keys.
Press `Escape` to return to command mode from insert mode.

Other useful *command mode* features in vim:

* If you enter `:` you can enter some additional commands. For example:
  * `:w` will save the file,
  * `:q` will quit vim,
  * `:wq` will save and quit vim, and
  * `:q!` will discard modifications and quit vim.
  * `:set nu` will enable line numbers.
* If you enter `/` you can search for some matching string.
  E.g. `/word` will find the first match for `word` within the text (starting from the cursor position).
  Go to the next match by pressing `n` and the previous match by pressing `N`.
* Some vim commands use "verbs" and "actions".
  For example, by pressing `d` you start the action to delete something.
  If you type `d3w` you will execute the action "delete 3 words".
  If you enter the "verb" two times in a row it will generally affect the whole line, e.g. `dd` deletes the current line.
  Note that special characters such as `-` are interpreted as a word, e.g. "a-b" counts as 3 words and can be deleted with `d3w`.
* Use the `y` (yank) "verb" to copy some text.
  * `y3w` copies 3 words starting from the cursor.
  * `yy` copies the whole line.
* Use `p` or `P` to paste.
  * `p` pastes the text to the line below.
  * `P` pastes the text above the current line.
* Use `o` or `O` to insert a new line.
  * `o` inserts a new line below and enters insert mode.
  * `O` inserts a new line above and enters insert mode.
* Use `u` to undo the last action.
  This can be repeated to undo several actions.
* `b` moves the cursor one word back and `w` moves the cursor one word forward.

Other useful *insert mode* features in vim:

* If you have marked or copied some text, you can copy or paste it with `Ctrl+Shift+c` and `Ctrl+Shift+v`, respectively.

vimtutor is a tool which opens in vim and contains a beginner's guide to vim.
To enter vimtutor:

```console
vimtutor
```

*NOTE: To exit vimtutor you have to use vim commands (either `:q` or `:q!`).*
