#!/bin/bash

echo "Setting up authentication for tests..."

# Use hardcoded test credentials from cognito.tf
echo "Using test user credentials from configuration..."
TEST_EMAIL="testuser@example.com"
TEST_PASSWORD="TestPassword123!"
CLIENT_ID="1uq6q3bbn21s9fddqccv5tjlv0"

echo "Authenticating user: $TEST_EMAIL"

# Authenticate with Cognito
AUTH_RESPONSE=$(aws cognito-idp initiate-auth \
    --auth-flow USER_PASSWORD_AUTH \
    --client-id "$CLIENT_ID" \
    --auth-parameters USERNAME="$TEST_EMAIL",PASSWORD="$TEST_PASSWORD" \
    --region us-east-1 \
    --output json)

# Extract Access token (required for API authentication, not ID token)
ACCESS_TOKEN=$(echo "$AUTH_RESPONSE" | jq -r '.AuthenticationResult.AccessToken')

if [ "$ACCESS_TOKEN" = "null" ] || [ -z "$ACCESS_TOKEN" ]; then
    echo "Failed to get authentication token"
    echo "Response: $AUTH_RESPONSE"
    exit 1
fi

echo "Authentication successful!"

# Use correct API Gateway URL (fallback since terraform output may not work without AWS auth)
API_URL="https://0ioimabj53.execute-api.us-east-1.amazonaws.com/prod"

# Update dev.bru environment file
cat > "$(dirname "$0")/environments/dev.bru" << EOF
vars {
  base_url: $API_URL
  jwt_token: $ACCESS_TOKEN
}
EOF

echo "Updated dev.bru with new JWT token"
echo "Authentication setup complete! You can now run your tests."
