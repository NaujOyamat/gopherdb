package main

import (
	"context"
	"log"
	"time"

	"github.com/NaujOyamat/gopherdb"
)

// runCreateIndexes creates indexes on the collection.
func runCreateIndexes(ctx context.Context, coll *gopherdb.Collection) {
	log.Println()
	log.Println("Running create indexes...")

	benchmarkResult := CreateIndexResultBenchmark{
		Title: "Create indexes",
	}

	start := time.Now()
	err := coll.IndexManager.CreateMany(
		ctx,
		[]gopherdb.IndexModel{
			{
				Fields: []gopherdb.IndexField{
					{
						Name:  "name",
						Order: 1,
					},
					{
						Name:  "last_name",
						Order: 1,
					},
					{
						Name:  "age",
						Order: 1,
					},
				},
			},
		},
	)

	if err != nil {
		log.Fatal(err)
	}

	benchmarkResult.CreateIndexTime = time.Since(start)
	benchmarkResult.IndexCount = int64(len(coll.IndexManager.List()))

	printCreateIndexResult(benchmarkResult)
}

// printCreateIndexResult prints the result of the create index benchmark.
func printCreateIndexResult(result CreateIndexResultBenchmark) {
	log.Println()
	log.Println()
	log.Println("--------------------------------")
	log.Println(result.Title)
	log.Println("Create index time:", result.CreateIndexTime)
	log.Println("Index count:", result.IndexCount)
	log.Println("--------------------------------")
}
