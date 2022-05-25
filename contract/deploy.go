package contract

//
//var Gas int64 = 8000000
//var BaseFee int64 = 300_000_000_000
//
//// todo: consider remove this function, which has been moved into utils/contract package.
//
//func Deploy(chainId *big.Int, client *ethclient.Client, privKey *ecdsa.PrivateKey, bytecodeJSON string) (*common.Address, *types.Receipt, error) {
//	account := crypto.PubkeyToAddress(privKey.PublicKey)
//	nonce, err := client.NonceAt(context.Background(), account, nil)
//	if err != nil {
//		return nil, nil, errors.Wrapf(err, "failed to query nonce for account %q.", account)
//	}
//
//	bytecodeBytes := common.Hex2Bytes(bytecodeJSON)
//
//	txdata := &types.DynamicFeeTx{
//		ChainID:    chainId,
//		Nonce:      nonce,
//		To:         nil,
//		Gas:        uint64(Gas),
//		GasFeeCap:  big.NewInt(BaseFee), // maxgascost = 2.1ether
//		GasTipCap:  big.NewInt(1),
//		AccessList: nil,
//		Data:       bytecodeBytes,
//	}
//	tx := types.NewTx(txdata)
//	signer := types.LatestSignerForChainID(chainId)
//	signature, err := crypto.Sign(signer.Hash(tx).Bytes(), privKey)
//	if err != nil {
//		return nil, nil, errors.Wrapf(err, "failed to sign transaction %v with private key %v.", tx, privKey)
//	}
//
//	txSigned, err := tx.WithSignature(signer, signature)
//	if err != nil {
//		return nil, nil, errors.Wrap(err, "failed to generate new signed transaction")
//	}
//
//	err = client.SendTransaction(context.Background(), txSigned)
//	if err != nil {
//		return nil, nil, errors.Wrap(err, "failed to send transaction")
//	}
//	txHash := txSigned.Hash()
//
//	time.Sleep(5 * time.Second)
//	rcp, err := client.TransactionReceipt(context.Background(), txHash)
//	if err != nil {
//		return nil, nil, errors.Wrapf(err, "failed to query transaction receipt for txHash %q.", txHash.Hex())
//	}
//	logger.Debug("Deployed a smart contract",
//		logger.Field{"txHash", txHash.Hex()},
//		logger.Field{"receipt", rcp})
//
//	return &rcp.ContractAddress, rcp, nil
//}
