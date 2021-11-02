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

### Workflow

1. Parse request body;
2. Validate request body;
3. Verify password. If password incorrect, respond 403 Forbidden;
4. Record client metadata;
5. Respond `Account`.

## Verify Email

## Request Mobile Auth Verification Code

```
PUT /auth/mobile/verification
```

### Request Body

```typescript
interface VerifierParams {
    mobile: string;
}
```

### Workflow

1. Parse request body, return 400 if invalid.
2. Retrieve user account by mobile.
3. Create a verification code. The user account found in the previous step, user id will be attached to this code.
4. Save the verification code to db.
5. Ask SMS service provider to send the code to user's device.
6. Returns 204 No Content if everything works.

## Verify Mobile Auth Code

```
POST /auth/mobile/verification
```

### Request Body

```typescript
interface VerifierParams {
    mobile: string;
    code: string;
    deviceToken?: string;
}
```

### Workflow

1. Parse quest body as JSON and validate.
2. Flag the verification code as used.
3. Return 200 OK with body:

```json
{
  "id": "string | null"
}
```

The returned result is the unique id of the user. If a user already has mobile set, it is always attached to the verification code; otherwise the id is null. This is a hint to client indicating whether user is logging in with mobile for the first time. If this is the first time of mobile-login, no user id could be found and client should ask user to link to an existing email account or perform signup; otherwise client should use the id to retrieve user's account data.

## Link Mobile to an Existing Email Account

```
POST /auth/mobile/link
```

Used after the previous step returned `id: null`.

### Request Body

```typescript
interface MobileLinkParams {
    email: string;
    password: string;
    mobile: string;
    devieToken?: string;
}
```

### Workflow

1. Parse request body and validate.
2. Verify email + password and retrieve the uuid of this email. If email is not found, returns 404; if password mismatched, returns 403 Forbidden.
3. Use the uuid retrieve in previous step to retrieve user account.
4. If the retrieve account already has a mobile set, check if the mobile in request body matches the one under this account. If matched, it indicates the mobile and email already linked. Return this account; otherwise the account is linked to another mobile, deny the request and returns reason:

```json
{
  "message": "This email account is already linked to another mobile",
  "error": {
    "field": "mobile",
    "code": "already_existss"
  }
}
```

5. Set mobile to this account's mobile field.
6. Save the mobile to db.
7. Return the updated account which is a [Account](./common_types.md) instance.

## Link Mobile to a New Email Account

```
POST /auth/mobile/link
```

User could choose to link to a new email account upon initial login. This one is similar to email signup step, with the only difference is that mobile number exists upon initial creation.

### Request Body

```typescript
interface MobileSignUpParams {
    email: string;
    password: string;
    mobile: string;
    deviceToken?: string;
    sourceURL?: string; // The base url to perform email verification since we need to send a verification letter upon creating email account. This enables you running web app on mutiple domains.
}
```

### Workflow

1. Parse request body and validate.
2. Create account. This step might fail if the mobile is already used by other accounts.
3. Return the new [Account](./common_types.md)
