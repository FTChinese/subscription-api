## Membership

Required headers: `X-User-Id` or `X-Union-Id`, or both for linked account.

* GET `/membership` Get a user's membership details
* PATCH `/membership` Update a user's membership. NOT implemented.
* PUT `/membership` Create a new membership. NOT implemented.
* GET `/membership/snapshots` Get a list of membership change history.
* POST `/membership/addons` Transfer addon to expiration time.
* PATCH `/membership/addons` Add addon to existing one.

## Get current membership

```
GET /membership
```

## List membership change history

```
GET /membership/snapshots
```

```typescript
interface SnapshotList {
    total: number;
    page: number;
    limit: number;
    data: Membersnapshot[];
}
```

## Claim AddOn

```
POST /membership/addons
```

Transfer current addon invoices to membership expiration date.

Returns the updated `Membership`

## Create an invoice for addon

```
PATCH /membership/addons
```

### Request Body

```typescript
interface AddOnParams {
    source: 'user_purchase' | 'carry_over' | 'compensation';
    tier: 'standard' | 'premium';
    cycle: 'month' | 'year',
    years: number;
    months: number;
    days: number;
    orderId?: string;
    paidAmount: number; // Default 0
    payMethod: 'alipay' | 'wechat';
    priceId?: string;
}
```

When `source` is `user_purchase`, `orderId` is required; otherwise it must not exist. `paidAmount`, `payMethod` and `priceId` should be copies from original order as is. The API does not check if the order actually exists.

When `source` is `compensation` or `carry_over`, there is actually no order for it and user did not pay. You should select any of `alipay` or `wechat` payment method.

### Response

```typescript
interface AddOnInvoiceCreated {
    invoice: Invoice;
    membership: Membership;
    snapshot: Snapshot;
}
```
