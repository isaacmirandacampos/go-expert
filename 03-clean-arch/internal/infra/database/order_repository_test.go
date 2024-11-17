package database

import (
	"database/sql"
	"testing"

	"github.com/devfullcycle/20-CleanArch/internal/entity"
	"github.com/stretchr/testify/suite"

	// sqlite3
	_ "github.com/mattn/go-sqlite3"
)

type OrderRepositoryTestSuite struct {
	suite.Suite
	Db *sql.DB
}

func (suite *OrderRepositoryTestSuite) SetupSuite() {
	db, err := sql.Open("sqlite3", ":memory:")
	suite.NoError(err)
	db.Exec("CREATE TABLE orders (id varchar(255) NOT NULL, price float NOT NULL, tax float NOT NULL, final_price float NOT NULL, PRIMARY KEY (id))")
	suite.Db = db
}

func (suite *OrderRepositoryTestSuite) TearDownTest() {
	suite.Db.Exec("DELETE FROM orders")
}

func (suite *OrderRepositoryTestSuite) TearDownSuite() {
	suite.Db.Close()
}

func (suite *OrderRepositoryTestSuite) orderFactory(id string, price float64, tax float64) *entity.Order {
	order := entity.Order{
		ID:    id,
		Price: price,
		Tax:   tax,
	}
	err := order.CalculateFinalPrice()
	_, err = suite.Db.Exec("INSERT INTO orders (id, price, tax, final_price) VALUES (?, ?, ?, ?)", id, price, tax, order.FinalPrice)
	suite.NoError(err)
	return &order
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(OrderRepositoryTestSuite))
}

func (suite *OrderRepositoryTestSuite) TestGivenAnOrder_WhenSave_ThenShouldSaveOrder() {
	order, err := entity.NewOrder("123", 10.0, 2.0)
	suite.NoError(err)
	suite.NoError(order.CalculateFinalPrice())
	repo := NewOrderRepository(suite.Db)
	err = repo.Save(order)
	suite.NoError(err)

	var orderResult entity.Order
	err = suite.Db.QueryRow("Select id, price, tax, final_price from orders where id = ?", order.ID).
		Scan(&orderResult.ID, &orderResult.Price, &orderResult.Tax, &orderResult.FinalPrice)

	suite.NoError(err)
	suite.Equal(order.ID, orderResult.ID)
	suite.Equal(order.Price, orderResult.Price)
	suite.Equal(order.Tax, orderResult.Tax)
	suite.Equal(order.FinalPrice, orderResult.FinalPrice)
}

func (suite *OrderRepositoryTestSuite) TestGivenAnOrder_WhenGetTotal_ThenShouldReturnTotalOrders() {
	suite.orderFactory("123", 10.0, 2.0)
	suite.orderFactory("456", 10.0, 2.0)
	suite.orderFactory("789", 10.0, 2.0)
	repo := NewOrderRepository(suite.Db)
	total, err := repo.GetTotal()
	suite.NoError(err)
	suite.Equal(3, total)
}

func (suite *OrderRepositoryTestSuite) TestGivenAnOrder_WhenListOrders_ThenShouldReturnOrders() {
	suite.orderFactory("123", 10.0, 2.0)
	suite.orderFactory("456", 10.0, 2.0)
	suite.orderFactory("789", 10.0, 2.0)
	repo := NewOrderRepository(suite.Db)
	orders, _ := repo.List()
	suite.Equal(3, len(orders))
	suite.Equal("123", orders[0].ID)
	suite.Equal("456", orders[1].ID)
	suite.Equal("789", orders[2].ID)
	suite.Equal(10.0, orders[0].Price)
	suite.Equal(10.0, orders[1].Price)
	suite.Equal(10.0, orders[2].Price)
	suite.Equal(2.0, orders[0].Tax)
	suite.Equal(2.0, orders[1].Tax)
	suite.Equal(2.0, orders[2].Tax)
	suite.Equal(12.0, orders[0].FinalPrice)
	suite.Equal(12.0, orders[1].FinalPrice)
	suite.Equal(12.0, orders[2].FinalPrice)
}
