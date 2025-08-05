package main

import (
	"log"

	"github.com/NaujOyamat/gopherdb"
)

func runUpdate(coll *gopherdb.Collection) {
	log.Println()
	log.Println("Running updates...")

	person := Person{
		Name:     "Genny",
		LastName: "Gutierrez",
		Age:      25,
	}

	resultInsert := coll.InsertOne(person)

	if resultInsert.Err != nil {
		log.Fatal(resultInsert.Err)
	}

	log.Println("Inserted person with id:", resultInsert.InsertedID)

	filter := map[string]any{"_id": resultInsert.InsertedID}
	person.ID = resultInsert.InsertedID.(string)
	person.Age = 36
	person.LastName = "Gutierrez Vieda"

	resultUpdate := coll.UpdateOne(filter, person)

	if resultUpdate.Err != nil {
		log.Fatal(resultUpdate.Err)
	}

	resultFind := coll.FindByID(resultInsert.InsertedID)

	if resultFind.Err != nil {
		log.Fatal(resultFind.Err)
	}

	log.Println("Updated person:", resultFind.Document())
}
