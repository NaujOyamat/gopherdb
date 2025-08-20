package main

import "time"

type FindResultBenchmark struct {
	Title         string
	QueryTime     time.Duration
	UnmarshalTime time.Duration
	IndexUsed     string
	TotalCount    int64
}

type InsertResultBenchmark struct {
	Title       string
	InsertTime  time.Duration
	InsertCount int64
}

type CreateIndexResultBenchmark struct {
	Title           string
	CreateIndexTime time.Duration
	IndexCount      int64
}

type DeleteResultBenchmark struct {
	Title       string
	DeleteTime  time.Duration
	DeleteCount int64
}

type UpdateResultBenchmark struct {
	Title       string
	UpdateTime  time.Duration
	UpdateCount int64
}

type Person struct {
	ID       string `bson:"_id"`
	Name     string `bson:"name"`
	LastName string `bson:"last_name"`
	Age      int    `bson:"age"`
}
