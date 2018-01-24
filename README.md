# Badger Boi
## Running
### Recommended
#### 1. Build images for all services that need building
```
  make build
```

#### 2. Start postgres container
In a separate shell:
```
  make postgres
```

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
