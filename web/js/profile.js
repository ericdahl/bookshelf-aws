document.addEventListener('DOMContentLoaded', function() {
    // Check authentication before loading profile
    if (!checkAuthentication()) {
        window.location.href = 'signin.html';
        return;
    }

    // DOM Elements
    const profileForm = document.getElementById('profile-form');
    const emailInput = document.getElementById('email');
    const timezoneSelect = document.getElementById('timezone');
    const saveButton = document.getElementById('save-button');
    const loadingOverlay = document.getElementById('loading-overlay');
    const successMessage = document.getElementById('success-message');
    const errorMessage = document.getElementById('error-message');

    // Initialize the profile page
    initProfile();

    // Initialize profile page
    function initProfile() {
        loadUserProfile();
        populateTimezones();
        setupEventListeners();
    }

    // Setup event listeners
    function setupEventListeners() {
        profileForm.addEventListener('submit', handleFormSubmit);
    }

    // Load user profile data
    function loadUserProfile() {
        showLoading();
        
        // Get user email from localStorage
        const userEmail = localStorage.getItem('userEmail');
        if (userEmail) {
            emailInput.value = userEmail;
        }

        // Get user attributes from Cognito
        getUserAttributes()
            .then(attributes => {
                // Look for zoneinfo attribute
                const timezone = attributes.find(attr => attr.getName() === 'zoneinfo');
                if (timezone) {
                    timezoneSelect.value = timezone.getValue();
                }
                hideLoading();
            })
            .catch(error => {
                console.error('Error loading user profile:', error);
                showError('Failed to load user profile. Please try again.');
                hideLoading();
            });
    }

    // Get user attributes from Cognito
    function getUserAttributes() {
        return new Promise((resolve, reject) => {
            try {
                const poolData = {
                    UserPoolId: window.APP_CONFIG.COGNITO.userPoolId,
                    ClientId: window.APP_CONFIG.COGNITO.clientId
                };
                
                const userPool = new AmazonCognitoIdentity.CognitoUserPool(poolData);
                const cognitoUser = userPool.getCurrentUser();
                
                if (!cognitoUser) {
                    reject(new Error('No current user found'));
                    return;
                }
                
                cognitoUser.getSession((err, session) => {
                    if (err) {
                        reject(err);
                        return;
                    }
                    
                    if (!session.isValid()) {
                        reject(new Error('Invalid session'));
                        return;
                    }
                    
                    cognitoUser.getUserAttributes((err, attributes) => {
                        if (err) {
                            reject(err);
                            return;
                        }
                        
                        resolve(attributes || []);
                    });
                });
            } catch (error) {
                reject(error);
            }
        });
    }

    // Update user attributes in Cognito
    function updateUserAttributes(attributes) {
        return new Promise((resolve, reject) => {
            try {
                const poolData = {
                    UserPoolId: window.APP_CONFIG.COGNITO.userPoolId,
                    ClientId: window.APP_CONFIG.COGNITO.clientId
                };
                
                const userPool = new AmazonCognitoIdentity.CognitoUserPool(poolData);
                const cognitoUser = userPool.getCurrentUser();
                
                if (!cognitoUser) {
                    reject(new Error('No current user found'));
                    return;
                }
                
                cognitoUser.getSession((err, session) => {
                    if (err) {
                        reject(err);
                        return;
                    }
                    
                    if (!session.isValid()) {
                        reject(new Error('Invalid session'));
                        return;
                    }
                    
                    // Convert attributes to Cognito format
                    const cognitoAttributes = attributes.map(attr => 
                        new AmazonCognitoIdentity.CognitoUserAttribute(attr)
                    );
                    
                    cognitoUser.updateAttributes(cognitoAttributes, (err, result) => {
                        if (err) {
                            reject(err);
                            return;
                        }
                        
                        resolve(result);
                    });
                });
            } catch (error) {
                reject(error);
            }
        });
    }

    // Populate timezone select options
    function populateTimezones() {
        // List of common timezones
        const timezones = [
            { value: 'America/New_York', label: 'Eastern Time (US & Canada)' },
            { value: 'America/Chicago', label: 'Central Time (US & Canada)' },
            { value: 'America/Denver', label: 'Mountain Time (US & Canada)' },
            { value: 'America/Los_Angeles', label: 'Pacific Time (US & Canada)' },
            { value: 'America/Anchorage', label: 'Alaska' },
            { value: 'Pacific/Honolulu', label: 'Hawaii' },
            { value: 'Europe/London', label: 'London' },
            { value: 'Europe/Paris', label: 'Paris, Berlin, Rome' },
            { value: 'Europe/Moscow', label: 'Moscow' },
            { value: 'Asia/Tokyo', label: 'Tokyo' },
            { value: 'Asia/Shanghai', label: 'Beijing, Shanghai' },
            { value: 'Asia/Kolkata', label: 'Mumbai, Delhi, Kolkata' },
            { value: 'Asia/Dubai', label: 'Dubai' },
            { value: 'Australia/Sydney', label: 'Sydney, Melbourne' },
            { value: 'Pacific/Auckland', label: 'Auckland' },
            { value: 'America/Sao_Paulo', label: 'Brasilia, São Paulo' },
            { value: 'America/Mexico_City', label: 'Mexico City' },
            { value: 'America/Toronto', label: 'Toronto' },
            { value: 'America/Vancouver', label: 'Vancouver' },
            { value: 'Africa/Cairo', label: 'Cairo' },
            { value: 'Africa/Johannesburg', label: 'Johannesburg' },
            { value: 'UTC', label: 'UTC (Coordinated Universal Time)' }
        ];

        // Sort timezones alphabetically by label
        timezones.sort((a, b) => a.label.localeCompare(b.label));

        // Add options to select
        timezones.forEach(tz => {
            const option = document.createElement('option');
            option.value = tz.value;
            option.textContent = tz.label;
            timezoneSelect.appendChild(option);
        });
    }

    // Handle form submission
    function handleFormSubmit(event) {
        event.preventDefault();
        
        const timezone = timezoneSelect.value;
        
        if (!timezone) {
            showError('Please select a timezone.');
            return;
        }

        showLoading();
        hideMessages();

        // Prepare attributes to update
        const attributes = [
            {
                Name: 'zoneinfo',
                Value: timezone
            }
        ];

        // Update user attributes in Cognito
        updateUserAttributes(attributes)
            .then(() => {
                hideLoading();
                showSuccess('Profile updated successfully!');
            })
            .catch(error => {
                console.error('Error updating profile:', error);
                hideLoading();
                showError('Failed to update profile. Please try again.');
            });
    }

    // Show loading overlay
    function showLoading() {
        loadingOverlay.classList.remove('hidden');
        saveButton.disabled = true;
    }

    // Hide loading overlay
    function hideLoading() {
        loadingOverlay.classList.add('hidden');
        saveButton.disabled = false;
    }

    // Show success message
    function showSuccess(message) {
        successMessage.textContent = message;
        successMessage.classList.remove('hidden');
        // Hide after 5 seconds
        setTimeout(() => {
            successMessage.classList.add('hidden');
        }, 5000);
    }

    // Show error message
    function showError(message) {
        errorMessage.textContent = message;
        errorMessage.classList.remove('hidden');
        // Hide after 10 seconds
        setTimeout(() => {
            errorMessage.classList.add('hidden');
        }, 10000);
    }

    // Hide all messages
    function hideMessages() {
        successMessage.classList.add('hidden');
        errorMessage.classList.add('hidden');
    }

    // Authentication functions (duplicated from app.js for standalone operation)
    function checkAuthentication() {
        const accessToken = localStorage.getItem('accessToken');
        const refreshToken = localStorage.getItem('refreshToken');
        
        if (!accessToken || !refreshToken) {
            return false;
        }
        
        // Check if access token is expired
        try {
            const tokenPayload = JSON.parse(atob(accessToken.split('.')[1]));
            const currentTime = Math.floor(Date.now() / 1000);
            
            if (tokenPayload.exp < currentTime) {
                // Token is expired, try to refresh
                return refreshAccessToken();
            }
            
            return true;
        } catch (error) {
            console.error('Error validating token:', error);
            clearAuthTokens();
            return false;
        }
    }

    async function refreshAccessToken() {
        const refreshToken = localStorage.getItem('refreshToken');
        
        if (!refreshToken) {
            return false;
        }
        
        try {
            // Initialize Cognito SDK if not already done
            const poolData = {
                UserPoolId: window.APP_CONFIG.COGNITO.userPoolId,
                ClientId: window.APP_CONFIG.COGNITO.clientId
            };
            
            const userPool = new AmazonCognitoIdentity.CognitoUserPool(poolData);
            const cognitoUser = userPool.getCurrentUser();
            
            if (!cognitoUser) {
                clearAuthTokens();
                return false;
            }
            
            return new Promise((resolve) => {
                cognitoUser.getSession((err, session) => {
                    if (err) {
                        console.error('Error refreshing token:', err);
                        clearAuthTokens();
                        resolve(false);
                        return;
                    }
                    
                    if (session.isValid()) {
                        // Update stored tokens
                        localStorage.setItem('accessToken', session.getAccessToken().getJwtToken());
                        localStorage.setItem('idToken', session.getIdToken().getJwtToken());
                        localStorage.setItem('refreshToken', session.getRefreshToken().getToken());
                        resolve(true);
                    } else {
                        clearAuthTokens();
                        resolve(false);
                    }
                });
            });
        } catch (error) {
            console.error('Error refreshing token:', error);
            clearAuthTokens();
            return false;
        }
    }

    function clearAuthTokens() {
        localStorage.removeItem('accessToken');
        localStorage.removeItem('idToken');
        localStorage.removeItem('refreshToken');
        localStorage.removeItem('userEmail');
    }
});