# Casdoor Integration Guide

This document explains how to run `raidark` against a Casdoor server, which values are required, where each value comes from, and how to validate the full authentication flow end to end.

## Integration Model in Raidark

`raidark` already includes a Casdoor-backed authentication provider. The current flow is:

1. The user is redirected to Casdoor's OAuth authorization page.
2. Casdoor redirects the browser back to your application's callback URL with `code` and `state`.
3. Your frontend or callback handler sends that `code` and `state` to `POST /auth/exchange`.
4. `raidark` exchanges the authorization code for tokens, validates the access token, stores a local session, returns the access token, and sets the `app_session` cookie.
5. Protected `raidark` routes accept the returned Bearer token.

Important implementation detail:

- `raidark` does not expose a `/callback` route by itself.
- `CASDOOR_REDIRECT_URI` must therefore point to a frontend route, reverse proxy route, or dedicated callback page that you own.
- That callback page is responsible for reading `code` and `state` from the URL and forwarding them to `POST /auth/exchange`.

## What You Need

Before enabling Casdoor in `raidark`, make sure you have:

- A running Casdoor server.
- An administrator account in Casdoor.
- A Casdoor organization for the users of your application.
- A Casdoor application for `raidark`.
- A JWT certificate assigned to that Casdoor application.
- A frontend callback URL that can receive `code` and `state`.
- A persistent `raidark` datastore, because auth sessions are stored in the database.

## Environment Variables

`raidark` reads the following values for the Casdoor provider.

### Required Variables

| Variable | Required | What it means | Where to get it / what it should contain |
| --- | --- | --- | --- |
| `AUTH_PROVIDER_TYPE` | Yes | Selects the auth provider implementation. Must be `casdoor` to enable Casdoor. | Set it manually in `.env`. |
| `CASDOOR_ENDPOINT` | Yes | Base URL of your Casdoor server. | The public base URL of your Casdoor instance, for example `http://localhost:8000` or `https://sso.example.com`. Do not add a trailing slash. |
| `CASDOOR_CLIENT_ID` | Yes | OAuth client ID of the Casdoor application used by `raidark`. | Copy it from the Casdoor application details page after creating the application. |
| `CASDOOR_CLIENT_SECRET` | Yes | OAuth client secret of the Casdoor application used by `raidark`. | Copy it from the Casdoor application details page after creating the application. Treat it as a secret. |
| `CASDOOR_CERTIFICATE` | Yes | Public key used to validate JWTs issued for the application. | Create or reuse a JWT certificate in Casdoor, open the certificate edit page, and copy or download the public key content. This is the public key, not the private key. |
| `CASDOOR_ORGANIZATION` | Yes | Casdoor organization name that owns the application and users. | The organization `Name` value in Casdoor, for example `acme` or `raidark-dev`. |
| `CASDOOR_APPLICATION` | Yes | Casdoor application name used by `raidark`. | The application `Name` value in Casdoor, for example `raidark-web`. |

### Required in Practice

| Variable | Required | What it means | Where to get it / what it should contain |
| --- | --- | --- | --- |
| `CASDOOR_REDIRECT_URI` | Yes in real deployments | The URL Casdoor redirects the browser to after login. | This must be a callback page in your frontend or gateway, for example `http://localhost:3000/callback`. It must match the Redirect URL configured in the Casdoor application. Even though `raidark` has a default value, you should set it explicitly because `raidark` does not provide this route automatically. |

### Minimal Example

```env
AUTH_PROVIDER_TYPE=casdoor

CASDOOR_ENDPOINT=http://localhost:8000
CASDOOR_CLIENT_ID=your_client_id
CASDOOR_CLIENT_SECRET=your_client_secret
CASDOOR_CERTIFICATE="-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAt...
-----END PUBLIC KEY-----"
CASDOOR_ORGANIZATION=raidark-dev
CASDOOR_APPLICATION=raidark-web
CASDOOR_REDIRECT_URI=http://localhost:3000/callback

DATASTORE_TYPE=sqlite
DB_DATABASE=raidark.db
API_PORT=8080
```

Notes:

