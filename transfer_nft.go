package main

import (
	"fmt"
	"os"

	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/joho/godotenv"
)

func main() {

	//Loads the .env file and throws an error if it cannot load the variables from that file corectly
	err := godotenv.Load(".env")
	if err != nil {
		panic(fmt.Errorf("Unable to load enviroment variables from .env file. Error:\n%v\n", err))
	}

	//Grab your testnet account ID and private key from the .env file
	myAccountId, err := hedera.AccountIDFromString(os.Getenv("MY_ACCOUNT_ID"))
	if err != nil {
		panic(err)
	}

	myPrivateKey, err := hedera.PrivateKeyFromString(os.Getenv("MY_PRIVATE_KEY"))
	if err != nil {
		panic(err)
	}

	//Print your testnet account ID and private key to the console to make sure there was no error
	fmt.Printf("The account ID is = %v\n", myAccountId)
	fmt.Printf("The private key is = %v\n", myPrivateKey)

	//Create your testnet client
	client := hedera.ClientForTestnet()
	client.SetOperator(myAccountId, myPrivateKey)

	//Create a treasury Key
	treasuryKey, err := hedera.GeneratePrivateKey()
	treasuryPublicKey := treasuryKey.PublicKey()

	//Create treasury account
	treasuryAccount, err := hedera.NewAccountCreateTransaction().
		SetKey(treasuryPublicKey).
		SetInitialBalance(hedera.NewHbar(10)).
		Execute(client)
	
	//Get the receipt of the transaction
	receipt, err := treasuryAccount.GetReceipt(client)

	//Get the account ID
	treasuryAccountId := *receipt.AccountID

	//Alice Key
	aliceKey, err := hedera.GeneratePrivateKey()
	alicePublicKey := aliceKey.PublicKey()

	//Create Alice's account
	aliceAccount, err := hedera.NewAccountCreateTransaction().
		SetKey(alicePublicKey).
		SetInitialBalance(hedera.NewHbar(10)).
		Execute(client)
	
	//Get the receipt of the transaction
	receipt2, err := aliceAccount.GetReceipt(client)

	//Get the account ID
	aliceAccountId := *receipt2.AccountID

	//Create a supply key
	supplyKey, err := hedera.GeneratePrivateKey()

	//Create the NFT
	nftCreate, err := hedera.NewTokenCreateTransaction().
		SetTokenName("diploma").
		SetTokenSymbol("GRAD").
		SetTokenType(hedera.TokenTypeNonFungibleUnique).
		SetDecimals(0).
		SetInitialSupply(0).
		SetTreasuryAccountID(treasuryAccountId).
		SetSupplyType(hedera.TokenSupplyTypeFinite).
		SetMaxSupply(250).
		SetSupplyKey(supplyKey).
		FreezeWith(client)

		//Sign the transaction with the treasury key
	nftCreateTxSign := nftCreate.Sign(treasuryKey)

	//Submit the transaction to a Hedera network
	nftCreateSubmit, err := nftCreateTxSign.Execute(client)

	//Get the transaction receipt
	nftCreateRx, err := nftCreateSubmit.GetReceipt(client)

	//Get the token ID
	tokenId := *nftCreateRx.TokenID

	//Log the token ID
	fmt.Println("Created NFT with token ID", tokenId)

	// IPFS content identifiers for which we will create a NFT
	CID := "QmTzWcVfk88JRqjTpVwHzBeULRTNzHY7mnBSG42CpwHmPa"

	// Minet new NFT
	mintTx, err := hedera.NewTokenMintTransaction().
		SetTokenID(tokenId).
		SetMetadata([]byte(CID)).
		FreezeWith(client)

	//Sign the transaction with the supply key
	mintTxSign := mintTx.Sign(supplyKey)

	//Submit the transaction to a Hedera network
	mintTxSubmit, err := mintTxSign.Execute(client)

	//Get the transaction receipt
	mintRx, err := mintTxSubmit.GetReceipt(client)

	//Log the serial number
	fmt.Println("Created NFT", tokenId, "with serial:", mintRx.SerialNumbers)

	//Create the associate transaction
	associateAliceTx, err := hedera.NewTokenAssociateTransaction().
		SetAccountID(aliceAccountId).
		SetTokenIDs(tokenId).
		FreezeWith(client)

	//Sign with Alice's key
	signTx := associateAliceTx.Sign(aliceKey)

	//Submit the transaction to a Hedera network
	associateAliceTxSubmit, err := signTx.Execute(client)

	//Get the transaction receipt
	associateAliceRx, err := associateAliceTxSubmit.GetReceipt(client)

	//Confirm the transaction was successful
	fmt.Println("NFT association with Alice's account:", associateAliceRx.Status)

	// Check the balance before the transfer for the treasury account
	balanceCheckTreasury, err := hedera.NewAccountBalanceQuery().SetAccountID(treasuryAccountId).Execute(client)
	fmt.Println("Treasury balance:", balanceCheckTreasury.Tokens, "NFTs of ID", tokenId)

	// Check the balance before the transfer for Alice's account
	balanceCheckAlice, err := hedera.NewAccountBalanceQuery().SetAccountID(aliceAccountId).Execute(client)
	fmt.Println("Alice's balance:", balanceCheckAlice.Tokens, "NFTs of ID", tokenId)

	// Transfer the NFT from treasury to Alice
	tokenTransferTx, err := hedera.NewTransferTransaction().
		AddNftTransfer(hedera.NftID{TokenID: tokenId, SerialNumber: 1}, treasuryAccountId, aliceAccountId).
		FreezeWith(client)
	
	// Sign with the treasury key to authorize the transfer
	signTransferTx := tokenTransferTx.Sign(treasuryKey)

	tokenTransferSubmit, err := signTransferTx.Execute(client)
	tokenTransferRx, err := tokenTransferSubmit.GetReceipt(client)

	fmt.Println("NFT transfer from Treasury to Alice:", tokenTransferRx.Status)

	// Check the balance of the treasury account after the transfer
	balanceCheckTreasury2, err := hedera.NewAccountBalanceQuery().SetAccountID(treasuryAccountId).Execute(client)
	fmt.Println("Treasury balance:", balanceCheckTreasury2.Tokens, "NFTs of ID", tokenId)

	// Check the balance of Alice's account after the transfer
	balanceCheckAlice2, err := hedera.NewAccountBalanceQuery().SetAccountID(aliceAccountId).Execute(client)
	fmt.Println("Alice's balance:", balanceCheckAlice2.Tokens, "NFTs of ID", tokenId)

}
	