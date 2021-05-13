# Data Types Related to User Account

## Type Annotation

Here we use TypeScript interface to denote data types.

## Account

A user's full account consists of 3 parts: email account data, wechat user info, and membership.

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

interface Wechat {
    nickname?: string;
    avatarUrl?: string;
}

interface Membership {
    tier?: 'standard' | 'premium';
    cycle?: 'month' | 'year';
    expireDate?: string;
    payMethod?: 'alipay' | 'wechat' | 'stripe' | 'apple' | 'b2b';
    stripeSubsId?: string;
    autoRenew: boolean;
    status?: 'active' | 'canceled' | 'incomplete' | 'incomplete_expired' | 'past_due' | 'trialing' | 'unpaid'; // Stripe subscription status
    b2bLicenceId?: string;
    standardAddOn: number; // Default 0
    premiumAddOn: number;
    vip: boolean;
}

// The complete account data.
interface Account extends BaseAccount {
    loginMethod: 'email' | 'wechat' | 'mobile';
    wechat: Wechat;
    membership: Membership;
}
```


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
