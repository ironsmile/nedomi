package lru

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/ironsmile/nedomi/mock"
	"github.com/ironsmile/nedomi/types"
)

func printLru(lru *TieredLRUCache) {
	fmt.Printf("TieredLru:\n cfg : %#v\n tierListSize: %d; requests: %d ; hits %d\n Tiers: \n",
		lru.cfg, lru.tierListSize, lru.requests, lru.hits)
	for index, tier := range lru.tiers {
		fmt.Printf("%d:", index)
		for el := tier.Front(); el != nil; el = el.Next() {
			oi := el.Value.(types.ObjectIndex)
			fmt.Printf("%s,", &oi)
		}

		fmt.Printf("\n")
	}

	for key, value := range lru.lookup {
		fmt.Printf("%s: %#v\n", hex.EncodeToString(key[:]), value)

	}
}

func printMockLogger(t *testing.T, logger *mock.Logger) {
	logs := logger.Logged()
	for _, log := range logs {
		fmt.Println(log)
	}
}
