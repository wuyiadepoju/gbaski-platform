package seed

import (
	"fmt"
	"gbaski-host/internal/event"
	"strings"

	"github.com/gbaski/gbaski-ext/auth"
	postgres "github.com/gbaski/gbaski-shared/postgres"
	redis "github.com/gbaski/gbaski-shared/redis"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db        *sqlx.DB
	redis     *redis.Redis
	authRepo  *auth.Repository
	eventRepo *event.Repository
}

func NewRepository() *Repository {
	return &Repository{
		db:        sqlx.NewDb(postgres.New(), "postgres"),
		redis:     redis.New(),
		authRepo:  auth.NewRepository(),
		eventRepo: event.NewRepository(),
	}
}

func (r *Repository) SeedEventCategories() error {
	categories := []event.EventCategory{
		{
			Name:        "Concerts and Shows",
			Description: "Music concerts, comedy shows, talent showcases, and live performances that bring people together for unforgettable entertainment.",
		},
		{
			Name:        "Competitions and Contests",
			Description: "Beauty pageants, talent competitions, voting contests, and gaming tournaments where participants compete and audiences engage.",
		},
		{
			Name:        "Promotional Events",
			Description: "Brand activations, marketing campaigns, and influencer-led events to showcase products and services.",
		},
		{
			Name:        "Corporate Events",
			Description: "Conferences, seminars, workshops, and product launches designed for businesses, professionals, and networking opportunities.",
		},
		{
			Name:        "Sports and Fitness",
			Description: "Football matches, marathons, fitness challenges, and e-sports tournaments that keep participants active and excited.",
		},
		{
			Name:        "Parties and Social Gatherings",
			Description: "Birthday parties, wedding receptions, reunions, and themed parties to celebrate life's special moments.",
		},
		{
			Name:        "Educational Events",
			Description: "Academic fairs, quiz competitions, career expos, and lectures that inspire learning and growth.",
		},
		{
			Name:        "Community Events",
			Description: "Fundraisers, cultural festivals, church programs, and charity events that unite people for a cause.",
		},
		{
			Name:        "Interactive Gaming Events",
			Description: "Trivia games, Tussle games, Steal or Split challenges, and task-based engagements to entertain and involve audiences.",
		},
		{
			Name:        "Other",
			Description: "Custom category not listed above",
		},
	}

	for _, category := range categories {
		category.Public = true
		categoryId, err := r.eventRepo.CreateEventCategory(category)
		if err != nil {
			return fmt.Errorf("error creating category %s: %v", category.Name, err)
		}
		fmt.Printf("Created category: %s (ID: %d)\n", category.Name, categoryId)
	}

	fmt.Println("Successfully seeded event categories")
	return nil
}

func toSlug(name string) string {
	return strings.ReplaceAll(strings.ToLower(name), " ", "-")
}
