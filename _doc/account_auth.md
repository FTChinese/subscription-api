# 邮箱、手机号以及微信登录

## Endpoints

* GET `/auth/email/exists?v=<email address>` Check if an email exists
* POST `/auth/email/login` Login with email + password
* POST `/auth/email/signup` Create a new email account
* POST `/auth/email/verification/{token}` Verify email inside browser
* PUT `/auth/mobile/verification` Create an SMS code and send to user mobile phone to perform login using phone number.
* POST `/auth/mobile/verification` Verify SMS code sent in last step.
* POST `/auth/mobile/link` Link mobile to an existing email account.
* POST `/auth/mobile/signup` When the mobile is used for login for the first time, user should be asked to link to an email account, and if the email has not signed-up, create one here.
* POST `/auth/password-reset` Reset password
* POST `/auth/password-reset/letter` Request a password reset email
* GET `/auth/password-reset/tokens/{token}` Verify password reset link in browser
* GET `/auth/password-reset/codes?email=<xxx>&code=<xxx>` Verify a password reset code in native apps.
* POST `/auth/wx/login` Wechat OAuth login
* PUT `/auth/wx/refresh` Refresh wechat OAuth if expired.

## Email Exists

```
GET /auth/email/exists?v=<email address>
```

### Workflow

1. Acquire email from URL parameters `v`;
2. Validate it is an email;
3. Check DB if the email exists.
4. Respond 404 if not found.
5. Respond 204 if found.

## Email Login

```
POST /auth/email/login
```

### Request Body

```typescript
interface EmailLoginParams {
    email: string;
    password: string;
    deviceToken?: string; // Required only for Android App
}
```

### Workflow

1. Parse request body;
2. Validate request body;
3. Verify password. If password incorrect, respond 403 Forbidden;
4. Record client metadata;
5. Respond `Account`.

## Email Signup

```
POST /auth/email/signup
```

### Request Body

```typescript
interface EmailLoginParams {
    email: string;
    password: string;
    deviceToken?: string;
}
```
