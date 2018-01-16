#!/bin/bash

# we might need to use docker import or something to pull in the golang image

docker build -t `cat IMAGE_NAME` .
