package addrs

import (
	"fmt"
	crypto2 "github.com/avalido/mpc-controller/utils/crypto"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
	"testing"
)

var Keys = []string{
	"59d1c6956f08477262c9e827239457584299cf583027a27c1d472087e8c35f21",
	"6c326909bee727d5fc434e2c75a3e0126df2ec4f49ad02cdd6209cf19f91da33",
	"5431ed99fbcc291f2ed8906d7d46fdf45afbb1b95da65fecd4707d16a6b3301b",
}

func TestPrivateKeyToAddress(t *testing.T) {
	for _, k := range Keys {
		privKey, err := crypto.HexToECDSA(k)
		require.Nil(t, err)

		addr := crypto.PubkeyToAddress(privKey.PublicKey)
		fmt.Println(addr.Hex())
		/*
			0x3051bA2d313840932B7091D2e8684672496E9A4B
			0x7Ac8e2083E3503bE631a0557b3f2A8543EaAdd90
			0x3600323b486F115CE127758ed84F26977628EeaA
		*/
	}
}

func TestPubKeyHexToAddress(t *testing.T) {
	pubkeyHex := "03384ebafc6f500033058392a5a85438e011b9556486a6687e167a93b307ac1116"
	addr, err := PubKeyHexToAddress(pubkeyHex)

	require.Nil(t, err)
	require.Equal(t, "0x00F4CA03D670b47e47CCA5087dF363f4d5C2A955", addr.String())
}

func TestEthPubkeyHexToAddress(t *testing.T) {
	ethPubKey := "6e7007dd52295f38a0d72546399a7ed64fd88386009fbfd8859e5d61b68cdae04037ddfb7ac737e456c2c1bc327239d2afa4bfd4d2cd213f744230fcb6bde27f"
	addr, err := EthPubkeyHexToAddress(ethPubKey)
	require.Nil(t, err)

	expected := "0xe840A89F2f875c6952F88424975161F89D3f2C05"
	got := addr.String() // or addr.Hex()
	require.Equal(t, expected, got)

	fmt.Println(got) // output: 0xD35Ba5D5d264defc89A90327B90A43212a3D37bd
}

func TestDenormizedPubKeyHexToAccount(t *testing.T) {
	dnmPubkeyHex := "378e06ecb5c799909cb7048c091720b6b780773a46504315b183eee53d57f8f0fa59fe1698fa4c32817568fb05faac2d2a777497c49d395b5d17556188f64070"
	//x := dnmPubkeyHex[:64]
	//for _, v := range x {
	//	fmt.Printf("%T", v)
	//}
	//require.Equal(t, x, "02aa647003c0aad7e754180f9a9a6f77cfa91efe3c0a9424fe65e84d0adf9452f3")

	//addr, err := PubKeyHexToAddress(dnmPubkeyHex)
	//require.Nil(t, err)
	//fmt.Println(addr)

	// Retract the compressed value
	b := common.Hex2Bytes(dnmPubkeyHex)
	fmt.Println(b)
	// [170 100 112 3 192 170 215 231 84 24 15 154 154 111 119 207 169 30 254 60 10 148 36 254 101 232 77 10 223 148 82 243 225 191 25 74 252 61 20 155 43 47 143 116 251 212 102 155 204 194 50 92 20 12 75 16 205 4 12 39 163 201 28 102]

	fmt.Println(b[:32])

	v := make([]byte, 33)

	v[0] = 2
	copy(v[1:], b[:32])

	//v = append(v, b[:32]...)
	fmt.Println(v)
	fmt.Println(common.Bytes2Hex(v) == "02aa647003c0aad7e754180f9a9a6f77cfa91efe3c0a9424fe65e84d0adf9452f3")

	//UnmarshalPubKeyHex(common.Bytes2Hex(v))
	addr, err := PubKeyHexToAddress(common.Bytes2Hex(v))
	require.Nil(t, err)
	fmt.Println(addr)
}

func TestBytesToAddress(t *testing.T) {
	for _, k := range Keys {
		privKey, err := crypto.HexToECDSA(k)
		require.Nil(t, err)

		addr := crypto.PubkeyToAddress(privKey.PublicKey)

		// Implementation of crypto.PubkeyToAddress:
		/*
			pubBytes := crypto.FromECDSAPub(&privKey.PublicKey)
			addr := common.BytesToAddress(crypto.Keccak256(pubBytes[1:])[12:])
		*/
		fmt.Println(addr.Hex())
		/*
			0x3051bA2d313840932B7091D2e8684672496E9A4B
			0x7Ac8e2083E3503bE631a0557b3f2A8543EaAdd90
			0x3600323b486F115CE127758ed84F26977628EeaA
		*/
	}

	// ---
	genPubKeyHex := "03322e8792ab970e485b860c6645f191c249f9446ac13d2e0765e5472557b2de83"
	b := common.Hex2Bytes(genPubKeyHex)
	fmt.Println(b)

	genPubKey, err := crypto2.UnmarshalPubKeyHex(genPubKeyHex)
	require.Nil(t, err)

	genPubBytes := crypto.FromECDSAPub(genPubKey)
	fmt.Println(genPubBytes)
	addr := common.BytesToAddress(crypto.Keccak256(genPubBytes[1:])[12:])
	fmt.Println(addr.Hex())
	// 0x4cc6F9f25648186DB4e60eCa342B01947c9D8A5d

	require.Equal(t, "0x4cc6F9f25648186DB4e60eCa342B01947c9D8A5d", "0x4cc6F9f25648186DB4e60eCa342B01947c9D8A5d")
}
