package main

import (
	"fmt"
	"os"

	"github.com/hashgraph/hedera-sdk-go/v2"
	"github.com/joho/godotenv"
)

func main() {

	//LOADS THE .ENV FILE AND THROWS AN EROOR IF IT CANNOT LOAD THE VARIABLES
	err := godotenv.Load(".env")
	if err != nil {
		panic(fmt.Errorf("Unable to load enviroment variables from .env file. Error:\n%v\n", err))
	}

	//GRAB YOUR TESTNET ACCOUNT ID AND KEY FROMZ THE .ENV FILE
	myAccountId, err := hedera.AccountIDFromString(os.Getenv("MY_ACCOUNT_ID"))
	if err != nil {
		panic(err)
	}

	myPrivateKey, err := hedera.PrivateKeyFromString(os.Getenv("MY_PRIVATE_KEY"))
	if err != nil {
		panic(err)
	}

	//PRINT ACCOUNT ID AND KEY TO MAKE SURE THERE WASN'T AN ERROR READING FROM THE .ENV FILE
	fmt.Printf("The account ID is = %v\n", myAccountId)
	fmt.Printf("The private key is = %v\n", myPrivateKey)

	//CREATE TESTNET CLIENT
	client := hedera.ClientForTestnet()
	client.SetOperator(myAccountId, myPrivateKey)

	//CREATE TREASURY KEY
	treasuryKey, err := hedera.GeneratePrivateKey()
	treasuryPublicKey := treasuryKey.PublicKey()

	//CREATE TREASURY ACCOUNT
	treasuryAccount, err := hedera.NewAccountCreateTransaction().
		SetKey(treasuryPublicKey).
		SetInitialBalance(hedera.NewHbar(10)).
		Execute(client)
	
	//GET THE RECEIPT OF THE TRANSACTION
	receipt, err := treasuryAccount.GetReceipt(client)

	//GET THE ACCOUNT ID
	treasuryAccountId := *receipt.AccountID

	//ALICE'S KEY
	aliceKey, err := hedera.GeneratePrivateKey()
	alicePublicKey := aliceKey.PublicKey()

	//CREATE AILICE'S ACCOUNT
	aliceAccount, err := hedera.NewAccountCreateTransaction().
		SetKey(alicePublicKey).
		SetInitialBalance(hedera.NewHbar(10)).
		Execute(client)

	//GET THE RECEIPT OF THE TRANSACTION
	receipt2, err := aliceAccount.GetReceipt(client)

	//GET ALICE'S ACCOUNT ID
	aliceAccountId := *receipt2.AccountID

	//CREATE SUPPLY KEY
	supplyKey, err := hedera.GeneratePrivateKey()

	//CREATE FUNGIBLE TOKEN (STABLECOIN)
	tokenCreateTx, err := hedera.NewTokenCreateTransaction().
		SetTokenName("USD Bar").
		SetTokenSymbol("USDB").
		SetTokenType(hedera.TokenTypeFungibleCommon).
		SetDecimals(2).
		SetInitialSupply(10000).
		SetTreasuryAccountID(treasuryAccountId).
		SetSupplyType(hedera.TokenSupplyTypeInfinite).
		SetSupplyKey(supplyKey).
		FreezeWith(client)

	//SIGN WITH TREASURY KEY
	tokenCreateSign := tokenCreateTx.Sign(treasuryKey)

	//SUBMIT THE TRANSACTION
	tokenCreateSubmit, err := tokenCreateSign.Execute(client)

	//GET THE TRANSACTION RECEIPT
	tokenCreateRx, err := tokenCreateSubmit.GetReceipt(client)

	//GET THE TOKEN ID
	tokenId := *tokenCreateRx.TokenID

	//LOG THE TOKEN ID TO THE CONSOLE
	fmt.Println("Created fungible token with token ID", tokenId)

	//TOKEN ASSOCIATION WITH ALICE's ACCOUNT
	associateAliceTx, err := hedera.NewTokenAssociateTransaction().
		SetAccountID(aliceAccountId).
		SetTokenIDs(tokenId).
		FreezeWith(client)

	//SIGN WITH ALICE'S KEY TO AUTHORIZE THE ASSOCIATION
	signTx := associateAliceTx.Sign(aliceKey)

	//SUBMIT THE TRANSACTION
	associateAliceTxSubmit, err := signTx.Execute(client)

	//GET THE RECEIPT OF THE TRANSACTION
	associateAliceRx, err := associateAliceTxSubmit.GetReceipt(client)

	//LOG THE TRANSACTION STATUS
	fmt.Println("STABLECOIN token association with Alice's account:", associateAliceRx.Status)

	//Check the balance before the transfer for the treasury account
	balanceCheckTreasury, err := hedera.NewAccountBalanceQuery().SetAccountID(treasuryAccountId).Execute(client)
	fmt.Println("Treasury balance:", balanceCheckTreasury.Tokens, "units of token ID", tokenId)

	//Check the balance before the transfer for Alice's account
	balanceCheckAlice, err := hedera.NewAccountBalanceQuery().SetAccountID(aliceAccountId).Execute(client)
	fmt.Println("Alice's balance:", balanceCheckAlice.Tokens, "units of token ID", tokenId)

	//Transfer the STABLECOIN from treasury to Alice
	tokenTransferTx, err := hedera.NewTransferTransaction().
		AddTokenTransfer(tokenId, treasuryAccountId, -2500).
		AddTokenTransfer(tokenId, aliceAccountId, 2500).
		FreezeWith(client)

	//SIGN WITH THE TREASURY KEY TO AUTHORIZE THE TRANSFER
	signTransferTx := tokenTransferTx.Sign(treasuryKey)

	//SUBMIT THE TRANSACTION
	tokenTransferSubmit, err := signTransferTx.Execute(client)

	//GET THE TRANSACTION RECEIPT
	tokenTransferRx, err := tokenTransferSubmit.GetReceipt(client)

	fmt.Println("Token transfer from Treasury to Alice:", tokenTransferRx.Status)

	//CHECK THE BALANCE AFTER THE TRANSFER FOR THE TREASURY ACCOUNT
	balanceCheckTreasury2, err := hedera.NewAccountBalanceQuery().SetAccountID(treasuryAccountId).Execute(client)
	fmt.Println("Treasury balance:", balanceCheckTreasury2.Tokens, "units of token", tokenId)

	//CHECK THE BALANCE AFTER THE TRANSFER FOR ALICE'S ACCOUNT
	balanceCheckAlice2, err := hedera.NewAccountBalanceQuery().SetAccountID(aliceAccountId).Execute(client)
	fmt.Println("Alice's balance:", balanceCheckAlice2.Tokens, "units of token", tokenId)

}
