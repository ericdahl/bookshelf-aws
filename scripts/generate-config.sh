#!/bin/bash

# Generate configuration file from Terraform outputs
# This script reads Terraform outputs and generates web/js/config.js

set -e  # Exit on error

echo "Generating configuration file from Terraform outputs..."

# Check if terraform is available
if ! command -v terraform &> /dev/null; then
    echo "Error: terraform command not found"
    exit 1
fi

# Get Terraform outputs
API_URL=$(terraform output -raw api_gateway_invoke_url)
USER_POOL_ID=$(terraform output -raw cognito_user_pool_id)
CLIENT_ID=$(terraform output -raw cognito_user_pool_client_id)

# Validate outputs
if [ -z "$API_URL" ] || [ -z "$USER_POOL_ID" ] || [ -z "$CLIENT_ID" ]; then
    echo "Error: Failed to get all required Terraform outputs"
    exit 1
fi

# Generate config.js
cat > web/js/config.js << EOF
// Configuration file for Bookshelf app
// This file contains environment-specific configuration
// Auto-generated from Terraform outputs - DO NOT EDIT MANUALLY
const APP_CONFIG = {
    // API Gateway URL
    API_BASE_URL: '${API_URL}',
    
    // Cognito configuration
    COGNITO: {
        userPoolId: '${USER_POOL_ID}',
        clientId: '${CLIENT_ID}',
        region: 'us-east-1'
    }
};

// Make configuration available globally
window.APP_CONFIG = APP_CONFIG;
EOF

echo "Configuration file generated at web/js/config.js"
echo "API_BASE_URL: ${API_URL}"
echo "COGNITO.userPoolId: ${USER_POOL_ID}"
echo "COGNITO.clientId: ${CLIENT_ID}"