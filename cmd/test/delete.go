package main

import (
	"log"

	"github.com/NaujOyamat/gopherdb"
)

func runDelete(coll *gopherdb.Collection) {
	log.Println()
	log.Println("Running deletes...")

	person := Person{
		Name:     "Genny",
		LastName: "Gutierrez",
		Age:      25,
	}

	person2 := Person{
		Name:     "Juan",
		LastName: "Perez",
		Age:      30,
	}

	resultInsert := coll.InsertOne(person)
	if resultInsert.Err != nil {
		log.Fatal(resultInsert.Err)
	}

	resultInsert2 := coll.InsertOne(person2)
	if resultInsert2.Err != nil {
		log.Fatal(resultInsert2.Err)
	}

	filter := map[string]any{"_id": resultInsert.InsertedID}
	resultDelete := coll.DeleteOne(filter)

	if resultDelete.Err != nil {
		log.Fatal(resultDelete.Err)
	}

	log.Println("Deleted person with id:", resultDelete.DeletedID)

	results := coll.Find(nil)

	log.Println("Total count:", results.TotalCount)
}
