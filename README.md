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
  make updb
```

#### 3. Start web server container
In a separate shell:
```
  make upweb
```
#### 4. Start job server container
In a separate shell:
```
  make upjob
```

## Code reloading
The go code will be automatically recompiled inside the running containers using fresh.
