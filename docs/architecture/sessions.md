# Author sessions

Access tokens live for 24 hours. Refresh sessions live for 30 days; on use with
less than seven days remaining, the server replaces the refresh token and grants
30 days from renewal. The previous refresh token becomes invalid. With regular
use the session continues; after 30 days without renewal the author signs in
again. Server time controls expiry.

The server stores only SHA-256 hashes of random refresh tokens in SQLite. Browser
refresh tokens are HttpOnly, SameSite=Strict cookies scoped to `/api/auth` and
Secure by default (plain HTTP is allowed only on loopback for development). Browser login/refresh/logout require `X-Edda-Session: cookie` to
use cookies; refresh secrets are omitted from their JSON responses. Non-browser
clients receive the pair as JSON. CLI credentials remain in the user's private
config directory, outside projects.

Web and CLI renew before an authenticated request when access has less than one
minute left, or refresh has less than seven days left. Web also retries one
rejected authenticated request after a successful refresh. Temporary network or
server failures preserve the session. Invalid refresh returns to login, retaining
the requested route and existing local file drafts. Web Locks serialize browser
session changes across tabs; the CLI uses a config-directory lock across commands.

CLI logout removes local credentials before attempting server revocation; an
unreachable server or a rejected revocation produces a warning without retaining
the saved login. Server revocation is bounded to ten seconds. Web logout requires
successful server revocation before clearing the browser session. Existing access
tokens remain valid until their 24-hour expiry; immediate revocation of access
JWTs is not implemented. Each device has a separate refresh session. Credentials
issued before this feature need one new login to obtain a refresh token.

## Rotation policy and its limits

Refresh credentials are reused until they enter the final seven-day window; they
are not rotated on every access renewal. This preserves the agreed 24-hour /
30-day / less-than-seven-days policy. Rotation can begin after day 23, but that is
not a maximum theft-exposure window: the original token remains valid until it
is rotated, revoked, or reaches its 30-day expiry. Anyone holding a valid token
can renew it and obtain its replacement. There is no token-family reuse detection
or automatic revocation of a whole lineage. Per-use rotation and reuse detection
would be a separate security-policy change, not an incidental bug fix.
