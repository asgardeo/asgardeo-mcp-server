# Self Service MCP Server

The Self Service MCP Server lets an AI agent manage **the signed-in user's own account** in Asgardeo or WSO2 Identity Server, using that user's access token. You host it over HTTP for the users of your own application.

With it, a user can ask their agent to:

- **Profile**: view and update their profile; add, remove or change the primary email and mobile number.
- **Sessions**: see where they're signed in; sign out one device or everywhere.
- **Sign-in security**: get an overview of MFA, passkeys, push devices, authenticator app and backup codes; rename or remove them.
- **Linked accounts**: list and unlink linked local accounts and social logins.
- **Apps**: list their apps; see which apps have access to their account and revoke access.
- **Consents**: see and withdraw app consents and accepted policies; unsubscribe from communication preferences.

Every tool acts only on the user the token belongs to.

---

## How to Use

### On Asgardeo / WSO2 Identity Server

1. **Create an application** that users sign in to (for example a Single Page or Traditional Web Application).
2. **Scopes**: request `internal_login`. It covers every tool.
3. **Get an access token**: sign in to the application as the user and copy the access token.

### Build

Build the executable from the repository root:
```bash
go build -o self-service-mcp ./cmd/self-service-mcp
```

### Host it (HTTP)

Run one server for all the users of your application. It serves the MCP Streamable HTTP transport at `/mcp`.

```bash
PORT=8000 BASE_URL=https://api.asgardeo.io/t/<asgardeo organization> ./self-service-mcp
```

- Your MCP client must send the access token the user receives when signing in to your application, as a bearer token in the `Authorization` header of each request (`Authorization: Bearer <user access token>`). The server uses this token for its calls to Asgardeo / WSO2 Identity Server, so each call acts as the user who sent it.
- The server keeps no session or token. Asgardeo / WSO2 Identity Server validates the token on every call.
- Serve it behind HTTPS (for example a reverse proxy or load balancer that terminates TLS).

### Configuration

| Variable | Required | Description |
|---|---|---|
| `BASE_URL` | yes | `https://api.asgardeo.io/t/<organization>` or `https://<is-host>/t/<tenant-domain>` |
| `TRANSPORT` | no | `http` (default) or `stdio` |
| `ACCESS_TOKEN` | stdio only | The user's access token |
| `PORT` | no | HTTP port. Default `8000` |
| `CLIENT_ID` | no | Client ID of the application the token came from. The agent keeps this app's access when revoking apps. |
| `PORTAL_NAME` | no | The name the agent uses for your self-service portal when it asks users to complete something there. Default `self-service portal` |
| `CERT_PATH` | no | CA certificate file, for WSO2 Identity Server with a self-signed or internal certificate. |

---

## Available Tools

Tools that change something ask the agent to confirm the exact change with the user first, and are marked with MCP tool annotations (`readOnlyHint` / `destructiveHint`) so clients can ask for approval.

| Area | Tool | Description |
|---|---|---|
| Profile | `get_profile` | Get the profile (optionally only some `attributes`) |
| | `list_profile_attributes` | List profile attributes and whether they can be changed |
| | `update_profile` | Update profile `attributes`, named as `list_profile_attributes` returns them |
| | `update_email` | Add or remove an email, or make one primary (`action`, `email`) |
| | `update_mobile` | Add or remove a mobile number, or make one primary (`action`, `mobile`) |
| Sessions | `list_sessions` | List active sessions |
| | `terminate_session` | Sign out one session (`session_id`) |
| | `terminate_all_sessions` | Sign out all sessions |
| Sign-in security | `get_security_overview` | Authenticator app, MFA options, backup codes, passkeys and push devices |
| | `rename_passkey` | Rename a passkey (`credential_id`, `display_name`) |
| | `remove_passkey` | Remove a passkey (`credential_id`) |
| | `remove_push_device` | Remove a push device (`device_id`) |
| | `remove_authenticator_app` | Remove the authenticator app |
| | `delete_backup_codes` | Delete the remaining backup codes |
| Linked accounts | `list_linked_accounts` | List linked local accounts and social logins |
| | `unlink_local_account` | Unlink a local account (`user_id`) |
| | `unlink_social_account` | Unlink a social login (`id`) |
| | `unlink_all_local_accounts` | Unlink all local accounts |
| | `unlink_all_social_accounts` | Unlink all social logins |
| Apps | `list_my_applications` | List the user's applications (`filter`, `limit`, `offset`) |
| | `get_my_application` | Get one application (`application_id`) |
| | `list_authorized_apps` | List apps with access to the account |
| | `revoke_app_access` | Revoke an app's access (`application_id`) |
| Consents | `list_consents` | List app consents, policies and communication preferences (`state`) |
| | `get_consent` | Get a policy or communication preference (`consent_id`) |
| | `get_app_consent` | Get the details of an app consent (`consent_id`) |
| | `withdraw_app_consent` | Withdraw an app consent (`consent_id`) |
| | `revoke_consent` | Withdraw a policy or unsubscribe from a communication preference (`consent_id`) |

---

## Example Prompts

```
Where am I signed in? Sign out the Windows session.
```
```
Change my last name to Perera.
```
```
How secure is my account?
```
```
Which apps have access to my account?
```
```
What have I consented to?
```

---

## Troubleshooting

- **401**: the access token is missing, expired or revoked. The user needs to sign in again to get a new token.
- **403**: the token doesn't have the `internal_login` scope.
- **404**: the feature isn't available on this server version.
