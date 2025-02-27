package auction

import (
	"context"
	"fullcycle-auction_go/configuration/logger"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"fullcycle-auction_go/internal/internal_error"
	"log"
	"os"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var mutex sync.Mutex

type AuctionEntityMongo struct {
	Id          string                          `bson:"_id"`
	ProductName string                          `bson:"product_name"`
	Category    string                          `bson:"category"`
	Description string                          `bson:"description"`
	Condition   auction_entity.ProductCondition `bson:"condition"`
	Status      auction_entity.AuctionStatus    `bson:"status"`
	Timestamp   int64                           `bson:"timestamp"`
	EndsAt      int64                           `bson:"ends_at"`
}
type AuctionRepository struct {
	Collection *mongo.Collection
}

func NewAuctionRepository(database *mongo.Database) *AuctionRepository {
	return &AuctionRepository{
		Collection: database.Collection("auctions"),
	}
}

func (ar *AuctionRepository) CreateAuction(
	ctx context.Context,
	auctionEntity *auction_entity.Auction) *internal_error.InternalError {
	auctionEntityMongo := &AuctionEntityMongo{
		Id:          auctionEntity.Id,
		ProductName: auctionEntity.ProductName,
		Category:    auctionEntity.Category,
		Description: auctionEntity.Description,
		Condition:   auctionEntity.Condition,
		Status:      auctionEntity.Status,
		Timestamp:   auctionEntity.Timestamp.Unix(),
	}
	_, err := ar.Collection.InsertOne(ctx, auctionEntityMongo)
	if err != nil {
		logger.Error("Error trying to insert auction", err)
		return internal_error.NewInternalServerError("Error trying to insert auction")
	}
	ar.startAuctionExpiryCheck(ctx, auctionEntity.Id)

	return nil
}

func (ar *AuctionRepository) startAuctionExpiryCheck(ctx context.Context, auctionID string) {
	auctionDuration := os.Getenv("AUCTION_DURATION")
	duration, err := time.ParseDuration(auctionDuration)
	if err != nil {
		duration = time.Minute * 5
	}

	go func() {
		timer := time.NewTimer(duration)
		<-timer.C

		mutex.Lock()
		defer mutex.Unlock()

		_, err := ar.Collection.UpdateOne(ctx, bson.M{"_id": auctionID}, bson.M{"$set": bson.M{"status": auction_entity.Completed, "ends_at": time.Now().Unix()}})
		if err != nil {
			log.Printf("Erro ao fechar o leilão %s: %v", auctionID, err)
		}
	}()
}
