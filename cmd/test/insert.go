package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/NaujOyamat/gopherdb"
)

// runInsert runs the insert benchmark.
func runInsert(ctx context.Context, coll *gopherdb.Collection) {
	log.Println()
	log.Println("Running inserts...")

	const seedSize = 1_000_000

	benchmarkResult := InsertResultBenchmark{
		Title: fmt.Sprintf("Insert %d people", seedSize),
	}

	people := generatePeople(seedSize)

	start := time.Now()
	result := coll.Insert(people)
	benchmarkResult.InsertTime = time.Since(start)
	benchmarkResult.InsertCount = int64(len(result.InsertedIDs))

	if result.Err != nil {
		log.Fatal(result.Err)
	}

	printInsertResult(benchmarkResult)
}

// printInsertResult prints the result of the insert benchmark.
func printInsertResult(result InsertResultBenchmark) {
	log.Println()
	log.Println()
	log.Println("--------------------------------")
	log.Println(result.Title)
	log.Println("Insert time:", result.InsertTime)
	log.Println("Insert count:", result.InsertCount)
	log.Println("--------------------------------")
}
