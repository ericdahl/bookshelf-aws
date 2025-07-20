// Configuration file for Bookshelf app
// This file contains environment-specific configuration
// Auto-generated from Terraform outputs - DO NOT EDIT MANUALLY
const APP_CONFIG = {
    // API Gateway URL
    API_BASE_URL: 'https://qklqndn9th.execute-api.us-east-1.amazonaws.com/prod',
    
    // Cognito configuration
    COGNITO: {
        userPoolId: 'us-east-1_iBy6QgduI',
        clientId: '5ulggimd8m012tc6f5fjknphfv',
        region: 'us-east-1'
    }
};

// Make configuration available globally
window.APP_CONFIG = APP_CONFIG;
