//go:generate abigen --sol src/AvaLido.sol --pkg contract --out AvaLido.go --type AvaLido
//go:generate abigen --sol src/MpcManager.sol --pkg contract --out MpcManager.go --type MpcManager

package contract

// Compatible abigen version: 1.10.17-stable. Source code download links:
// https://github.com/ethereum/go-ethereum/archive/refs/tags/v1.10.17.zip
// https://github.com/ethereum/go-ethereum/archive/refs/tags/v1.10.17.tar.gz
