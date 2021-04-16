package footprint

// OrderClient is used to log client metadata when creating an order.
type OrderClient struct {
	OrderID string `db:"order_id"`
	Client
}
