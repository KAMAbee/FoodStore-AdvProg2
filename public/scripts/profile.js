document.addEventListener('DOMContentLoaded', function() {
    const userId = localStorage.getItem('userId');
    const username = localStorage.getItem('username');
        const profileContent = document.getElementById('profile-content');
    const usernameDisplay = document.getElementById('username-display');
    const userIdDisplay = document.getElementById('user-id-display');    
    const logoutButton = document.getElementById('logout-button');

    if (!userId || !username) {
        window.location.href = '/login';
        return;
    }

    if (usernameDisplay) usernameDisplay.textContent = username;
    if (userIdDisplay) userIdDisplay.textContent = userId;

    fetch(`/api/users/${userId}`)
        .then(response => {
            if (!response.ok) {
                throw new Error('Failed to load profile data');
            }
            return response.json();
        })
        .then(data => {
            console.log('Profile data loaded:', data);
            
            if (loadingElement) loadingElement.style.display = 'none';
            if (profileContent) profileContent.style.display = 'block';
            
            if (usernameDisplay) usernameDisplay.textContent = data.username;
            if (userIdDisplay) userIdDisplay.textContent = data.id;
        })
        .catch(error => {
            console.error('Profile loading error:', error);
            
            if (loadingElement) loadingElement.style.display = 'none';
            
            if (errorMessage) {
                errorMessage.textContent = error.message || 'Error loading profile';
                errorMessage.style.display = 'block';
            }
            
            if (profileContent) profileContent.style.display = 'block';
        });

    if (logoutButton) {
        logoutButton.addEventListener('click', function() {
            localStorage.removeItem('userId');
            localStorage.removeItem('username');
            
            window.location.href = '/login';
        });
    }
});