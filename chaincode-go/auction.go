/*
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"crypto/sha1"
	"encoding/hex"
	"math"
	"sort"
	"strconv"
	"strings"
)

type BuyerBid struct {
	Address    string  `json:"address"`
	Price      float64 `json:"price"`
	Time       int     `json:"time"`
	Quantities []int   `json:"quantity"`
}

type SellerBid struct {
	Address    string    `json:"address"`
	Prices     []float64 `json:"price"`
	Times      []int     `json:"time"`
	Quantities []int     `json:"quantity"`
}

type Auction struct {
	Closed     bool             `json:"closed"`
	Buyers     []BuyerBid       `json:"buyersBid"`
	Sellers    [3][]SellerBid   `json:"sellersBid"`
	Allocation [100][100][3]int `json:"allocation"`
	BuyersPay  [100]float64     `json:"buyersPay"`
	SellersPay [100]float64     `json:"sellersPay"`
}

type Accounts struct {
	Address []string  `json:"address"`
	Balance []float64 `json:"balance"`
}

type Feedback struct {
	Address            []string  `json:"address"`
	Ratings            []float64 `json:"ratings"`
	ResourceVolumes    []float64 `json:"resourceVolumes"`
	NewResourceVolumes []float64 `json:"newResourceVolumes"`
}

type ByDensity []BuyerBid

func (a ByDensity) Len() int { return len(a) }
func (a ByDensity) Less(i, j int) bool {
	if a[i].Price*a[i].Price/float64(a[i].Quantities[0]+2*a[i].Quantities[1]+4*a[i].Quantities[2]) > a[j].Price*a[j].Price/float64(a[j].Quantities[0]+2*a[j].Quantities[1]+4*a[j].Quantities[2]) {
		return true
	} else {
		return false
	}
}
func (a ByDensity) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

type ByPrice1 []SellerBid

func (a ByPrice1) Len() int           { return len(a) }
func (a ByPrice1) Less(i, j int) bool { return a[i].Prices[0] < a[j].Prices[0] }
func (a ByPrice1) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

type ByPrice2 []SellerBid

func (a ByPrice2) Len() int           { return len(a) }
func (a ByPrice2) Less(i, j int) bool { return a[i].Prices[1] < a[j].Prices[1] }
func (a ByPrice2) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

type ByPrice3 []SellerBid

func (a ByPrice3) Len() int           { return len(a) }
func (a ByPrice3) Less(i, j int) bool { return a[i].Prices[2] < a[j].Prices[2] }
func (a ByPrice3) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func AddBid(addr string, times []int, prices []float64, quantities []int, auction *Auction) {
	if len(times) == 1 {
		b := new(BuyerBid)
		b.Address = addr
		b.Price = prices[0]
		b.Time = times[0]
		b.Quantities = append(b.Quantities, quantities[0])
		b.Quantities = append(b.Quantities, quantities[1])
		b.Quantities = append(b.Quantities, quantities[2])
		auction.Buyers = append(auction.Buyers, *b)
		sort.Sort(ByDensity(auction.Buyers))
	} else {
		b := new(SellerBid)
		b.Address = addr
		b.Prices = append(b.Prices, prices[0])
		b.Prices = append(b.Prices, prices[1])
		b.Prices = append(b.Prices, prices[2])
		b.Times = append(b.Times, times[0])
		b.Times = append(b.Times, times[1])
		b.Times = append(b.Times, times[2])
		b.Quantities = append(b.Quantities, quantities[0])
		b.Quantities = append(b.Quantities, quantities[1])
		b.Quantities = append(b.Quantities, quantities[2])
		auction.Sellers[0] = append(auction.Sellers[0], *b)
		auction.Sellers[1] = append(auction.Sellers[1], *b)
		auction.Sellers[2] = append(auction.Sellers[2], *b)
		sort.Sort(ByPrice1(auction.Sellers[0]))
		sort.Sort(ByPrice2(auction.Sellers[1]))
		sort.Sort(ByPrice3(auction.Sellers[2]))
	}
}

func Allocate(buyers []BuyerBid, sellers [3][]SellerBid) ([100]bool, [100]float64, [100][100][3]int) {
	buyersNum := len(buyers)
	sellersNum := len(sellers[0])
	winners := [100]bool{}
	reservPrices := [100]float64{}
	costMatrix := [100][3]int{}
	supplyMatrix := [100][100][3]int{}
	for i := 0; i < 100; i++ {
		for j := 0; j < 100; j++ {
			for k := 0; k < 3; k++ {
				supplyMatrix[i][j][k] = 0
				if i == 0 {
					costMatrix[j][k] = 0
					if k == 0 {
						winners[j] = false
						reservPrices[j] = 0
					}
				}
			}
		}
	}
	for i := 0; i < buyersNum; i++ {
		tmpSupply := [100][3]int{}
		granted := true
		for k := 0; k < 3; k++ {
			unallocated := buyers[i].Quantities[k]
			for j := 0; j < sellersNum; j++ {
				availableResource := sellers[k][j].Quantities[k] - costMatrix[j][k]
				if unallocated == 0 {
					break
				}
				if availableResource == 0 || buyers[i].Time > sellers[k][j].Times[k] {
					continue
				}
				if unallocated > availableResource {
					tmpSupply[j][k] += availableResource
				}
				if unallocated <= availableResource {
					tmpSupply[j][k] += unallocated
				}
				if j == sellersNum-1 {
					reservPrices[i] = 10000
				} else {
					reservPrices[i] += sellers[k][j+1].Prices[k] * float64(tmpSupply[j][k]*buyers[i].Time)
				}
				unallocated -= tmpSupply[j][k]
			}
			if unallocated > 0 || buyers[i].Price < reservPrices[i] {
				granted = false
				reservPrices[i] = 0
				break
			}
		}
		if granted {
			winners[i] = true
			for j := 0; j < sellersNum; j++ {
				for k := 0; k < 3; k++ {
					if tmpSupply[j][k] == 0 {
						continue
					}
					str := sellers[k][j].Address
					index := FindIndex(str, 0, sellers)
					supplyMatrix[i][index][k] += tmpSupply[j][k]
					costMatrix[j][k] += tmpSupply[j][k]
				}
			}
		}
	}
	return winners, reservPrices, supplyMatrix
}

func DeterminePayment(buyers []BuyerBid, sellers [3][]SellerBid, winners [100]bool, reservPrices [100]float64, supplyMatrix [100][100][3]int, ratings []float64) ([]float64, []float64, [100][100][3]int) {
	buyersNum := len(buyers)
	sellersNum := len(sellers[0])
	buyersPay := [100]float64{}
	sellersPay := [100]float64{}
	newBuyers := make([]BuyerBid, len(buyers))
	copy(newBuyers, buyers)
	for i := 0; i < buyersNum; i++ {
		if winners[i] {
			for j := i; j < buyersNum-1; j++ {
				newBuyers[j] = newBuyers[j+1]
			}
			newBuyers[buyersNum-1] = buyers[i]
			newBuyers[buyersNum-1].Price = 0
			newWinners, _, _ := Allocate(newBuyers, sellers)
			maxPrice := 10000.0
			for j := i; j < buyersNum-1; j++ {
				if newWinners[j] && !winners[j+1] {
					currentPrice := buyers[j+1].Price / math.Sqrt(float64(buyers[j+1].Quantities[0]+2*buyers[j+1].Quantities[1]+4*buyers[j+1].Quantities[2])) * math.Sqrt(float64(buyers[i].Quantities[0]+2*buyers[i].Quantities[1]+4*buyers[i].Quantities[2]))
					if currentPrice >= reservPrices[i] {
						maxPrice = currentPrice
						break
					} else {
						break
					}
				}
			}
			if maxPrice == 10000 {
				buyersPay[i] = reservPrices[i]
			} else {
				buyersPay[i] = maxPrice
			}
		}
	}

	sellersCost := [100]float64{}
	for i := 0; i < buyersNum; i++ {
		for j := 0; j < sellersNum; j++ {
			for k := 0; k < 3; k++ {
				sellersCost[j] += sellers[0][j].Prices[k] * float64(supplyMatrix[i][j][k]*buyers[i].Time)
			}
		}
	}
	reservSellersPrices := [100]float64{}
	for i := 0; i < buyersNum; i++ {
		for j := 0; j < sellersNum; j++ {
			for k := 0; k < 3; k++ {
				if supplyMatrix[i][j][k] > 0 {
					index := FindIndex(sellers[0][j].Address, k, sellers)
					reservSellersPrices[j] += sellers[k][index+1].Prices[k] * float64(supplyMatrix[i][j][k]*buyers[i].Time)
				}
			}
		}
	}
	totalSurplus := Sum(buyersPay[:]) - Sum(reservSellersPrices[:])
	totalShare := 0.0
	for i := 0; i < buyersNum; i++ {
		for j := 0; j < sellersNum; j++ {
			totalShare += float64((supplyMatrix[i][j][0]+2*supplyMatrix[i][j][1]+4*supplyMatrix[i][j][2])*buyers[i].Time) * ratings[j]
		}
	}
	for j := 0; j < sellersNum; j++ {
		individualShare := 0.0
		for i := 0; i < buyersNum; i++ {
			individualShare += float64((supplyMatrix[i][j][0]+2*supplyMatrix[i][j][1]+4*supplyMatrix[i][j][2])*buyers[i].Time) * ratings[j]
		}
		sellersPay[j] = reservSellersPrices[j] + individualShare/totalShare*totalSurplus
	}
	return buyersPay[:], sellersPay[:], supplyMatrix
}

func FindIndex(addr string, k int, sellers [3][]SellerBid) int {
	for i := 0; i < len(sellers[0]); i++ {
		if addr == sellers[k][i].Address {
			return i
		}
	}
	return -1
}

func FindBuyer(addr string, buyers []BuyerBid) int {
	for i := 0; i < len(buyers); i++ {
		if addr == buyers[i].Address {
			return i
		}
	}
	return -1
}

func GetFeedbackIndex(addr string, feedback *Feedback) int {
	index := -1
	for i := 0; i < len(feedback.Address); i++ {
		if feedback.Address[i] == addr {
			index = i
			break
		}
	}
	return index
}

func ChangeBalance(addr string, fee float64, accounts *Accounts) {
	index := -1
	for i := 0; i < len(accounts.Address); i++ {
		if addr == accounts.Address[i] {
			index = i
			break
		}
	}
	if index != -1 {
		accounts.Balance[index] += fee
	}
}

func StrToIntArr(s string) []int {
	res := []int{}
	ss := strings.Split(s, ",")
	for _, v := range ss {
		tmp, _ := strconv.Atoi(v)
		res = append(res, tmp)
	}
	return res
}

func StrToFloatArr(s string) []float64 {
	res := []float64{}
	ss := strings.Split(s, ",")
	for _, v := range ss {
		tmp, _ := strconv.ParseFloat(v, 64)
		res = append(res, tmp)
	}
	return res
}

func Sum(arr []float64) float64 {
	res := 0.0
	for i := 0; i < len(arr); i++ {
		res += arr[i]
	}
	return res
}

func Hash(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	sha1_hash := hex.EncodeToString(h.Sum(nil))
	return sha1_hash
}
