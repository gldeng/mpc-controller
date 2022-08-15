//go:generate abigen --sol src/AvaLidoMock.sol --pkg contract --out AvaLidoMock.go --type AvaLidoMock
//go:generate abigen --sol src/MpcManager.sol --pkg contract --out MpcManager.go --type MpcManager
//go:generate abigen --sol src/OracleManager.sol --pkg contract --out OracleManager.go --type OracleManager
//go:generate abigen --sol src/Oracle.sol --pkg contract --out Oracle.go --type Oracle

package contract

// Compatible abigen version: 1.10.17-stable. Source code download links:
// https://github.com/ethereum/go-ethereum/archive/refs/tags/v1.10.17.zip
// https://github.com/ethereum/go-ethereum/archive/refs/tags/v1.10.17.tar.gz
