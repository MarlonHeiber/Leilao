package auction_usecase_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"leilao/internal/entity/auction_entity"
	"leilao/internal/entity/bid_entity"
	"leilao/internal/internal_error"
	"leilao/internal/usecase/auction_usecase"

	"github.com/stretchr/testify/assert"
)

// MockAuctionRepository implementa AuctionRepositoryInterface
type MockAuctionRepository struct {
	store map[string]*auction_entity.Auction
	mu    sync.RWMutex
}

func NewMockAuctionRepository() *MockAuctionRepository {
	return &MockAuctionRepository{
		store: make(map[string]*auction_entity.Auction),
	}
}

func (m *MockAuctionRepository) CreateAuction(ctx context.Context, a *auction_entity.Auction) *internal_error.InternalError {
	m.mu.Lock()
	m.store[a.Id] = a
	m.mu.Unlock()

	// Simula a goroutine de fechamento automático com intervalo curto
	go func() {
		select {
		case <-time.After(2 * time.Second):
			m.mu.Lock()
			defer m.mu.Unlock()
			if auction, exists := m.store[a.Id]; exists {
				auction.Status = auction_entity.Completed
			}
		case <-ctx.Done():
			return
		}
	}()

	return nil
}

func (m *MockAuctionRepository) FindAuctions(ctx context.Context, status auction_entity.AuctionStatus, category, productName string) ([]auction_entity.Auction, *internal_error.InternalError) {
	return nil, nil
}

func (m *MockAuctionRepository) FindAuctionById(ctx context.Context, id string) (*auction_entity.Auction, *internal_error.InternalError) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if a, ok := m.store[id]; ok {
		return a, nil
	}
	return nil, internal_error.NewNotFoundError("auction not found")
}

// DummyBidRepository implementa BidEntityRepository
type DummyBidRepository struct{}

func (d *DummyBidRepository) CreateBid(ctx context.Context, bids []bid_entity.Bid) *internal_error.InternalError {
	return nil
}

func (d *DummyBidRepository) FindBidByAuctionId(ctx context.Context, auctionId string) ([]bid_entity.Bid, *internal_error.InternalError) {
	return nil, nil
}

func (d *DummyBidRepository) FindWinningBidByAuctionId(ctx context.Context, auctionId string) (*bid_entity.Bid, *internal_error.InternalError) {
	return nil, nil
}

func TestAuctionAutoClosure(t *testing.T) {
	ctx := context.Background()
	mockRepo := NewMockAuctionRepository()
	bidRepo := &DummyBidRepository{}
	usecase := auction_usecase.NewAuctionUseCase(mockRepo, bidRepo)

	input := auction_usecase.AuctionInputDTO{
		ProductName: "Produto Teste",
		Category:    "Categoria",
		Description: "Descrição válida para teste",
		Condition:   0,
	}

	err := usecase.CreateAuction(ctx, input)
	assert.Nil(t, err)

	time.Sleep(3 * time.Second)

	var auction *auction_entity.Auction
	for _, a := range mockRepo.store {
		auction = a
		break
	}
	assert.NotNil(t, auction)
	assert.Equal(t, auction_entity.Completed, auction.Status)
}
