package client

type OrderClient struct {
	OrderID string `db:"order_id"`
	Client
}
