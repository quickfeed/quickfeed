# Database Tool

## Installation

    pip install -r requirements.txt

## Example usage

```bash
    # Shows available options
    python3 main.py --help
    # Anonymizes the database
    python3 main.py --database /path/to/database/file.db --anonymize true
    # Anonymizes the database, but excludes anonymization of the user with login `superuser`
    python3 main.py --database /path/to/database/file.db --anonymize true --exclude superuser
    # Sets the admin status of the user with login `superuser` to true
    python3 main.py --database /path/to/database/file.db --admin superuser
    # Prints the ID of the user with login `superuser`
    python3 main.py --database /path/to/database/file.db --user-id superuser
    # Sets the remote ID of the user with ID 1 to `123456789`
    python3 main.py --database /path/to/database/file.db --remote 1 123456789
```
