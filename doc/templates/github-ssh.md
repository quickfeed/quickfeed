# GitHub: SSH Authentication

The steps below are meant for more experienced GitHub users.
They are necessary when you want to use SSH instead HTTPS for GitHub authentication.

1. Register your public SSH key (typically `~/.ssh/id_rsa.pub`) at [GitHub](https://github.com/settings/ssh).

2. Run the following command:

   ```console
   git config --global url."git@github.com:".insteadOf https://github.com/
   ```

   This will make Git rewrite GitHub URLs to use SSH instead of HTTPS.
   This "hack" is necessary since it is not possible to specify the `go get` tool to use SSH authentication.

## Problems with Access Rights

If you have multiple GitHub users, you may need to clear the cache.
