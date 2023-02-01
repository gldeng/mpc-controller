GOPATH?=$(shell go env GOPATH)

.PHONY: lint-prepare
lint-prepare:
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh| sh -s latest

.PHONY: lint
lint:
	./bin/golangci-lint run -v ./...

.PHONY: fmt
fmt:
	@go fmt ./...

.PHONY: test
test:
	@go test -v -race ./...

.PHONY: abigen-prepare
abigen-prepare:
	mkdir -p bin && wget -c https://gethstore.blob.core.windows.net/builds/geth-alltools-linux-amd64-1.10.17-25c9b49f.tar.gz -O - | tar -xz -C bin
	wget -c https://github.com/ethereum/solidity/releases/download/v0.8.10/solc-static-linux -O bin/solc && chmod +x bin/solc

.PHONY: abijson-prepare
abijson-prepare:
	git submodule update --init --recursive --remote

.PHONY: abijson
abijson: abijson-prepare
	mkdir -p temp
	rm temp/combined.json
	bin/solc contract/contracts/src/MpcManager.sol --base-path contract/contracts/src --include-path contract/contracts/lib --combined-json abi -o temp

.PHONY: abigen
abigen: abijson
	bin/geth-alltools-linux-amd64-1.10.17-25c9b49f/abigen --combined-json temp/combined.json --pkg contract --out contract/MpcManager.go --type MpcManager