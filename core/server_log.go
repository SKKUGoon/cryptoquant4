package core

import (
	"log"
	"strconv"
	"time"

	database "cryptoquant.com/m/data/database"
	pb "cryptoquant.com/m/gen/traderpb"
	binancerest "cryptoquant.com/m/internal/binance/rest"
	upbitrest "cryptoquant.com/m/internal/upbit/rest"
	"github.com/google/uuid"
)

func (s *Server) StartKimchiTradeLog() {
	log.Println("Starting Kimchi trade log")
	go func() {
		for {
			select {
			case <-s.ctx.Done():
				return
			case row := <-s.kimchiTradeLog:
				s.Database.InsertStrategyKimchiOrderLog(row)
			}
		}
	}()
}

func (s *Server) StartWalletLog() {
	log.Println("Starting wallet log")
	go func() {
		for {
			select {
			case <-s.ctx.Done():
				return
			case row := <-s.walletLog:
				s.TimeScale.InsertAccountSnapshot([]database.AccountSnapshot{row})
			}
		}
	}()
}

func (s *Server) CreateKimchiOrderLog(
	isPairEnter bool,
	pairOrder *pb.PairOrderSheet,
	upbitResp *upbitrest.OrderResult,
	binanceResp *binancerest.OrderResult,
	orderTime, executionTime time.Time,
) ([]database.KimchiOrderLog, error) {
	var pairSide string
	var uuid = uuid.New().String()

	upbitOrder := pairOrder.UpbitOrder
	binanceOrder := pairOrder.BinanceOrder
	anchorPrice := pairOrder.ExchangeRate

	if isPairEnter {
		pairSide = "long"
	} else {
		pairSide = "short"
	}

	binancePrice, err := strconv.ParseFloat(binanceResp.Success.Price, 64)
	if err != nil {
		log.Printf("Failed to parse binance price: %v", err)
		return nil, err
	}

	upbitLog := database.KimchiOrderLog{
		PairID:        uuid,
		OrderTime:     orderTime,
		ExecutionTime: executionTime,
		Symbol:        upbitOrder.Symbol,
		PairSide:      pairSide,
		Exchange:      "upbit",
		Side:          upbitOrder.Side,
		OrderPrice:    upbitOrder.Price,
		AnchorPrice:   anchorPrice,
	}

	binanceLog := database.KimchiOrderLog{
		PairID:        uuid,
		OrderTime:     orderTime,
		ExecutionTime: executionTime,
		Symbol:        upbitOrder.Symbol, // Unify symbol
		PairSide:      pairSide,
		Exchange:      "binance",
		Side:          binanceOrder.Side,
		OrderPrice:    binanceOrder.Price,
		ExecutedPrice: binancePrice,
		AnchorPrice:   anchorPrice,
	}

	return []database.KimchiOrderLog{upbitLog, binanceLog}, nil
}
