package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

func New(databaseURL string) *pgxpool.Pool {
	dbpool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		log.Fatal().Msg("Failed to connect to DB")
	}

	log.Info().Msg("Sucessfully connect to postgres")

	return dbpool
}
