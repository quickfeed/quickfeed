from functools import partial
from faker import Faker
import argparse
import sqlite3


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
        # List of tuples (select statement, update statement, functions to generate fake data)
        self.updateList = [
            ("SELECT DISTINCT organization_id FROM repositories", "UPDATE repositories SET organization_id=? WHERE organization_id=?", (self.fake.random_number,)),
            ("SELECT * FROM repositories", "UPDATE repositories SET html_url=?, repository_id=? WHERE id=?", (self.fake.url, partial(self.fake.random_number, digits=8),)),
            ("SELECT * FROM users", "UPDATE users SET name=?, email=?, login=?, student_id=?, avatar_url=? WHERE id=?", (self.fake.name, self.fake.email, self.fake.user_name, partial(self.fake.random_number, digits=6), self.fake.url,)),
            ("SELECT * FROM remote_identities", "UPDATE remote_identities SET access_token=?, remote_id=? WHERE id=?", (partial(self.fake.password, length=20), partial(self.fake.random_number, digits=6),)),
            ("SELECT * FROM groups", "UPDATE groups SET name=? WHERE id=?", (self.fake.slug,)),
        ]

    def fetch(self, statement: str) -> list:
        self.cur.execute(statement)
        rows = self.cur.fetchall()
        return rows

    def close(self):
        self.conn.close()

    def anonymize(self):
        for statement in self.updateList:
            rows = self.fetch(statement[0])
            for row in rows:
                # Generate fake data for each column in the row (row[0] is the value of the first column, passed to the WHERE clause)
                self.updateCur.execute(statement[1], tuple(func() for func in statement[2]) + (row[0],))
                self.conn.commit()


def main():
    parser = argparse.ArgumentParser(description='Database anonymizer')
    parser.add_argument('--database', dest='database', type=str, help='Name of the database file to anonymize')
    args = parser.parse_args()
    
    if args.database is None:
        parser.print_help()
        return
    
    db = DatabaseAnonymizer(args.database)
    db.anonymize()
    db.close()

if __name__ == "__main__":
    main()
