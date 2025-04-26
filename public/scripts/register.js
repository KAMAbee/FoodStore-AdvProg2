document.addEventListener('DOMContentLoaded', function() {
    const registerForm = document.getElementById('register-form');
    const registerMessage = document.getElementById('register-message');

    const token = localStorage.getItem('token');
    if (token) {
        window.location.href = '/profile';
        return;
    }

    if (registerForm) {
        registerForm.addEventListener('submit', function(e) {
            e.preventDefault();
            console.log("Form submitted");
            
            const username = document.getElementById('username').value;
            const password = document.getElementById('password').value;
            const confirmPassword = document.getElementById('confirm-password').value;
            
            if (registerMessage) {
                registerMessage.textContent = '';
                registerMessage.style.display = 'none';
            }
            
            if (!username || !password || !confirmPassword) {
                if (registerMessage) {
                    registerMessage.textContent = 'Fill all fields';
                    registerMessage.style.display = 'block';
                }
                return;
            }
            
            if (password !== confirmPassword) {
                if (registerMessage) {
                    registerMessage.textContent = 'Passwords are not same';
                    registerMessage.style.display = 'block';
                }
                return;
            }
            
            if (password.length < 6) {
                if (registerMessage) {
                    registerMessage.textContent = 'Password must be at least 6 characters long';
                    registerMessage.style.display = 'block';
                }
                return;
            }
                        
            fetch('/api/users/register', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ username, password })
            })
            .then(response => {
                if (!response.ok) {
                    return response.json().then(errorData => {
                        console.error("Server error details:", errorData);
                        
                        if (response.status === 409) {
                            throw new Error(errorData.error || 'Username already exists');
                        } else {
                            throw new Error(errorData.error || 'Registration failed');
                        }
                    }).catch(jsonError => {
                        if (response.status === 409) {
                            throw new Error('Username already exists');
                        } else {
                            throw new Error(`Registration failed (${response.status})`);
                        }
                    });
                }
                
                return response.json();
            })
            .then(data => {
                console.log("Registration successful:", data);
                
                localStorage.setItem('userId', data.user.id);
                localStorage.setItem('username', data.user.username);
                localStorage.setItem('token', data.token);
                
                window.location.href = '/profile';
            })
            .catch(error => {
                console.error("Registration error:", error);
                
                if (registerMessage) {
                    registerMessage.textContent = error.message || 'Registration failed';
                    registerMessage.style.color = 'red';
                    registerMessage.style.display = 'block';
                }
            });
        });
    } else {
        console.error("Register form not found in the document");
    }
});