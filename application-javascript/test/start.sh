cd ../../../test-network
./network.sh up createChannel -ca
./network.sh deployCC -ccn auction -ccp ../double-auction/chaincode-go/ -ccl go -ccep "OR('Org1MSP.peer','Org2MSP.peer')"
cd ../double-auction/application-javascript/test
node ../enrollAdmin.js org1
node ../enrollAdmin.js org2
node ../registerEnrollUser.js org1 auctioneer
node ../initMarket.js org1 auctioneer
node ../createAuction.js org1 auctioneer 000
touch measure_bid.txt
touch measure_withdraw.txt
touch measure_score.txt
python generateBids.py
bash accountReg.sh
bash bidConfig.sh 000
for (( i = 1; i <= 100 ; i ++ ))
do
  node ../withdraw.js org1 buyer1 000
  node ../updateRating.js org1 buyer1 8 seller1
done
