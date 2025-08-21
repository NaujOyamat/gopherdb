package main

import (
	"log"
	"time"

	"github.com/NaujOyamat/gopherdb/v2"
)

// runFindAll runs the find all benchmark.
func runFindAll(coll *gopherdb.Collection) {
	log.Println()
	log.Println("Running find all...")

	benchmarkResult := FindResultBenchmark{
		Title: "Find all",
	}

	filter := map[string]any{}

	start := time.Now()
	results := coll.Find(filter)

	if results.Err != nil {
		log.Fatal(results.Err)
	}

	benchmarkResult.QueryTime = time.Since(start)

	if results.IndexUsed != nil {
		benchmarkResult.IndexUsed = results.IndexUsed.Options.Name
	}

	people := make([]Person, 0)

	start = time.Now()
	err := results.Unmarshal(&people)
	benchmarkResult.TotalCount = results.TotalCount

	if err != nil {
		log.Fatal(err)
	}

	benchmarkResult.UnmarshalTime = time.Since(start)

	printFindAllResult(benchmarkResult)
}

// runFindWithFilterWithSimpleIndex runs the find with filter with simple index benchmark.
func runFindWithFilterWithSimpleIndex(coll *gopherdb.Collection) {
	log.Println()
	log.Println("Running find with filter with simple index...")

	benchmarkResult := FindResultBenchmark{
		Title: "Find with filter with simple index",
	}

	filter := map[string]any{
		"age": 30,
	}

	start := time.Now()
	results := coll.Find(filter)

	if results.Err != nil {
		log.Fatal(results.Err)
	}

	benchmarkResult.QueryTime = time.Since(start)

	if results.IndexUsed != nil {
		benchmarkResult.IndexUsed = results.IndexUsed.Options.Name
	}

	people := make([]Person, 0)

	start = time.Now()
	err := results.Unmarshal(&people)
	benchmarkResult.TotalCount = results.TotalCount

	if err != nil {
		log.Fatal(err)
	}

	benchmarkResult.UnmarshalTime = time.Since(start)

	printFindAllResult(benchmarkResult)
}

// runFindWithFilterWithCompoundIndex runs the find with filter with compound index benchmark.
func runFindWithFilterWithCompoundIndex(coll *gopherdb.Collection) {
	log.Println()
	log.Println("Running find with filter with compound index...")

	benchmarkResult := FindResultBenchmark{
		Title: "Find with filter with compound index",
	}

	filter := map[string]any{
		"name":      "Patricia",
		"last_name": "Beahan",
		"age":       map[string]any{"$gte": 40},
	}

	start := time.Now()
	results := coll.Find(filter)

	if results.Err != nil {
		log.Fatal(results.Err)
	}

	benchmarkResult.QueryTime = time.Since(start)

	if results.IndexUsed != nil {
		benchmarkResult.IndexUsed = results.IndexUsed.Options.Name
	}

	people := make([]Person, 0)

	start = time.Now()
	err := results.Unmarshal(&people)
	benchmarkResult.TotalCount = results.TotalCount

	if err != nil {
		log.Fatal(err)
	}

	benchmarkResult.UnmarshalTime = time.Since(start)

	printFindAllResult(benchmarkResult)
}

// printFindAllResult prints the result of the find all benchmark.
func printFindAllResult(result FindResultBenchmark) {
	log.Println()
	log.Println()
	log.Println("--------------------------------")
	log.Println(result.Title)
	log.Println("Query time:", result.QueryTime)
	log.Println("Unmarshal time:", result.UnmarshalTime)
	log.Println("Index used:", result.IndexUsed)
	log.Println("Total count:", result.TotalCount)
	log.Println("--------------------------------")
}
