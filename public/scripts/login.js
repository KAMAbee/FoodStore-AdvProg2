document.addEventListener('DOMContentLoaded', function() {
    const loginForm = document.querySelector('.main__form');
    const loginMessage = document.getElementById('login-message');

    if (loginForm) {
        loginForm.addEventListener('submit', function(e) {
            e.preventDefault();
            
            const username = document.getElementById('username').value;
            const password = document.getElementById('password').value;
            
            if (!username || !password) {
                loginMessage.textContent = 'Fill all fields';
                return;
            }
            
            fetch('/api/users/login', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ username, password })
            })
            .then(response => {
                if (!response.ok) {
                    throw new Error('Invalid username or password');
                }
                return response.json();
            })
            .then(data => {
                localStorage.setItem('userId', data.id);
                localStorage.setItem('username', data.username);
                window.location.href = '/profile';
            })
            .catch(error => {
                loginMessage.textContent = error.message;
            });
        });
    }
});