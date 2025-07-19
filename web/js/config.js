// Configuration file for Bookshelf app
// This file contains environment-specific configuration
// Auto-generated from Terraform outputs - DO NOT EDIT MANUALLY
const APP_CONFIG = {
    // API Gateway URL
    API_BASE_URL: 'https://wsl1l84gmi.execute-api.us-east-1.amazonaws.com/prod',
    
    // Cognito configuration
    COGNITO: {
        userPoolId: 'us-east-1_BzQaDh104',
        clientId: '1uq6q3bbn21s9fddqccv5tjlv0',
        region: 'us-east-1'
    }
};

// Make configuration available globally
window.APP_CONFIG = APP_CONFIG;
