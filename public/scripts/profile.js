document.addEventListener('DOMContentLoaded', function() {
    const userId = localStorage.getItem('userId');
    const username = localStorage.getItem('username');
    const token = localStorage.getItem('token');
    const userRole = localStorage.getItem('userRole') || 'user';
    const profileContent = document.getElementById('profile-content');
    const usernameDisplay = document.getElementById('username-display');
    const userIdDisplay = document.getElementById('user-id-display');
    const userRoleDisplay = document.getElementById('user-role-display');
    const logoutButton = document.getElementById('logout-button');
    const errorMessage = document.getElementById('error-message');
    const loadingElement = document.getElementById('loading');

    console.log("Profile.js executed, checking auth...");
    
    if (!token) {
        console.log("No token in localStorage, redirecting to login");
        window.location.href = '/login';
        return;
    }

    if (usernameDisplay) usernameDisplay.textContent = username || "Unknown";
    if (userIdDisplay) userIdDisplay.textContent = userId || "Unknown";
    if (userRoleDisplay) userRoleDisplay.textContent = userRole;

    if (userRole !== 'admin') {
        const adminLink = document.querySelector('a[href="/admin"]');
        if (adminLink) {
            adminLink.style.display = 'none';
        }
    }

    fetch(`/api/users/${userId}`, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}` 
        }
    })
    .then(response => {
        console.log("Profile API response status:", response.status);
        if (!response.ok) {
            if (response.status === 401 || response.status === 403) {
                console.log("Auth failed, clearing stored data");
                localStorage.removeItem('token');
                localStorage.removeItem('userId');
                localStorage.removeItem('username');
                localStorage.removeItem('userRole');
                document.cookie = "auth_token=; path=/; max-age=0";
                window.location.href = '/login';
                return null;
            }
            throw new Error('Failed to load profile data');
        }
        return response.json();
    })
    .then(data => {
        if (!data) return; 
        console.log('Profile data loaded successfully:', data);
        
        if (loadingElement) loadingElement.style.display = 'none';
        if (profileContent) profileContent.style.display = 'block';
        
        if (usernameDisplay) usernameDisplay.textContent = data.username;
        if (userIdDisplay) userIdDisplay.textContent = data.id;
        if (userRoleDisplay) userRoleDisplay.textContent = data.role;
        
        if (data.role) {
            localStorage.setItem('userRole', data.role);
        }
    })
    .catch(error => {

    });

    if (logoutButton) {
        logoutButton.addEventListener('click', function() {
            console.log("Logging out...");
            localStorage.removeItem('userId');
            localStorage.removeItem('username');
            localStorage.removeItem('token');
            localStorage.removeItem('userRole');
            document.cookie = "auth_token=; path=/; max-age=0";
            window.location.href = '/login';
        });
    }
});