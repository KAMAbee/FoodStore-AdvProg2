if (typeof axios !== 'undefined') {
    axios.interceptors.request.use(function (config) {
        const token = localStorage.getItem('token');
        if (token) {
            config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
    });
}

window.authenticatedFetch = function(url, options = {}) {
    const token = localStorage.getItem('token');
    
    const fetchOptions = {
        ...options,
        headers: {
            ...(options.headers || {}),
            'Authorization': `Bearer ${token}`
        }
    };
    
    return fetch(url, fetchOptions);
};

function isAuthenticated() {
    return !!localStorage.getItem('token');
}