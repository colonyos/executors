# Pre
Create a .pgpass file in the home directory of the account that pg_dump will run as.
hostname:port:database:username:password

Then, set the file's mode to 0600. Otherwise, it will be ignored.
chmod 600 ~/.pgpass
