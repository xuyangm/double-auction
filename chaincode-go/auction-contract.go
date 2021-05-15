/*
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type AuctionContract struct {
	contractapi.Contract
}

func (c *AuctionContract) AuctionExists(ctx contractapi.TransactionContextInterface, auctionID string) (bool, error) {
	data, err := ctx.GetStub().GetState(auctionID)
	if err != nil {
		return false, err
	}
	return data != nil, nil
}

func (c *AuctionContract) RatingExists(ctx contractapi.TransactionContextInterface, addr string) (bool, error) {
	data, err := ctx.GetStub().GetState("auctioneer")
	if err != nil {
		return false, fmt.Errorf("no feedback system")
	}
	feedback := new(Feedback)
	err = json.Unmarshal(data, feedback)
	if err != nil {
		return false, fmt.Errorf("could not unmarshal world state data to type Feedback")
	}
	index := GetFeedbackIndex(addr, feedback)
	if index == -1 {
		return false, err
	}
	return true, nil
}

func (c *AuctionContract) InitFeedbackSystem(ctx contractapi.TransactionContextInterface, addr string) error {
	if addr != "auctioneer" {
		return fmt.Errorf("illegal user %s tries to build feedback system", addr)
	}
	feedback := new(Feedback)
	bytes, _ := json.Marshal(feedback)
	ctx.GetStub().PutState(addr, bytes)
	accounts := new(Accounts)
	bytes, _ = json.Marshal(accounts)
	return ctx.GetStub().PutState("acc", bytes)
}

func (c *AuctionContract) RegisterAccount(ctx contractapi.TransactionContextInterface, addr string) error {
	bytes, _ := ctx.GetStub().GetState("acc")
	accounts := new(Accounts)
	err := json.Unmarshal(bytes, accounts)
	if err != nil {
		return fmt.Errorf("could not unmarshal world state data to type Auction")
	}
	accounts.Address = append(accounts.Address, Hash(addr))
	accounts.Balance = append(accounts.Balance, 10000)
	bytes, _ = json.Marshal(accounts)
	err = ctx.GetStub().PutState("acc", bytes)
	return err
}

// CreateAuction creates a new instance of Auction
func (c *AuctionContract) CreateAuction(ctx contractapi.TransactionContextInterface, auctionID string) error {
	exists, err := c.AuctionExists(ctx, auctionID)
	if err != nil {
		return fmt.Errorf("could not read from world state. %s", err)
	} else if exists {
		return fmt.Errorf("the auction %s already exists", auctionID)
	}

	auction := new(Auction)
	auction.Closed = false
	bytes, _ := json.Marshal(auction)
	return ctx.GetStub().PutState(auctionID, bytes)
}

// QueryAuction retrieves an instance of Auction from the world state
func (c *AuctionContract) QueryAuction(ctx contractapi.TransactionContextInterface, auctionID string) (string, error) {
	exists, err := c.AuctionExists(ctx, auctionID)
	if err != nil {
		return "", fmt.Errorf("could not read from world state. %s", err)
	} else if !exists {
		return "", fmt.Errorf("the auction %s does not exist", auctionID)
	}

	bytes, _ := ctx.GetStub().GetState(auctionID)
	auction := new(Auction)
	err = json.Unmarshal(bytes, auction)

	if err != nil {
		return "", fmt.Errorf("could not unmarshal world state data to type Auction")
	}

	return string(bytes), nil
}

// DeleteAuction deletes an instance of Auction from the world state
func (c *AuctionContract) CloseAuction(ctx contractapi.TransactionContextInterface, auctionID string) error {
	exists, err := c.AuctionExists(ctx, auctionID)
	if err != nil {
		return fmt.Errorf("could not read from world state. %s", err)
	} else if !exists {
		return fmt.Errorf("the auction %s does not exist", auctionID)
	}

	return ctx.GetStub().DelState(auctionID)
}

func (c *AuctionContract) Bid(ctx contractapi.TransactionContextInterface, auctionID string, prices string, times string, quantities string, addr string) error {
	exists, err := c.AuctionExists(ctx, auctionID)
	if err != nil {
		return fmt.Errorf("could not read from world state. %s", err)
	} else if !exists {
		return fmt.Errorf("no %s auction", auctionID)
	}

	bytes, _ := ctx.GetStub().GetState(auctionID)
	auction := new(Auction)
	json.Unmarshal(bytes, auction)
	AddBid(Hash(addr), StrToIntArr(times), StrToFloatArr(prices), StrToIntArr(quantities), auction)
	newBytes, _ := json.Marshal(auction)

	return ctx.GetStub().PutState(auctionID, newBytes)
}

func (c *AuctionContract) Withdraw(ctx contractapi.TransactionContextInterface, auctionID string, addr string) ([]string, error) {
	res := []string{}
	exists, err := c.AuctionExists(ctx, auctionID)
	if err != nil {
		return res, fmt.Errorf("could not read from world state. %s", err)
	} else if !exists {
		return res, fmt.Errorf("no %s auction", auctionID)
	}

	bytes, _ := ctx.GetStub().GetState(auctionID)
	auction := new(Auction)
	json.Unmarshal(bytes, auction)

	rbytes, _ := ctx.GetStub().GetState("auctioneer")
	feedback := new(Feedback)
	err = json.Unmarshal(rbytes, feedback)

	if !auction.Closed {
		ratings := [100]float64{}
		for i := 0; i < len(auction.Sellers[0]); i++ {
			tmpSeller := auction.Sellers[0][i].Address
			index := GetFeedbackIndex(tmpSeller, feedback)
			if index != -1 {
				ratings[i] = feedback.Ratings[index]
			} else {
				feedback.Address = append(feedback.Address, tmpSeller)
				feedback.Ratings = append(feedback.Ratings, 6.0)
				feedback.ResourceVolumes = append(feedback.ResourceVolumes, 0.0)
				feedback.NewResourceVolumes = append(feedback.NewResourceVolumes, 0.0)
				ratings[i] = 6.0
			}
		}

		winners, reservPrices, supplyMatrix := Allocate(auction.Buyers, auction.Sellers)
		buyersPay, sellersPay, supplyMatrix := DeterminePayment(auction.Buyers, auction.Sellers, winners, reservPrices, supplyMatrix, ratings[:])
		for i := 0; i < 100; i++ {
			for j := 0; j < 100; j++ {
				for k := 0; k < 3; k++ {
					auction.Allocation[i][j][k] = supplyMatrix[i][j][k]
				}
			}
		}

		bytes, _ = ctx.GetStub().GetState("acc")
		accounts := new(Accounts)
		err = json.Unmarshal(bytes, accounts)
		if err != nil {
			return res, fmt.Errorf("could not unmarshal world state data to type Auction")
		}

		for i := 0; i < len(auction.Buyers); i++ {
			baddr := auction.Buyers[i].Address
			auction.BuyersPay[i] = buyersPay[i]
			ChangeBalance(baddr, -buyersPay[i], accounts)
		}
		for i := 0; i < len(auction.Sellers[0]); i++ {
			saddr := auction.Sellers[0][i].Address
			auction.SellersPay[i] = sellersPay[i]
			ChangeBalance(saddr, -sellersPay[i], accounts)
		}
		bytes, _ = json.Marshal(accounts)
		ctx.GetStub().PutState("acc", bytes)
	}

	index := FindBuyer(Hash(addr), auction.Buyers)
	for j := 0; j < len(auction.Sellers[0]); j++ {
		for k := 0; k < 3; k++ {
			if auction.Allocation[index][j][k] > 0 {
				str := "seller " + auction.Sellers[0][j].Address + " provides " + fmt.Sprint(auction.Allocation[index][j][k]) + " VM" + fmt.Sprint(k) + " to buyer " + auction.Buyers[index].Address
				res = append(res, str)
			}
		}
	}
	w := [3]int{1, 2, 4}
	for i := 0; i < len(auction.Buyers); i++ {
		for j := 0; j < len(auction.Sellers[0]); j++ {
			for k := 0; k < 3; k++ {
				if auction.Allocation[i][j][k] > 0 {
					feedback.NewResourceVolumes[GetFeedbackIndex(auction.Sellers[0][j].Address, feedback)] = float64(auction.Allocation[i][j][k] * w[k] * auction.Buyers[i].Time)
				}
			}
		}
	}
	newBytes, _ := json.Marshal(feedback)
	ctx.GetStub().PutState("auctioneer", newBytes)

	return res, err
}

func (c *AuctionContract) UpdateRating(ctx contractapi.TransactionContextInterface, score string, addr string, target string) (string, error) {
	exists, err := c.RatingExists(ctx, Hash(target))
	if err != nil {
		return "", fmt.Errorf("could not read from world state. %s", err)
	} else if !exists {
		return "", fmt.Errorf("no rating for %s", target)
	}

	bytes, _ := ctx.GetStub().GetState("auctioneer")
	feedback := new(Feedback)
	json.Unmarshal(bytes, feedback)
	index := GetFeedbackIndex(Hash(target), feedback)
	rsc := feedback.NewResourceVolumes[index]
	sc := StrToFloatArr(score)[0]
	feedback.Ratings[index] = (feedback.Ratings[index]*feedback.ResourceVolumes[index] + sc*rsc) / (feedback.ResourceVolumes[index] + rsc)
	feedback.ResourceVolumes[index] += rsc
	bytes, _ = json.Marshal(feedback)

	return string(bytes), ctx.GetStub().PutState("auctioneer", bytes)
}

func (c *AuctionContract) ViewFeedback(ctx contractapi.TransactionContextInterface, addr string) (string, error) {
	if addr != "auctioneer" {
		return "", fmt.Errorf("no right to check ratings")
	}

	bytes, _ := ctx.GetStub().GetState("auctioneer")
	feedback := new(Feedback)
	err := json.Unmarshal(bytes, feedback)
	if err != nil {
		return "", fmt.Errorf("could not unmarshal world state data to type Feedback")
	}
	return string(bytes), err
}
