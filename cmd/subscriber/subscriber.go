package main

import (
	"context"
	"encoding/json"

	"github.com/aldrichzefanya3/sistem-manajemen-armada/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

func ProcessMessage(ctx context.Context, db *pgxpool.Pool, payload []byte) {
	var data model.MockData

	if err := json.Unmarshal(payload, &data); err != nil {
		log.Error().Err(err).Msg("Invalid JSON payload")
		return
	}

	if err := saveLocation(ctx, db, data); err != nil {
		log.Warn().
			Err(err).
			Str("vehicle_id", data.VehicleID).
			Msg("Message can't be processed")
		return
	}

	log.Info().
		Str("vehicle_id", data.VehicleID).
		Msg("Message processed successfully")
}

func saveLocation(ctx context.Context, db *pgxpool.Pool, data model.MockData) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	_, err = tx.Exec(
		ctx,
		`INSERT INTO vehicle_locations (vehicle_id, latitude, longitude, timestamp)
		 VALUES ($1, $2, $3, $4)`,
		data.VehicleID,
		data.Latitude,
		data.Longitude,
		data.Timestamp,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to insert vehicle location")
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	log.Info().Msg("Sucessfully inserted vehicle location")

	return nil
}
