## Running

###Recommended
#### Build all services
```
  make build
```

#### Start cassandra
In a separate shell:
```
  make upcassandra
```
Then wait until cassandra outputs that it is listening.
#### Start transaction server
In a separate shell:
```
  make uptx
```
#### Start web server
In a separate shell:
```
  make upweb
```

#### Reloading a container after making code changes
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
Create seed script for cassandra
