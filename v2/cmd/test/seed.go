package main

import "github.com/brianvoe/gofakeit/v6"

// generatePeople generates a slice of Person with the given number of people.
func generatePeople(n int) []Person {
	people := make([]Person, n)

	for i := range n {
		people[i] = Person{
			Name:     gofakeit.FirstName(),
			LastName: gofakeit.LastName(),
			Age:      gofakeit.Number(18, 100),
		}
	}

	return people
}
