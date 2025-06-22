### bitcoin-testnet-wallet is a project created by reason to get practice on generating keys, creating addresses and transactions. 
*There is only education content but structure and deployment practice closed to real project (for several exclusions)*

The main idea is create own bitcoin wallet to receiving/sending funds from/to the bitcoin testnet faucet

### Functional scheme
![alt text](images/image-4.png)

# Usage:
## It's necessary to change secretPassphrase in config/config.yaml with your version

## start wallet by following commands:
```bash
make build
make run
```

## Chose one of bitcoin testnet faucet:
https://bitcoinfaucet.uo1.net (more reliable and clearly)
https://cryptopump.info/send.php

## In wallet input command:
```bash
wallsh> wallet address
```
*It's your bitcoin address*

## Make sure that your available balance equal zero:
```bash
wallsh> wallet balance
```

## Put your bitcoin address into faucet and click "Send":
![alt text](images/image5.png)

## After few seconds check your "On hold" balance
You can see that your hold balance isn't equal zero. It points on that your transaction is already in mempool.

## After few minutes check your "Available" balance again
If it isn't equal zero, successful, you have received your first bitcoins

## To return some bitcoins back to faucet use command:
```bash
wallsh> wallet send <faucet_address> <amount>
```
For the https://bitcoinfaucet.uo1.net faucet address is published on main page: 
![alt text](images/image6.png)

### After sending you can also check your wallet balance for understanding moving your funds