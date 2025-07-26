const APP_CONFIG = {
    API_BASE_URL: '/api',
    
    COGNITO: {
        userPoolId: '${cognito_user_pool_id}',
        clientId: '${cognito_user_pool_client_id}',
        region: '${aws_region}'
    }
};

// Make configuration available globally
window.APP_CONFIG = APP_CONFIG;
