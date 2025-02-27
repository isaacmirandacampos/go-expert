package auction

import (
	"context"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

const timeSleep = 3 * time.Second

func TestCreateAuction(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	auctionEntity := &auction_entity.Auction{
		Id:          "1",
		ProductName: "Leilão Teste",
		Category:    "Categoria Teste",
		Description: "Descrição Teste",
		Condition:   auction_entity.New,
		Status:      auction_entity.Active,
		Timestamp:   time.Now(),
	}

	mt.Run("create auction successfully", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateSuccessResponse())
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "auctions.create", mtest.FirstBatch, bson.D{
			{"_id", auctionEntity.Id},
			{"product_name", auctionEntity.ProductName},
			{"status", auctionEntity.Status},
		}))

		// Configuração inicial
		db := mt.DB
		repo := NewAuctionRepository(db)

		// Define o tempo de expiração como 2 segundos
		err := os.Setenv("AUCTION_DURATION", "2s")
		if err != nil {
			t.Fatalf("Erro ao configurar variável de ambiente: %v", err)
		}

		// Cria o leilão
		errInternal := repo.CreateAuction(context.Background(), auctionEntity)
		if errInternal != nil {
			t.Fatalf("Erro ao criar leilão: %v", errInternal)
		}

		// Verifica se o leilão foi inserido no banco de dados
		var result AuctionEntityMongo
		err = repo.Collection.FindOne(context.Background(), bson.M{"_id": auctionEntity.Id}).Decode(&result)
		if err != nil {
			t.Fatalf("Erro ao buscar leilão: %v", err)
		}

		assert.Equal(t, auctionEntity.ProductName, result.ProductName)

		auctionDuration := os.Getenv("AUCTION_DURATION")
		duration, err := time.ParseDuration(auctionDuration)
		if err != nil {
			duration = time.Minute * 5
		}
		if duration < timeSleep {
			mt.AddMockResponses(bson.D{
				{"ok", 1},
				{"nModified", 1},
			})

			time.Sleep(timeSleep)

			mt.AddMockResponses(mtest.CreateCursorResponse(2, "auctions.update", mtest.FirstBatch, bson.D{
				{"_id", auctionEntity.Id},
				{"product_name", auctionEntity.ProductName},
				{"status", auction_entity.Closed},
			}))
		} else {
			mt.AddMockResponses(mtest.CreateCursorResponse(2, "auctions.update", mtest.FirstBatch, bson.D{
				{"_id", auctionEntity.Id},
				{"product_name", auctionEntity.ProductName},
				{"status", auctionEntity.Status},
			}))

			time.Sleep(timeSleep)

			mt.AddMockResponses(bson.D{
				{"ok", 1},
				{"nModified", 0},
			})
		}

		// Verifica se o leilão foi fechado automaticamente
		err = repo.Collection.FindOne(context.Background(), bson.M{"_id": auctionEntity.Id}).Decode(&result)
		if err != nil {
			t.Fatalf("Erro ao buscar leilão: %v", err)
		}

		assert.Equal(t, auction_entity.Closed, result.Status)

	})
}
