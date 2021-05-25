# Invoice

Invoice is the result of a successfully paid order, or a carry-over action.

Currently, invoices might be generated when user explicitly places order, or caused by a switching action:

* an order user paid, for either create, renew or upgrade;
* remaining valid time of current membership when:
    * one-time-purchase standard edition switching to premium edition;
    * one-time-purchase switching to subscription model via IAP or Stripe.

For upgrading from one-time-purchase standard edition to premium edition, two invoices will be generated.

## Concepts

Before proceeding to the details of Invoice type, we should understand related concepts.

### One-time-purchase vs subscription

The two are different business model with distinct requirements. One-time-purchase is usually pay-as-you-go. You give me the money, and I give you my product. Then that's done. You can buy any many copies of the same product as you like as long as you pay me. I don't care how many items you buy and how you use it.

Subscription has a stronger link between user and seller. The seller need to track how many items a user purchased, how this user is using it, whether the items are consumed and user should replenish stocks.

As we are forced to use one-time-purchase to handle subscription business model, we have to bridge the gap between them; hence the add-on.

### Add-On

Add-on is introduced to solve the problem of:

* A valid standard subscription changes to premium edition, both using one-time-purchase;
* A valid one-time-purchase wants to use IAP or stripe.

Previously when case 1 occurred, the approach we adopted is to find out all valid orders with a future expiration date, calculate the remaining portion of all those orders in terms of money, and subtract it from the price of our premium product, and ask user to pay a tailored price. The challenged posed here are that:

* the remaining portions are hard to calculate;
* each user is using a tailored version of pricing which is hard to scale.
  
From the seller's point of view, we are sometime puzzled as to where does a user's paid price comes from.

When case 2 occurred, user might find out he paid twice for a single product.

A better solution might be reserve some portion of user purchase for future usage:

* When one-time-purchase user want to switch from standard edition to premium, we can reserve the standard's remaining period and reuse it after premium expired;
* When one-time-purchase want to switch to subscription, we should encourage such action since subscription could bring more stable and predictable income. The one-time-purchase's remaining portion, whether standard or premium, should be stored somewhere waiting for future usage. 
  
The client should, upon a one-time-purchase premium switching to subscription model of standard edition, warn user of such a "downgrading" action and leave it for user to decide.

The add-on acts like a gateway to such a reserving place. The actual storage is carried out by Invoice.

Add-on object look like this:

```json
{
  "standardAddOn": 30,
  "premiumAddOn": 67
}
``` 

These two fields exists under the Membership type, rather than being a standalone data object. Its existence is only a hint to the client that a membership's valid period is not ended yet, even if the expiration time is passed, and the client should not deny user access to content.

When current membership expired and add-on exists, client should ask server to transfer all invoices that is not consumed yet to membership, with premium edition taking higher precedence over standard edition. See section "How add-on is claimed?"

Please note that you should never replace the values in add-on field. Those fields should be zero by default, and you always perform *addition* on existing values (in your application). The only exception is the moment the add-on is claimed, and the corresponding field should be reset to 0.

### Carry-Over

Carry-over is an action performed on current membership with valid remaining subscription period.

When user decides to switch subscription plan, or subscription model, the remaining portion of current membership is carried over to the invoice store, and the remaining days is added to current one.

### How add-on is claimed?

When client find out that membership has add-on fields, and the expiration date is passed, a normalization step should be performed by adding the add-on the expiration date so that user could continue access contents as smoothly as possible. In background, the client should send a request the API's "Claim Add-on" endpoint. The server-side will retrieve all unconsumed invoices for add-on type, sum up all of their years, months and days field, adding it to the current time.

The invoices will be categorized into *standard* and *premium* groups, and if both exists, we will use the premium group first, with the invoices in the standard group untouched.

## Definition

```typescript
interface Invoice {
    id: string;
    compounId: string;
    tier: 'standard' | 'premium';
    cycle: 'year' | 'month';
    years: number;
    months: number;
    days: number;
    addOnSource?: 'carry_over' | 'compensation' | 'user_purchase';
    appleTxId?: string;
    orderId?: string;
    orderKind: 'create' | 'renew' | 'upgrade' | 'add_on';
    paidAmount: number;
    payMethod: 'alipay' | 'wechat' | 'stripe' | 'apple' | 'b2b';
    priceId?: string;
    stripeSubsId?: string;
    createUtc: string;
    consumedUtc?: string;
    startUtc?: string;
    endUtc?: string;
    carriedOver?: string;
}
```

* `compoundId` is either ftc uuid or wechat union id the moment this invoice is generated, with ftc uuid having higher priority.
* `years`, `months` and `days` are the time unit to decide the subscription period we should grant to user the moment this invoice should be consumed.
* `addOnSource` if the `orderKind` field is `add_on`, this field records where does the add-on come from.
* `appleTxId` the original transaction id of IAP. This is used to record which IAP transaction caused this invoice to be generated. It occurs when an existing valid one-time-purchase user choose to use IAP and current remaining time is reserved for future use. This action is performed when webhook received Apple's notification. In such case `orderKind` is `add_on` and `addOnSource` is `carry_over`
* `orderId` the one-time-purchase order id which generated this invoice. In such case the `order_kind` could be either of `create` or `renew`, while the `addOnSource` is `user_purchase`. For upgrading from standard edition to premium edition, two invoices will be generated: one is `orderKind: upgrade` and `addOnSource: user_purchase`; the other is current membership's carry-over with `orderKind: add_on` and `addOnSource: carry_over`.
* `priceId` the pricing plan id of a one-time purchase. For `orderKind: create | renew | upgrade`, this is the plan user purchased. For `orderKid: carry_over`, this is carried over from current membership. This occurs when a one-time-purchase standard edition changed to premium edition.
* `stripeSubsId` the subscription id of stripe which caused this invoiced generated when a valid one-time-purchase user selected to switch to Stripe. An invoice of type `orderKind: add_on` and `addOnSource: carry_over` is created for current membership's remaining period.
* `consumedUtc` the moment when this invoice's `year`, `months` and `days` are transferred to membership's expiration time. For `orderKind: create | renew | upgrade`, an invoice is consumed immediately upon creation; for `orderKind: add_on`, it is consumed at a future time, usually upon current membership expires.
* `carriedOver` a time timstamp added to old invoices when current membership is carried over. When performing a carry-over to current membership, how do we know the existing invoices used to generate current membership are carried over? This field exists only for reference.