- Keep the full PEM-formatted public key in `CASDOOR_CERTIFICATE`.
- If your shell or deployment platform handles multiline values poorly, store the certificate through your secret manager or environment injection mechanism instead of manually pasting it in a local shell export.
- `raidark` defaults `CASDOOR_ENDPOINT` to `http://localhost:8000` and `CASDOOR_REDIRECT_URI` to `http://localhost:8080/callback`, but the second default usually does not match a real setup.

## Casdoor Setup

### 1. Create the organization

In Casdoor:

1. Sign in as an administrator.
2. Open the Organizations section.
3. Create a new organization for your application users.
4. Choose a stable organization name, for example `raidark-dev`.
5. Save the organization.

Use that organization `Name` as:

```env
CASDOOR_ORGANIZATION=raidark-dev
```

### 2. Create or select a JWT certificate

In Casdoor:

1. Open the Certs section.
2. Create a JWT certificate if you do not already have one.
3. Open the certificate details or edit page.
4. Copy or download the public key.

Use the public key content as:

```env
CASDOOR_CERTIFICATE="-----BEGIN PUBLIC KEY----- ... -----END PUBLIC KEY-----"
```

Important:

- `raidark` needs the public key because it validates the JWT access token returned by Casdoor.
- Do not put the private key in `CASDOOR_CERTIFICATE`.

### 3. Create the Casdoor application

In Casdoor:

1. Open the Applications section.
2. Create a new application under the organization created above.
3. Give it a stable application name, for example `raidark-web`.
4. Set the Redirect URL to the callback page owned by your application, for example:

```text
http://localhost:3000/callback
```

5. Assign the JWT certificate you created or selected in the previous step as the token certificate for the application.
6. Save the application.
7. Copy the application's Client ID and Client Secret.

Use those values as:

```env
CASDOOR_APPLICATION=raidark-web
CASDOOR_CLIENT_ID=...
CASDOOR_CLIENT_SECRET=...
CASDOOR_REDIRECT_URI=http://localhost:3000/callback
```

Recommended checks on the application:

- The application belongs to the same organization configured in `CASDOOR_ORGANIZATION`.
- The Redirect URL exactly matches `CASDOOR_REDIRECT_URI`.
- The application has a token certificate assigned.

### 4. Create a test user

In Casdoor:

1. Open the Users section.
2. Create a user inside the same organization.
3. Confirm the account can sign in through the Casdoor UI.

This gives you a known user for end-to-end testing.

## Raidark Setup

After Casdoor is ready:

1. Put the Casdoor variables in your `.env`.
2. Run database migrations so the auth session model exists.
3. Start the API.

Example:

```bash
go run ./main dbmigrate
go run ./main api
```

`raidark` exposes these auth endpoints:

- `POST /auth/exchange`
- `POST /auth/refresh`
- `POST /auth/logout`

Protected API groups created with `AuthenticatedRootModule(...)` expect:

- `Authorization: Bearer <access_token>`

## Login URL

The current `raidark` Casdoor provider builds the authorization URL in this form:

```text
{CASDOOR_ENDPOINT}/login/oauth/authorize
  ?client_id={CASDOOR_CLIENT_ID}
  &redirect_uri={CASDOOR_REDIRECT_URI}
  &response_type=code
  &scope=openid%20profile%20email
  &state={opaque_random_state}
```

Example:

```text
http://localhost:8000/login/oauth/authorize?client_id=abc123&redirect_uri=http%3A%2F%2Flocalhost%3A3000%2Fcallback&response_type=code&scope=openid%20profile%20email&state=test-state-123
```

Recommendations:

- Generate a random `state` value per login attempt.
- Store the expected `state` client-side and verify it after the redirect before calling `/auth/exchange`.
- Use HTTPS in production.

## How the Frontend Gets the Authorization Code

The `code` does not exist until the browser is redirected to Casdoor first.

The complete browser flow is:

