/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

'use strict';

const { Gateway, Wallets } = require('fabric-network');
const path = require('path');
const performance = require('perf_hooks').performance;
const fs = require('fs')
const { buildCCPOrg1, buildCCPOrg2, buildWallet, prettyJSONString} = require('../../test-application/javascript/AppUtil.js');

const myChannel = 'mychannel';
const myChaincodeName = 'auction';

async function updateRating(ccp,wallet,user,score,target) {
	try {

		const gateway = new Gateway();

		//connect using Discovery enabled
		await gateway.connect(ccp,
			{ wallet: wallet, identity: user, discovery: { enabled: true, asLocalhost: true } });

		const network = await gateway.getNetwork(myChannel);
		const contract = network.getContract(myChaincodeName);

		let statefulTxn = contract.createTransaction('UpdateRating');

		console.log('\n--> Submit Transaction: Score a seller');
		let result = await statefulTxn.submit(score, user, target);
		console.log('*** Result: committed'+result);

		gateway.disconnect();
	} catch (error) {
		console.error(`******** FAILED to submit bid: ${error}`);
	}
}

async function main() {
	try {
		const start = performance.now();
		if (process.argv[2] === undefined || process.argv[3] === undefined ||
            process.argv[4] === undefined || process.argv[5] === undefined) {
			console.log('Usage: node updateRating.js org userID score sellerID');
			process.exit(1);
		}

		const org = process.argv[2];
		const user = process.argv[3];
		const score = process.argv[4];
		const target = process.argv[5];

		if (org === 'Org1' || org === 'org1') {
			const ccp = buildCCPOrg1();
			const walletPath = path.join(__dirname, 'wallet/org1');
			const wallet = await buildWallet(Wallets, walletPath);
			await updateRating(ccp,wallet,user,score,target);
		}
		else if (org === 'Org2' || org === 'org2') {
			const ccp = buildCCPOrg2();
			const walletPath = path.join(__dirname, 'wallet/org2');
			const wallet = await buildWallet(Wallets, walletPath);
			await updateRating(ccp,wallet,user,score,target);
		}  else {
			console.log('Usage: node updateRating.js org userID score sellerID');
			console.log('Org must be Org1 or Org2');
		}
		const end = performance.now();
		fs.appendFile('measure_score.txt', `${(end - start)/1000}`, err => {
			if (err) {
			  console.error(err)
			  return
			}
		})
	} catch (error) {
		console.error(`******** FAILED to run the application: ${error}`);
	}
}


main();
