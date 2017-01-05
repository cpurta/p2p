all : p2p docker

p2p:
	wgo build -o p2p && mv p2p ./bin

docker:
	docker build -f ./docker/Dockerfile p2p:latest .
