#!/bin/bash

# resotre the project
wgo restore

# build the binaries
make p2p

# now that we have the binaries, lets build the docker images
if [ -z which docker ]; then
    echo "Docker is not installed, skipping building docker image"
    return 0
fi


DOCKER_MACHINE_STATUS=$(docker-machine status default)
if [ $DOCKER_MACHINE_STATUS == "Stopped" ]; then
    echo "Attempting to start 'default' docker machine"
    docker-machine start default
    EXIT=$?
    if [[ $EXIT -ne ]]; then
        echo "An error occurred while starting 'default' docker machine"
        exit $EXIT
    else
        docker-machine env default | bash
        echo "Successfully built docker image"
        return 0
    fi
fi
