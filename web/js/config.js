// Configuration file for Bookshelf app
// This file contains environment-specific configuration
// Auto-generated from Terraform outputs - DO NOT EDIT MANUALLY
const APP_CONFIG = {
    // API Gateway URL
    API_BASE_URL: 'https://o4egm6gq21.execute-api.us-east-1.amazonaws.com/prod',
    
    // Cognito configuration
    COGNITO: {
        userPoolId: 'us-east-1_3gazjYas5',
        clientId: 'kiulela0t8ui1hqpr3hoe8s2d',
        region: 'us-east-1'
    }
};

// Make configuration available globally
window.APP_CONFIG = APP_CONFIG;