1. The user clicks `Sign in`.
2. Your frontend generates a random `state` value.
3. Your frontend stores that `state` locally so it can validate it after the redirect.
4. Your frontend builds the Casdoor authorization URL.
5. The browser is redirected to Casdoor.
6. The user signs in on Casdoor.
7. Casdoor redirects the browser back to `CASDOOR_REDIRECT_URI` with `code` and `state` in the query string.
8. Your callback page validates the returned `state`.
9. Your callback page sends `code` and `state` to `POST /auth/exchange`.
10. `raidark` returns the access token and sets the `app_session` cookie.

### What the frontend must know before redirecting

Your frontend needs these values:

- `CASDOOR_ENDPOINT`
- `CASDOOR_CLIENT_ID`
- `CASDOOR_REDIRECT_URI`
- a generated `state` value

The frontend does not need:

- `CASDOOR_CLIENT_SECRET`
- `CASDOOR_CERTIFICATE`

Those values must stay server-side.

### Example login button flow

This example shows the missing step: the request to Casdoor that produces the authorization `code`.

```ts
function generateState(): string {
  return crypto.randomUUID();
}

function buildCasdoorLoginUrl() {
  const casdoorEndpoint = "http://localhost:8000";
  const clientId = "your_client_id";
  const redirectUri = "http://localhost:3000/callback";
  const state = generateState();

  sessionStorage.setItem("casdoor_oauth_state", state);

  const url = new URL("/login/oauth/authorize", casdoorEndpoint);
  url.searchParams.set("client_id", clientId);
  url.searchParams.set("redirect_uri", redirectUri);
  url.searchParams.set("response_type", "code");
  url.searchParams.set("scope", "openid profile email");
  url.searchParams.set("state", state);

  return url.toString();
}

function signInWithCasdoor() {
  window.location.href = buildCasdoorLoginUrl();
}
```

When `signInWithCasdoor()` runs:

1. The browser navigates to Casdoor.
2. Casdoor renders the login page.
3. After successful authentication, Casdoor redirects to:

```text
http://localhost:3000/callback?code=AUTHORIZATION_CODE&state=THE_SAME_STATE
```

That redirected URL is how the frontend obtains the `code`.

### Example callback page flow

After Casdoor redirects back, the callback page should:

1. Read `code` and `state` from the query string.
2. Read the originally stored `state`.
3. Reject the flow if the values do not match.
4. Call `raidark` at `POST /auth/exchange`.
5. Clear the temporary stored `state`.

