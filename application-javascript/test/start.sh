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
  start=$(date +%s) 
  node ../withdraw.js org1 buyer1 000
  end=$(date +%s)
  time=`echo $start $end | awk '{print $2-$1}'`
  echo $time >> measure_withdraw.txt
  start=$(date +%s) 
  node ../updateRating.js org1 buyer1 8 seller1
  end=$(date +%s) 
  time=`echo $start $end | awk '{print $2-$1}'`
  echo $time >> measure_score.txt
done
