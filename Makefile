TAG_VERSION?=0.1.0

## usage: TAG_VERSION=0.1.0 make build
build:
	docker build -t testnet-wallet:$(TAG_VERSION) -f deployment/Dockerfile .

## usage: TAG_VERSION=0.1.0 make run
run:
	docker run -it --rm testnet-wallet:${TAG_VERSION}
