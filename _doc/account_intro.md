# Account Endpoints

## Authentication

See [details](./account_auth.md)

### With Email

* GET `/auth/email/exists?v=<email@example.org>` Check if an email exists.
* POST `/auth/email/login` Login with email
* POST `/auth/email/signup` Create a new account with email, with optional mobile.
* POST `/auth/email/verification/{token}` Verify email

### With Mobile

* PUT `/auth/mobile/verification` Send an SMS to user's device for authentication.
* POST `/auth/mobile/verification` Verify the SMS sent in the above step and returns an optional id linked to this mobile. Missing id indicates the mobile is used for the first time; otherwise client should retrieve account by the id.
* POST `/auth/mobile/link` If a mobile is used for the 1st time, user could link to an existing email account. If user wants to link to a new email account, send request to `/auth/email/signup`.
* POST `/auth/mobile/signup` Create a new account directly using a mobile in case user does not want to link to any email account. User will get an email derived from mobile, which is actually unusable.

### Password Reset

* POST `/auth/password-reset` Reset password
* POST `/auth/password-reset/letter` Send a password reset letter to user's email. The email will contain a link if request is sent in browser, or a 6-digit code if sent in native app.
* GET `/auth/password-reset/tokens/{token}` Verify the token send to user's email
* GET `/auth/password-reset/codes?email=<string>&code={number}` Verify the code for native app.

### With Wechat

* POST `/auth/wx/login` Wechat login
* PUT `/auth/wx/refresh` Refresh wechat login account.
* GET `/oauth/wx/callback/next-reader` Redirect for wechat login in browsers.

## Account Manipulation

Authorization header required.

See [details](./account_ftc.md)

* GET `/account` Load account
* DELETE `/account` Delete account
* PATCH `/account/email` Change email
* POST `/account/email/request-verification` Request an email verification letter
* PATCH `/account/name` Change username
* PATCH `/account/password` Change password
* PATCH `/account/mobile` Change mobile
* PATCH `/account/mobile/verification` Request a SMS before permitting mobile change.
* GET `/account/address` Load address
* PATCH `/account/address` Change address
* GET `/account/profile` Load profile
* PATCH `/account/profile` Change profile.
* POST `/account/wx` Load wechat account
* POST `/account/wx/signup` Wechat user creates a new email account and links to it.
* POST `/account/wx/link` Wechat and email link
* POST `/account/wx/unlink` Sever wechat-email links.