```ts
async function handleCasdoorCallback() {
  const params = new URLSearchParams(window.location.search);
  const code = params.get("code");
  const returnedState = params.get("state");
  const expectedState = sessionStorage.getItem("casdoor_oauth_state");

  if (!code || !returnedState) {
    throw new Error("Missing code or state in callback URL");
  }

  if (!expectedState || expectedState !== returnedState) {
    throw new Error("Invalid OAuth state");
  }

  const response = await fetch(
    `http://localhost:8080/auth/exchange?code=${encodeURIComponent(code)}&state=${encodeURIComponent(returnedState)}`,
    {
      method: "POST",
      credentials: "include",
    },
  );

  sessionStorage.removeItem("casdoor_oauth_state");

  if (!response.ok) {
    throw new Error("Authentication failed");
  }

  const session = await response.json();
  return session;
}
```

Expected result:

- `raidark` returns an `access_token`.
- `raidark` sets the `app_session` cookie.
- The frontend can now call protected endpoints with `Authorization: Bearer <access_token>`.

### Recommended frontend structure

For a production-ready setup, separate the frontend into two clear steps:

- A login action or login page that only builds the Casdoor authorization URL and redirects the browser.
- A callback page that only validates `state`, calls `/auth/exchange`, and then redirects the user to the authenticated area of the application.

### Common mistakes

- Trying to call `/auth/exchange` before the browser has been redirected to Casdoor.
- Forgetting to persist `state` before redirecting.
- Using a `CASDOOR_REDIRECT_URI` that does not exactly match the Redirect URL configured in Casdoor.
- Exposing `CASDOOR_CLIENT_SECRET` in the frontend.
- Forgetting `credentials: "include"` when the frontend expects the browser to store `app_session`.

## End-to-End Test Flow

This flow validates the full path from Casdoor login to a protected `raidark` endpoint.

### Phase 1: Prepare Casdoor

1. Create the organization in Casdoor.
2. Create a JWT certificate in Casdoor.
3. Create the application in Casdoor.
4. Set the application's Redirect URL to your frontend callback URL.
5. Assign the JWT certificate to the application.
6. Copy the Client ID, Client Secret, application name, and organization name.
7. Create a test user in the same organization.

### Phase 2: Prepare Raidark

1. Update `.env` to use `AUTH_PROVIDER_TYPE=casdoor`.
2. Fill every `CASDOOR_*` variable with the real values from Casdoor.
3. Run:

```bash
go run ./main dbmigrate
go run ./main api
```

4. Confirm the API starts successfully.

### Phase 3: Trigger Login

1. Build the Casdoor authorization URL using the values above.
2. Open that URL in a browser.
3. Sign in using the test user created in Casdoor.
4. Confirm Casdoor redirects back to your callback page with `code` and `state`.

Expected result:

- Your browser lands on `CASDOOR_REDIRECT_URI`.
- The callback URL contains `?code=...&state=...`.

### Phase 4: Exchange the Code in Raidark

From the callback page, call:

```bash
curl -i -X POST "http://localhost:8080/auth/exchange?code=THE_CODE&state=THE_STATE"
```

Expected result:

- HTTP `200 OK`
- JSON response with:
  - `access_token`
  - `token_type`
  - `expires_in`
  - `user.id`
  - `user.username`
  - `user.name`
  - `user.email`
- `Set-Cookie: app_session=...`

Important behavior:

- `raidark` stores the refresh token server-side in its session table.
- The refresh token is not returned to the frontend.

### Phase 5: Call a Protected Endpoint

Use the access token returned by `/auth/exchange`:

```bash
curl -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  http://localhost:8080/api/v1/ping
```

Expected result:

- The protected endpoint accepts the token.
- The request is authenticated with the claims parsed from the Casdoor JWT.

### Phase 6: Validate Token Refresh

Call the refresh endpoint with the `app_session` cookie:

```bash
curl -i -X POST "http://localhost:8080/auth/refresh" \
  --cookie "app_session=YOUR_SESSION_ID"
```

Expected result:

- HTTP `200 OK`
- JSON response with a new `access_token`
- The `app_session` cookie remains valid or is renewed

### Phase 7: Validate Logout

Call:

```bash
curl -i -X POST "http://localhost:8080/auth/logout" \
  --cookie "app_session=YOUR_SESSION_ID"
```

Expected result:

- HTTP `200 OK`
- A success response from `raidark`
- The `app_session` cookie is cleared

## Troubleshooting Checklist

If authentication fails, check these items first:

- `AUTH_PROVIDER_TYPE` is `casdoor`.
- `CASDOOR_ENDPOINT` points to the actual Casdoor base URL.
- `CASDOOR_REDIRECT_URI` exactly matches the Redirect URL configured in Casdoor.
- `CASDOOR_ORGANIZATION` matches the application's organization name.
- `CASDOOR_APPLICATION` matches the application's name.
- `CASDOOR_CLIENT_ID` and `CASDOOR_CLIENT_SECRET` were copied from the same Casdoor application.
- `CASDOOR_CERTIFICATE` contains the JWT public key, not a private key and not an unrelated certificate.
- You are sending both `code` and `state` to `POST /auth/exchange`.
- You ran `go run ./main dbmigrate` before testing.
- Your frontend sends requests to `/auth/exchange` with `credentials: "include"` if you expect the session cookie to be stored by the browser.

## Production Notes

- Set secure cookie behavior when serving over HTTPS. The current controllers create cookies with `Secure: false`, so production deployments should harden that behavior.
- Protect `CASDOOR_CLIENT_SECRET` with a secret manager.
- Use a real persistent database instead of ephemeral SQLite where appropriate.
- Keep the callback URL stable across environments and define one Casdoor application per environment when possible.

## References

- Casdoor application configuration: https://casdoor.org/docs/application/config/
- Casdoor SDK configuration: https://casdoor.org/docs/how-to-connect/sdk
- Casdoor certificate overview: https://casdoor.org/docs/cert/overview
- Casdoor core concepts: https://casdoor.org/docs/basic/core-concepts
