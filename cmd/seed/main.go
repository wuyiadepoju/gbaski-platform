package main

import (
	"fmt"
	"gbaski-host/seed"
	"log"

	"github.com/gbaski/gbaski-shared/config"
)

func main() {
	// Load configuration
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	repo := seed.NewRepository()

	// "bb99244d-c9fb-40c3-a7b6-96a5aee0d35e"

	if err := repo.SeedEventCategories(); err != nil {
		log.Fatalf("Error seeding event categories: %v", err)
	}

	// userId, err := uuid.Parse("bb99244d-c9fb-40c3-a7b6-96a5aee0d35e")
	// if err != nil {
	// 	log.Fatalf("Error parsing UUID: %v", err)
	// }

	// repo.CreateEventsForUser(userId)

	// Seed users and their events
	// if err := repo.SeedUsers(); err != nil {
	// 	log.Fatalf("Error seeding users: %v", err)
	// }

	fmt.Println("Seeding completed successfully!")
}
