# or
curl -X POST https://api.clerk.com/v1/testing_tokens \
  -H "Authorization: Bearer YOUR_CLERK_SECRET_KEY"


# First get your user's ID from Clerk dashboard, then:
curl -X POST https://api.clerk.com/v1/sessions \
  -H "Authorization: Bearer YOUR_CLERK_SECRET_KEY" \
  -H "Content-Type: application/json" \
  -d '{"user_id": "user_abc123"}'

# Grab the session ID from the response, then get a token:
curl -X POST https://api.clerk.com/v1/sessions/sess_xyz/tokens \
  -H "Authorization: Bearer YOUR_CLERK_SECRET_KEY"
