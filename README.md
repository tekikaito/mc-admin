# mc-admin

## Discord OAuth setup

1. Create an application in the [Discord Developer Portal](https://discord.com/developers/applications) and add an OAuth redirect URL that points to `https://your-domain/auth/discord/callback` (or `http://localhost:8080/auth/discord/callback` for local development).
2. Copy the Client ID and Client Secret for the application and set the following environment variables before running the server:

   ```bash
   export DISCORD_CLIENT_ID="your-client-id"
   export DISCORD_CLIENT_SECRET="your-client-secret"
   export DISCORD_REDIRECT_URI="http://localhost:8080/auth/discord/callback"
   export SESSION_SECRET="random-long-string"
   export DISCORD_ALLOWED_USER_IDS="123,456"   # optional allow-list
   ```

3. Start the web server as usual. All routes besides `/auth/*` now require a valid Discord session, and only the accounts listed in `DISCORD_ALLOWED_USER_IDS` (if provided) can sign in.
