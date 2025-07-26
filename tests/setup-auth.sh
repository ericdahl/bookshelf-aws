#!/bin/bash

echo "Setting up authentication for tests..."

# Use hardcoded test credentials from cognito.tf
echo "Using test user credentials from configuration..."
TEST_EMAIL="testuser@example.com"
TEST_PASSWORD="TestPassword123!"
CLIENT_ID="6fh10m5n9l23o93t2ofl33j0lp"

echo "Authenticating user: $TEST_EMAIL"

# Authenticate with Cognito
AUTH_RESPONSE=$(aws cognito-idp initiate-auth \
    --auth-flow USER_PASSWORD_AUTH \
    --client-id "$CLIENT_ID" \
    --auth-parameters USERNAME="$TEST_EMAIL",PASSWORD="$TEST_PASSWORD" \
    --region us-east-1 \
    --output json)

# Extract ID token (required for API Gateway JWT authorizer)
ID_TOKEN=$(echo "$AUTH_RESPONSE" | jq -r '.AuthenticationResult.IdToken')

if [ "$ID_TOKEN" = "null" ] || [ -z "$ID_TOKEN" ]; then
    echo "Failed to get authentication token"
    echo "Response: $AUTH_RESPONSE"
    exit 1
fi

echo "Authentication successful!"

# Use correct API Gateway URL (fallback since terraform output may not work without AWS auth)
API_URL="https://wsl1l84gmi.execute-api.us-east-1.amazonaws.com/prod"

# Update dev.bru environment file
cat > "$(dirname "$0")/environments/dev.bru" << EOF
vars {
  base_url: $API_URL
  jwt_token: $ID_TOKEN
}
EOF

echo "Updated dev.bru with new JWT token"
echo "Authentication setup complete! You can now run your tests."
