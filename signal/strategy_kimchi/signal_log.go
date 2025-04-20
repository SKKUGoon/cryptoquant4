package signalkimchi

import (
	"log"

	database "cryptoquant.com/m/data/database"
)

// StartTSLog starts the TimeScale log.
// It creates a goroutine that logs the premium data to the TimeScale database.
func (e *SignalContext) StartTSLog() {
	log.Println("Starting DB log")
	buffer := make([]database.PremiumLog, 0, 100) // Insert 100 at a time
	go func() {
		for {
			select {
			case <-e.ctx.Done():
				return
			case row := <-e.premiumLog:
				buffer = append(buffer, row)
				if len(buffer) >= 100 {
					bufferCopy := make([]database.PremiumLog, len(buffer))
					copy(bufferCopy, buffer) // Copy the buffer to avoid race condition
					go func(logs []database.PremiumLog) {
						if err := e.TimeScale.InsertPremiumLog(logs); err != nil {
							log.Printf("Failed to insert premium log: %v", err)
						} else {
							log.Printf("Inserted %v rows to TimeScale", len(logs))
						}
					}(bufferCopy)
					buffer = make([]database.PremiumLog, 0, 100)
				}
			}
		}
	}()
}
