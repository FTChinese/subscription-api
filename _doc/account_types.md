# Data Types Related to User Account

## Type Annotation

Here we use TypeScript interface to denote data types.

## Account

A user's full account consists of 3 parts: base account data, wechat user info, and membership.

### Base Account

```typescript
interface BaseAccount {
    id: string; // FTC uuid
    unionId?: string; // Wchat union id
    stripeId?: string; // Stripe customer id
    email: string;
    mobile?: string;
    userName?: string;
    avatarUrl?: string;
    isVerified: boolean; // Is email everified?
    campaignCode?: string;
}
```

### Wechat

```typescript
interface Wechat {
    nickname?: string;
    avatarUrl?: string;
}
```

### Membership

```typescript
interface Membership {
    ftcId: string | null;
    unionId: string | null,
    tier: 'standard' | 'premium' | null;
    cycle: 'month' | 'year' | null;
    expireDate: string | null;
    payMethod: 'alipay' | 'wechat' | 'stripe' | 'apple' | 'b2b' | null;
    ftcPlanId: string | null;
    stripeSubsId: string | null;
    autoRenew: boolean;
    status?: 'active' | 'canceled' | 'incomplete' | 'incomplete_expired' | 'past_due' | 'trialing' | 'unpaid' | null; // Stripe subscription status
    appleSubsId: string | null;
    b2bLicenceId: string | null;
    standardAddOn: number; // Default 0
    premiumAddOn: number;
    vip: boolean; // Default false
}
```

The combination of some fields are always mutually exclusive:

* When `payMethod` is either `alipay` or `wechat`, `ftcPlanId` must not be null and `autoRenew` must be `false`;
* When `payMethod` is `apple`, `appleSubsId` must not be null, and `autoRenew` exists;
* When `payMethod` is `stripe`, `stripeSubsId`, `status` must not be null, and `autoRenew` exists

When using any one of the above three groups, other payment methods' associated fields must be null.

### Full Account

```typescript
interface Account extends BaseAccount {
    loginMethod: 'email' | 'wechat' | 'mobile';
    wechat: Wechat;
    membership: Membership;
}
```

When injected into WebView in Android, it is defined as a global variable `androidUserInfo`.

## Address

```typescript
interface Address {
    country?: string;
    province?: string;
    city?: string;
    district?: string;
    street?: string;
    postcode?: string;
}
```

When injected into WebView in Android, it is defined as a global variable `androidUserAddress`.


## Client Metadata

```typescript
interface Client {
    platform?: 'web' | 'ios' | 'android'; // Which platform user is on
    client_version?: string; // The client app version that send request to API.
    user_ip?: string; // Client app ip.
    user_agent?: string; // User agent.
}
```

Client is used to record client-side app's metadata. It is optional, but you are recommended to provide them for some endpoints so that we could detect erratic behavior of user.

These fields are acquired from request header:

* `X-Client-Type` for `platform`;
* `X-Client-Version` for `client_version`;

For web browser, you should provide these headers:

* `X-User-Ip` for `user_ip`;
* `X-User-Agent` for `user_agent`.

In web browser, the request is sent to API via our app running on our server; therefore the web app should forward browser's ip and user agent.

For native apps, the requests are usually sent by an HTTP client such as okhttp, the ip and user agent could be acquired from the client request's default header fields.
