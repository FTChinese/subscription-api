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
