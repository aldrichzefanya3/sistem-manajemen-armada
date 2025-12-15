package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/aldrichzefanya3/sistem-manajemen-armada/internal/common/platform/postgres"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := godotenv.Load()
	if err != nil {
		log.Info().Msg("No .env file found, using environment variables")
	} else {
		log.Info().Msg(".env file loaded")
	}

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
	}))

	dbPool := postgres.New(os.Getenv("DATABASE_URL"))
	defer dbPool.Close()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.GET("/vehicles/:vehicle_id/location", func(c *gin.Context) {
		vehicleID := c.Param("vehicle_id")

		rows, err := dbPool.Query(
			ctx,
			`SELECT vehicle_id, latitude, longitude, timestamp
					FROM vehicle_locations
					WHERE vehicle_id = $1
					ORDER BY timestamp DESC LIMIT 1`,
			vehicleID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "error querying database"})
			return
		}
		defer rows.Close()

		var result gin.H

		for rows.Next() {
			var vID string
			var lat, lon float64
			var ts int

			if err := rows.Scan(&vID, &lat, &lon, &ts); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "error when scanning row"})
				return
			}

			result = gin.H{
				"vehicle_id": vID,
				"latitude":   lat,
				"longitude":  lon,
				"timestamp":  ts,
			}
		}

		if len(result) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"message": "vehicle's location not found"})
			return
		}

		c.JSON(http.StatusOK, result)
	})

	router.GET("/vehicles/:vehicle_id/history", func(c *gin.Context) {
		vehicleID := c.Param("vehicle_id")

		startStr := c.Query("start")
		endStr := c.Query("end")

		if startStr == "" || endStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "start and end query parameters are required",
			})
			return
		}

		startDate, err := strconv.ParseInt(startStr, 10, 64)
		if err != nil {
			c.JSON(400, gin.H{"error": "invalid start timestamp"})
			return
		}

		endDate, err := strconv.ParseInt(endStr, 10, 64)
		if err != nil {
			c.JSON(400, gin.H{"error": "invalid end timestamp"})
			return
		}

		rows, err := dbPool.Query(
			ctx,
			`SELECT vehicle_id, latitude, longitude, timestamp
					FROM vehicle_locations
					WHERE vehicle_id = $1
					AND timestamp >= to_timestamp($2)
					AND timestamp <= to_timestamp($3)
					ORDER BY timestamp DESC`,
			vehicleID,
			startDate,
			endDate)
		if err != nil {
			log.Error().Err(err).Msg("Error querying database")
			c.JSON(http.StatusInternalServerError, gin.H{"message": "error querying database"})
			return
		}
		defer rows.Close()

		var data []gin.H

		for rows.Next() {
			var vID string
			var lat, lon float64
			var ts int

			if err := rows.Scan(&vID, &lat, &lon, &ts); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"message": "error scanning row"})
				return
			}

			data = append(data, gin.H{
				"vehicle_id": vID,
				"latitude":   lat,
				"longitude":  lon,
				"timestamp":  ts,
			})
		}

		if len(data) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"message": "no history found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": data,
		})
	})

	port := os.Getenv("PORT")

	server := &http.Server{
		Addr:    fmt.Sprint(":", port),
		Handler: router,
	}

	go func() {
		log.Info().Msgf("Server running on port :%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("HTTP server error")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutdown signal received")

	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal().Err(err)
	}

	log.Info().Msg("Server exited gracefully")
}
