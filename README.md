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
Then wait until cassandra outputs that it is listening (something like `Starting listening for CQL clients on /0.0.0.0:9042`).

If this is your first time running, you'll need to bootstrap the database:
```
  ./scripts/seed.sh
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

#### Reloading a container after making code changes (e.g. after changing the web server)
This will re-build the corresponding image and start a new container with the image, attached to the docker network.

In the shell that the server was running in:
```
  make reloadtx
```
or
```
  make reloadweb
```

### TODO
Improve seed script for cassandra
