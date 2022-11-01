import argparse
import sqlite3
from functools import partial
from typing import List
from faker import Faker


class Statement:
    def __init__(self, select, update, value_functions):
        self.select = select
        self.update = update
        self.value_functions = value_functions

    def values(self):
        # Generate fake values for each column in the row
        # The resulting tuple contains values returned by
        # calling each function in the value_functions list
        values = tuple()
        for value_function in self.value_functions:
            values += (value_function(),)
        return values


class DatabaseAnonymizer:
    """
    The DatabaseAnonymizer class is used to anonymize the database. The updateList contains a list of tuples, each tuple containing:
    - A select statement to fetch the rows to update
    - An update statement to update the rows
    - A list of functions to generate fake data for each column in the row

    Using the updateList, the anonymize method fetches the rows to update, generates fake data for each column in the row, and updates the rows.

    To add a new table to anonymize, add a new tuple to the updateList.
    """

    def __init__(self, db: str):
        self.fake = Faker()
        self.conn = sqlite3.connect(db)
        self.cur = self.conn.cursor()
        self.updateCur = self.conn.cursor()
        # User to exclude from anonymization
        self.excludeUser = 0

    def get_update_list(self) -> List[Statement]:
        # List of tuples (select statement, update statement, functions to generate fake data)
        return [
            Statement(
                "SELECT DISTINCT organization_id FROM repositories",
                "UPDATE repositories SET organization_id=? WHERE organization_id=?",
                (self.fake.random_number,),
            ),
            Statement(
                "SELECT * FROM repositories",
                "UPDATE repositories SET html_url=?, repository_id=? WHERE id=?",
                (
                    self.fake.url,
                    partial(self.fake.random_number, digits=8),
                ),
            ),
            Statement(
                f"SELECT * FROM users WHERE id != {self.excludeUser}",
                "UPDATE users SET name=?, email=?, login=?, student_id=?, avatar_url=? WHERE id=?",
                (
                    self.fake.name,
                    self.fake.email,
                    self.fake.user_name,
                    partial(self.fake.random_number, digits=6),
                    self.fake.url,
                ),
            ),
            Statement(
                f"SELECT * FROM remote_identities WHERE user_id != {self.excludeUser}",
                "UPDATE remote_identities SET access_token=?, remote_id=? WHERE id=?",
                (
                    partial(self.fake.password, length=20),
                    partial(self.fake.random_number, digits=6),
                ),
            ),
            Statement(
                "SELECT * FROM groups",
                "UPDATE groups SET name=? WHERE id=?",
                (self.fake.slug,),
            ),
        ]

    def fetch(self, statement: str) -> list:
        self.cur.execute(statement)
        rows = self.cur.fetchall()
        return rows

    def close(self):
        self.conn.close()

    def anonymize(self):
        for statement in self.get_update_list():
            rows = self.fetch(statement.select)
            for row in rows:
                # the first selected column is passed to the WHERE clause
                where_clause = (row[0],)
                self.updateCur.execute(
                    statement.update, statement.values() + where_clause
                )
                self.conn.commit()

    def set_as_admin(self, login: str):
        self.updateCur.execute("UPDATE users SET is_admin=1 WHERE login=?", (login,))
        self.conn.commit()

    def set_remote_identity(self, user_id: int, remote_id: int):
        self.updateCur.execute(
            "UPDATE remote_identities SET remote_id=? WHERE user_id=?",
            (remote_id, user_id),
        )
        self.conn.commit()

    def exclude_user(self, login: str) -> bool:
        self.cur.execute("SELECT id FROM users WHERE login=?", (login,))
        user = self.cur.fetchone()
        if user is None:
            return False
        print(f"Excluding user {login} with id {user[0]} from anonymization")
        self.excludeUser = user[0]
        return True

    def get_user_id(self, login: str):
        self.cur.execute("SELECT id FROM users WHERE login=?", (login,))
        user = self.cur.fetchone()
        if user is None:
            print(f"User {login} not found")
            return
        print(f"User {login} has id {user[0]}")


def main():
    parser = argparse.ArgumentParser(description="Database anonymizer")
    parser.add_argument(
        "--database",
        dest="database",
        type=str,
        help="Name of the database file to anonymize",
        required=True,
    )
    parser.add_argument(
        "--anonymize", dest="anonymize", type=bool, help="Anonymize the database"
    )
    parser.add_argument(
        "--exclude",
        dest="exclude",
        type=str,
        metavar="LOGIN",
        help="User to exclude from anonymization. Only used if --anonymize is set",
    )
    parser.add_argument(
        "--admin",
        dest="admin",
        type=str,
        metavar="LOGIN",
        help="Set the user with the given login as admin",
    )
    parser.add_argument(
        "--remote",
        dest="remote",
        type=int,
        nargs=2,
        metavar=("USER_ID", "REMOTE_ID"),
        help="Set the remote identity of the user with the given USER_ID to the given REMOTE_ID",
    )
    parser.add_argument(
        "--user-id",
        dest="user_id",
        type=str,
        metavar="LOGIN",
        help="Get the user id of the user with the given login",
    )
    args = parser.parse_args()

    if args.database is None:
        parser.print_help()
        return

    db = DatabaseAnonymizer(args.database)

    if args.user_id:
        db.get_user_id(args.user_id)
        return

    if args.exclude is not None:
        if db.exclude_user(args.exclude) is False:
            # Abort if the user to exclude does not exist
            print(f"User {args.exclude} not found, aborting")
            return

    if args.anonymize:
        db.anonymize()

    if args.admin is not None:
        db.set_as_admin(args.admin)

    if args.remote is not None:
        db.set_remote_identity(args.remote[0], args.remote[1])

    db.close()


if __name__ == "__main__":
    main()
