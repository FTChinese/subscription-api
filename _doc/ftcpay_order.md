# Order

## Definition

```typescript
interface Order {
    id: string;
    ftcId?: string;
    unionId?: string;
    priceId: string;
    discountId?: string;
    price: price;
    tier: 'standard' | 'premium';
    cycle: 'year' | 'month';
    amount: number;
    currency: string;
    kind: 'create' | 'renew' | 'upgrade' | 'add_on';
    payMethod: 'alipay' | 'wechat' | 'apple' | 'stripe';
    createdAt: string;
    confirmedAt?: string;
    startDate?: string;
    endDate?; string;
}
```

## Get a list of orders

```
GET /orders?page=<int>&per_page=<int>
```

```typescript
interface OrderList {
    total: number;
    page: number;
    limit: number;
    data: Order[];
}
```

## Get an order

```
GET /orders/{id}
```

A single `Order`.

## Get payment result of an order

```
GET /orders/{id}/payment-result
```

Query API of Alipay or Wechat for an order's payment result.

### Verify an order's payment

```
POST /orders/{id}/verify-payment
```

Check an order's payment status against Alipay/Wechat API, and update membership if the order is successfully paid but membership is not updated.

