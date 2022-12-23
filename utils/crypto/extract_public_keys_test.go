package crypto

import (
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
	"testing"
)

var Keys = []string{
	"353fb105bbf9c29cbf46d4c93a69587ac478138b7715f0786d7ae1cc05230878",
	"7084300e7059ea4b308ec5b965ef581d3f9c9cd63714082ccf9b9d1fb34d658b",
	"5431ed99fbcc291f2ed8906d7d46fdf45afbb1b95da65fecd4707d16a6b3301b",
	"156177364ae1ca503767382c1b910463af75371856e90202cb0d706cdce53c33",
	"59d1c6956f08477262c9e827239457584299cf583027a27c1d472087e8c35f21",
	"6c326909bee727d5fc434e2c75a3e0126df2ec4f49ad02cdd6209cf19f91da33",
	"b17eac91d7aa2bd5fa72916b6c8a35ab06e8f0c325c98067bbc9645b85ce789f",
}

//
//var pubKeys = []string{
//	"3217bb0e66dda25bcd50e2ccebabbe599312ae69c76076dd174e2fc5fdae73d8bdd1c124d85f6c0b10b6ef24460ff4acd0fc2cd84bd5b9c7534118f472d0c7a1",
//	"72eab231c150b42e86cbe7398139432d2cad04289a820a922fe17b9d4ba577f4d3c33a90bd5b304344e1bea939ef7d16f428d50d25cada4225d9299d35ef1644",
//	"73ee5cd601a19cd9bb95fe7be8b1566b73c51d3e7e375359c129b1d77bb4b3e6f06766bde6ff723360cee7f89abab428717f811f460ebf67f5186f75a9f4288d",
//	"8196e06c3e803d0af06693a504ad14317550b4be4396ef57cf5f520c0f84833db8ed1056383ea329b8586cb62c37d80a3d7bb80742bc1bec6d650e6632a62905",
//	"c20e0c088bb20027a77b1d23ad75058df5349c7a2bfafff7516c44c6f69aa66defafb10f0932dc5c649debab82e6c816e164c7b7ad8abbe974d15a94cd1c2937",
//	"d0639e479fa1ca8ee13fd966c216e662408ff00349068bdc9c6966c4ea10fe3e5f4d4ffc52db1898fe83742a8732e53322c178acb7113072c8dc6f82bbc00b99",
//	"df7fb5bf5b3f97dffc98ecf8d660f604cad76f804a23e1b6cc76c11b5c92f3456dab26cdf995e6cb7cf772ba892044da9c64b095db7725d9e3c306c484cf54e2",
//}

var pubKeys = []string{
	"043217bb0e66dda25bcd50e2ccebabbe599312ae69c76076dd174e2fc5fdae73d8bdd1c124d85f6c0b10b6ef24460ff4acd0fc2cd84bd5b9c7534118f472d0c7a1",
	"0472eab231c150b42e86cbe7398139432d2cad04289a820a922fe17b9d4ba577f4d3c33a90bd5b304344e1bea939ef7d16f428d50d25cada4225d9299d35ef1644",
	"0473ee5cd601a19cd9bb95fe7be8b1566b73c51d3e7e375359c129b1d77bb4b3e6f06766bde6ff723360cee7f89abab428717f811f460ebf67f5186f75a9f4288d",
	"048196e06c3e803d0af06693a504ad14317550b4be4396ef57cf5f520c0f84833db8ed1056383ea329b8586cb62c37d80a3d7bb80742bc1bec6d650e6632a62905",
	"04c20e0c088bb20027a77b1d23ad75058df5349c7a2bfafff7516c44c6f69aa66defafb10f0932dc5c649debab82e6c816e164c7b7ad8abbe974d15a94cd1c2937",
	"04d0639e479fa1ca8ee13fd966c216e662408ff00349068bdc9c6966c4ea10fe3e5f4d4ffc52db1898fe83742a8732e53322c178acb7113072c8dc6f82bbc00b99",
	"04df7fb5bf5b3f97dffc98ecf8d660f604cad76f804a23e1b6cc76c11b5c92f3456dab26cdf995e6cb7cf772ba892044da9c64b095db7725d9e3c306c484cf54e2",
}

func TestExtractPubKeysForParticipants(t *testing.T) {
	pubKeys, err := ExtractPubKeysForParticipants(Keys)
	require.True(t, pubKeys != nil && err == nil)
	for _, k := range pubKeys {
		fmt.Printf("%x\n", k)
	}
}

func TestExtractPubKeysForParticipantsHex(t *testing.T) {
	pubKeys, err := ExtractPubKeysForParticipantsHex(Keys)
	require.True(t, pubKeys != nil && err == nil)
	for _, k := range pubKeys {
		fmt.Println(k)
	}
}

func TestUnmarshalPubKeyHex(t *testing.T) {
	for _, pubKeyStr := range pubKeys {
		pubKey, err := UnmarshalPubKeyHex(pubKeyStr)
		require.Nil(t, err)
		_ = pubKey

	}
}

func TestMarshalPubkey(t *testing.T) {
	for _, keyHexStr := range Keys {
		key, err := crypto.HexToECDSA(keyHexStr)
		require.Nil(t, err)
		pubKey := MarshalPubkey(&key.PublicKey)
		//fmt.Println(pubKey)
		fmt.Printf("%+x\n", pubKey)
		//results:
		/*
			043217bb0e66dda25bcd50e2ccebabbe599312ae69c76076dd174e2fc5fdae73d8bdd1c124d85f6c0b10b6ef24460ff4acd0fc2cd84bd5b9c7534118f472d0c7a1
			0472eab231c150b42e86cbe7398139432d2cad04289a820a922fe17b9d4ba577f4d3c33a90bd5b304344e1bea939ef7d16f428d50d25cada4225d9299d35ef1644
			0473ee5cd601a19cd9bb95fe7be8b1566b73c51d3e7e375359c129b1d77bb4b3e6f06766bde6ff723360cee7f89abab428717f811f460ebf67f5186f75a9f4288d
			048196e06c3e803d0af06693a504ad14317550b4be4396ef57cf5f520c0f84833db8ed1056383ea329b8586cb62c37d80a3d7bb80742bc1bec6d650e6632a62905
			04c20e0c088bb20027a77b1d23ad75058df5349c7a2bfafff7516c44c6f69aa66defafb10f0932dc5c649debab82e6c816e164c7b7ad8abbe974d15a94cd1c2937
			04d0639e479fa1ca8ee13fd966c216e662408ff00349068bdc9c6966c4ea10fe3e5f4d4ffc52db1898fe83742a8732e53322c178acb7113072c8dc6f82bbc00b99
			04df7fb5bf5b3f97dffc98ecf8d660f604cad76f804a23e1b6cc76c11b5c92f3456dab26cdf995e6cb7cf772ba892044da9c64b095db7725d9e3c306c484cf54e2
		*/
		//fmt.Println(pubKey[1:])
		//fmt.Printf("%+x\n", pubKey[1:])
	}
}
