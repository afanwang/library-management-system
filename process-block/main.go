package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Event Definition:
// - An event is emitted by a smart contract during a transaction execution.
// - Each event has a signature, which is a unique identifier for the type of event.
// - Events are stored in the transaction receipt's logs.
// - The event signature is the first topic (index 0) in the log entry.
func getBlockByNumber(client *ethclient.Client, blockNumber *big.Int) (*types.Block, error) {
	return client.BlockByNumber(context.Background(), blockNumber)
}

func getTransactionReceipt(client *ethclient.Client, txHash common.Hash) (*types.Receipt, error) {
	return client.TransactionReceipt(context.Background(), txHash)
}

// processBlock Retrieves a block using getBlockByNumber.
// Processes all transactions in the block.
// Fetches the receipt for each transaction and counts event logs using their signature (log.Topics[0]).
func processBlock(client *ethclient.Client, blockNumber *big.Int) (map[string]int, error) {
	// TODO: Implement this function to process a single block
	// It should get the block, process all transactions, and return event counts (event sig -> count)
	log.Printf("Processing block %s", blockNumber.String())
	block, err := getBlockByNumber(client, blockNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get block: %w", err)
	}

	eventCounts := make(map[string]int)

	for _, tx := range block.Transactions() {
		receipt, err := getTransactionReceipt(client, tx.Hash())
		if err != nil {
			log.Printf("failed to get receipt for tx %s: %v", tx.Hash().Hex(), err)
			continue
		}

		// Iterate over the logs to count event signatures
		for _, log := range receipt.Logs {
			sig := log.Topics[0].Hex() // First topic contains the event signature
			eventCounts[sig]++
		}
	}

	return eventCounts, nil
}

// Version 1: sequential block processing
// processBlockRange iterates over a block range from start to end, Calls processBlock for each block. Aggregates all event counts across the range.
// func processBlockRange(client *ethclient.Client, start, end *big.Int) (map[string]int, error) {
// 	// TODO: Implement this function to process a range of blocks
// 	// It should use processBlock for each block in the range and aggregate the results
// 	log.Printf("Processing block range from %s to %s", start.String(), end.String())

// 	totalEventCounts := make(map[string]int)

// 	for blockNum := new(big.Int).Set(start); blockNum.Cmp(end) <= 0; blockNum.Add(blockNum, big.NewInt(1)) {
// 		blockEventCounts, err := processBlock(client, blockNum)
// 		if err != nil {
// 			log.Printf("Error processing block %s: %v", blockNum.String(), err)
// 			continue
// 		}

// 		// Aggregate the counts from each block
// 		for sig, count := range blockEventCounts {
// 			totalEventCounts[sig] += count
// 		}
// 	}

// 	return totalEventCounts, nil
// }

// Worker function that processes blocks and sends results over a channel.
func worker(client *ethclient.Client, blockCh <-chan *big.Int, resultCh chan<- map[string]int, wg *sync.WaitGroup) {
	defer wg.Done()
	for blockNum := range blockCh {
		eventCounts, err := processBlock(client, blockNum)
		if err != nil {
			log.Printf("Error processing block %s: %v", blockNum.String(), err)
			continue
		}
		resultCh <- eventCounts
	}
}

// Version 2:
// Parallel processing of a block range using a worker pool.
func processBlockRange(client *ethclient.Client, start, end *big.Int) (map[string]int, error) {
	numWorkers := 10                      // Number of worker goroutines
	blockCh := make(chan *big.Int)        // Channel to send block numbers
	resultCh := make(chan map[string]int) // Channel to collect results

	var wg sync.WaitGroup

	// Start worker goroutines.
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(client, blockCh, resultCh, &wg)
	}

	// Send block numbers to the workers.
	go func() {
		for blockNum := new(big.Int).Set(start); blockNum.Cmp(end) <= 0; blockNum.Add(blockNum, big.NewInt(1)) {
			blockCh <- new(big.Int).Set(blockNum) // Send a copy of blockNum
		}
		close(blockCh) // Close channel after all blocks are sent
	}()

	// Collect results in a separate goroutine.
	go func() {
		wg.Wait()       // Wait for all workers to finish
		close(resultCh) // Close result channel after processing
	}()

	// Aggregate the results from all blocks.
	totalEventCounts := make(map[string]int)
	for eventCounts := range resultCh {
		for sig, count := range eventCounts {
			totalEventCounts[sig] += count
		}
	}

	return totalEventCounts, nil
}

func main() {
	client, err := ethclient.Dial("https://testnet.[some]rpc.io")
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	log.Printf("Connected to the Ethereum client")

	// Process the latest 10 blocks
	latestBlock := uint64(170024)

	start := new(big.Int).SetUint64(latestBlock - 9)
	end := new(big.Int).SetUint64(latestBlock)

	totalEventCount, err := processBlockRange(client, start, end)
	if err != nil {
		log.Fatalf("Failed to process block range: %v", err)
	}

	fmt.Println("Total event counts across the block range:")
	for signature, count := range totalEventCount {
		fmt.Printf("%s: %d\n", signature, count)
	}
}
