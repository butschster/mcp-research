-- PKCE support for OAuth authorization codes
ALTER TABLE oauth_codes ADD COLUMN code_challenge TEXT NOT NULL DEFAULT '';
ALTER TABLE oauth_codes ADD COLUMN code_challenge_method TEXT NOT NULL DEFAULT '';
