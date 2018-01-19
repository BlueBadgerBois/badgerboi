# Badger Boi
## Running
### Recommended
#### 1. Build images for all services that need building
```
  make build
```

#### 2. Start cassandra container
In a separate shell:
```
  make upcassandra
```


Optional: If this is your first time running, you'll need to bootstrap the database:
Wait until cassandra outputs that it is listening (something like `Starting listening for CQL clients on /0.0.0.0:9042`),
then do:
```
  ./scripts/seed.sh
```
This creates the keyspace and creates the users table.

#### 3. Start transaction server container
In a separate shell:
```
  make uptx
```
#### 4. Start web server container
In a separate shell:
```
  make upweb
```

#### 5. Run pending migrations
```
  ./scripts/migrate up
```

## Creating a migration
```
  ./scripts/migrate create my_fancy_migration
```
This will create two files in db/migrations: an "up" file and a "down" file.
You are responsible for adding the appropriate cql commands in these files.

The "up" file should contain the change you are making to the schema e.g. adding a new table.
The "down" file should contain the commands for REVERSING that change e.g. dropping the added table.

## Code reloading
The go code will be automatically recompiled inside the running containers using fresh.
