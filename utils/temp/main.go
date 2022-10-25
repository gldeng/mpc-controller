package main

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
)

func main() {
	//configControllerContracts()
	trimNEWPrefix()
	fmt.Println("\n")
	newContracts()
}

func trimNEWPrefix() {
	newStr := `NEW_AVALIDO_ADDRESS="0x8de5bc4c2471e16836232c07ec85a2c34c168367"
NEW_VALIDATOR_SELECTOR_ADDRESS="0xfffa76e0fff88ee3e9d5f1437113a620cf1708ed"
NEW_ORACLE_ADDRESS="0x5e492f8785561cf4b962357b06d02d16ee6b2995"
NEW_ORACLE_MANAGER_ADDRESS="0xe261d7406e77dc356f320a38cd1e78a57d489422"
NEW_MPC_MANAGER_ADDRESS="0x22fc30eb48542d3cc554aa2c75bc4a81967678a0"`

	m1 := regexp.MustCompile(`NEW_`)

	fmt.Println(m1.ReplaceAllString(newStr, ""))
}

func newContracts() {
	contracts := `Deployed AvaLido, 0x173c4ff1d268d351151cae5b06c86c790e23f4c2
  Deployed Validator Selector, 0x8214e4c18ba7dfac24970b08a5132e6b0432bfce
  Deployed Oracle, 0xeeaec956dbf75a086270b50d6bde3fef6e9917af
  Deployed Oracle Manager, 0x60f4d3c020947e596028c296187cc7c5d09c5f86
  Deployed MPC Manager, 0x22fc30eb48542d3cc554aa2c75bc4a81967678a0

`

	var addrs []string

	contractsSplit := strings.Split(contracts, "\n")
	for _, contract := range contractsSplit[:5] {
		addr := strings.Split(contract, ",")
		addrs = append(addrs, strings.TrimPrefix(addr[1], " "))
	}

	fmt.Printf("NEW_AVALIDO_ADDRESS=%q\n"+
		"NEW_VALIDATOR_SELECTOR_ADDRESS=%q\n"+
		"NEW_ORACLE_ADDRESS=%q\n"+
		"NEW_ORACLE_MANAGER_ADDRESS=%q\n"+
		"NEW_MPC_MANAGER_ADDRESS=%q\n", addrs[0], addrs[1], addrs[2], addrs[3], addrs[4])
}

func configControllerContracts() {
	fileBytes, err := ioutil.ReadFile("/home/zealy/Rockx/mpc-controller/tests/docker/mpc-controller/configs/controller1.yaml")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(fileBytes))

	m1 := regexp.MustCompile(`(?<=mpcManagerAddress).*$`)

	fmt.Println(m1.ReplaceAllString(string(fileBytes), ""))
}
