/*
 * Copyright Xuyang Ma. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

'use strict';

const { Gateway, Wallets } = require('fabric-network');
const path = require('path');
const performance = require('perf_hooks').performance;
const fs = require('fs')
const { exit } = require('process');
const { buildCCPOrg1, buildCCPOrg2, buildWallet} = require('../../test-application/javascript/AppUtil.js');

const myChannel = 'mychannel';
const myChaincodeName = 'auction';

async function withdraw(ccp,wallet,user,auctionID) {
	try {
		const gateway = new Gateway();
		await gateway.connect(ccp,
			{ wallet: wallet, identity: user, discovery: { enabled: true, asLocalhost: true } });

		const network = await gateway.getNetwork(myChannel);
		const contract = network.getContract(myChaincodeName);

		let statefulTxn = contract.createTransaction('Withdraw');

		console.log('\n--> Submit Transaction: Withdraw');
		let result = await statefulTxn.submit(auctionID, user);
		console.log('*** Result: committed' + result);
		gateway.disconnect();
	} catch (error) {
		console.error(`******** FAILED to submit bid: ${error}`);
		if (error.stack) {
			console.error(error.stack);
		}
		process.exit(1);
	}
}

async function main() {
	try {
		const start = performance.now();
		if (process.argv[2] === undefined || process.argv[3] === undefined || process.argv[4] === undefined) {
			console.log('Usage: node withdraw.js org userID auctionID');
			process.exit(1);
		}

		const org = process.argv[2];
		const user = process.argv[3];
		const auctionID = process.argv[4];

		if (org === 'Org1' || org === 'org1') {
			const ccp = buildCCPOrg1();
			const walletPath = path.join(__dirname, 'wallet/org1');
			const wallet = await buildWallet(Wallets, walletPath);
			await withdraw(ccp,wallet,user,auctionID);
		}
		else if (org === 'Org2' || org === 'org2') {
			const ccp = buildCCPOrg2();
			const walletPath = path.join(__dirname, 'wallet/org2');
			const wallet = await buildWallet(Wallets, walletPath);
			await withdraw(ccp,wallet,user,auctionID);
		}  else {
			console.log('Usage: node withdraw.js org userID auctionID');
			console.log('Org must be Org1 or Org2');
		}
		const end = performance.now();
		fs.appendFile('measure_withdraw.txt', `${(end - start)/1000}`, err => {
			if (err) {
			  console.error(err)
			  return
			}
		})
	} catch (error) {
		console.error(`******** FAILED to run the application: ${error}`);
		process.exit(1);
	}
}

main();
