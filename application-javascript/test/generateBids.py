import numpy as np
import math

class Buyer:

    def __init__(self, i, quantities, price, time):
        self.id = i
        self.quantities = quantities
        self.price = price
        self.time = time

    def self_print(self, f1, f2):
        print("node ../registerEnrollUser.js org1 buyer"+str(self.id), file=f1)
        print("node ../registerAccount.js org1 buyer"+str(self.id), file=f1)
        print("node ../bid.js org1 buyer"+str(self.id)+" $1", self.price, self.time, str(self.quantities[0])+","+str(self.quantities[1])+","+str(self.quantities[2]), file=f2)

class Seller:

    def __init__(self, i, quantities, prices, times):
        self.id = i
        self.quantities = quantities
        self.prices = prices
        self.times = times

    def self_print(self, f1, f2):
        print("node ../registerEnrollUser.js org2 seller"+str(self.id), file=f1)
        print("node ../registerAccount.js org2 seller"+str(self.id), file=f1)
        print("node ../bid.js org2 seller"+str(self.id)+" $1", str(self.prices[0])+","+str(self.prices[1])+","+str(self.prices[2]), str(self.times[0])+","+str(self.times[1])+","+str(self.times[2]), str(self.quantities[0])+","+str(self.quantities[1])+","+str(self.quantities[2]), file=f2)


def generate_clouds(number_of_buyers, number_of_sellers):
    buyers = []
    sellers = []
    types_of_resources = 3
    w = [1, 2, 4]
    lower_unit_cost = 0.1
    upper_unit_cost = 0.2

    # generate buyers and sellers
    quantity_in_buyer_needs = np.zeros([number_of_buyers, types_of_resources], dtype=int)
    unit_cost_in_buyer_needs = np.zeros([number_of_buyers, types_of_resources], dtype=float)
    price_in_buyer_needs = [0.0] * number_of_buyers
    quantity_in_seller_provision = np.zeros([number_of_sellers, types_of_resources], dtype=int)
    time_in_seller_provision = np.zeros([number_of_sellers, types_of_resources], dtype=int)
    price_in_seller_provision = np.zeros([number_of_sellers, types_of_resources], dtype=float)

    # generate quantities, U[1, 10]
    for i in range(0, types_of_resources):
        quantity_in_buyer_needs[:, i] = np.random.randint(0, 10, number_of_buyers)
        quantity_in_seller_provision[:, i] = np.random.randint(0, 10, number_of_sellers)

    # generate time, N(4, 1) and N(4, 1)
    time_in_buyer_needs = np.random.normal(4, 1, number_of_buyers)
    for i in range(0, number_of_buyers):
        if time_in_buyer_needs[i] < 0:
            time_in_buyer_needs[i] = 1
    for i in range(0, types_of_resources):
        time_in_seller_provision[:, i] = np.random.normal(4, 1, number_of_sellers)
        for j in range(0, number_of_sellers):
            if time_in_seller_provision[j][i] < 0:
                time_in_seller_provision[j][i] = 1

    # generate prices, U[1, 2]*time*quantities, U[1, 2]*w
    unit_cost_in_buyer_needs[:, 0] = np.random.uniform(lower_unit_cost, upper_unit_cost, number_of_buyers)
    unit_cost_in_buyer_needs[:, 1] = np.random.uniform(upper_unit_cost, upper_unit_cost * 2, number_of_buyers)
    unit_cost_in_buyer_needs[:, 2] = np.random.uniform(upper_unit_cost * 2, upper_unit_cost * 4, number_of_buyers)
    price_in_buyer_needs = np.random.uniform(lower_unit_cost, upper_unit_cost, number_of_buyers)
    
    for i in range(0, number_of_buyers):
        price_in_buyer_needs[i] = price_in_buyer_needs[i] * np.dot(quantity_in_buyer_needs[i], w) * time_in_buyer_needs[i]
    price_in_seller_provision[:, 0] = np.random.uniform(lower_unit_cost, upper_unit_cost, number_of_sellers)
    price_in_seller_provision[:, 1] = np.random.uniform(upper_unit_cost, upper_unit_cost * 2, number_of_sellers)
    price_in_seller_provision[:, 2] = np.random.uniform(upper_unit_cost * 2, upper_unit_cost * 4, number_of_sellers)

    np.around(price_in_buyer_needs, decimals=2, out=price_in_buyer_needs)
    np.around(price_in_seller_provision, decimals=2, out=price_in_seller_provision)
    for i in range(0, number_of_buyers):
        buyers.append(Buyer(i+1, quantity_in_buyer_needs[i], price_in_buyer_needs[i], int(time_in_buyer_needs[i])))
    for i in range(0, number_of_sellers):
        sellers.append(
            Seller(i+1, quantity_in_seller_provision[i], price_in_seller_provision[i], time_in_seller_provision[i].astype(int))
        )

    return buyers, sellers

number_of_buyers = 100
number_of_sellers = 100
buyers, sellers = generate_clouds(number_of_buyers, number_of_sellers)
f1 = open('accountReg.sh','w')
f2 = open('bidConfig.sh','w')

for buyer in buyers:
    buyer.self_print(f1, f2)

for seller in sellers:
    seller.self_print(f1, f2)

f1.close()
f2.close()
