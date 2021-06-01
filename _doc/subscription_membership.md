# Membership

## The importance of `payment_method` column

You would never be able to mix all supported payment method together without this field.

ftc_vip如下字段组合互斥：

payment_method == (alipay || wechat) 和 ftc_plan_id 同时存在 并且 auto_renewal为0;

payment_method == apple  和 apple_subscription、auto_renewal 同时存在;

payment_method == stripe 和 stripe_subscription_id、stripe_plan_id、sub_status、auto_renewal 同时存在.

## Type Definition

### Snapshot

```typescript
interface MemberSnashot extends Membership {
    id: string;
    createdBy?: string;
    createdUtc: string;
    orderId?: string;
}
```
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
