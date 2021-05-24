# double-auction
Double auction chaincode is saved in chaincode-go. Application code is saved in application-javascript. After test, time cost result will be saved in three txt files: measure_bid.txt, measure_withdraw.txt, measure_score.txt.<br>
## Install
1. Fork hyperledger/fabric-samples project and make sure you can run test-network <br>
2. Download our project and copy it to fabric-sample/ <br>
3. cd fabric-sample/hyperledger-fabric-double-auction/application-javascript <br>
4. npm init <br>
5. npm install <br>
6. npm install fabric-ca-client && npm install fabric-network && npm install perf_hooks<br>
7. npm audit fix<br>
## Start 100 buyers and 100 sellers test
cd test/ && bash start.sh<br>
## Reset all
cd test/ && bash reset.sh<br>
## Application
bid.js: submit bids<br>
enrollAdmin.js: enroll org CA admin<br>
registerEnrollUser.js: register a user in CA<br>
registerAccount.js: register an account for trading<br>
initMarket.js: init feedback system<br>
queryFeedback.js: query current ratings<br>
updateRating.js: submit a score for a seller<br>
createAuction.js: create an auction<br>
queryAuction.js: query an auction<br>
withdraw.js: get final allocation and payment<br>
closeAuction.js: close an auction<br>

