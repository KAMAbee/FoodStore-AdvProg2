const state = {
    products: [],
    cart: [],
    currentPage: 1,
    perPage: 6,
    total: 0,
    filters: {
        name: '',
        minPrice: '',
        maxPrice: ''
    },
    userId: '',
    orders: []
};

const DOM = {
    productsList: document.getElementById('products-list'),
    pagination: document.getElementById('pagination'),
    cartItems: document.getElementById('cart-items'),
    cartEmpty: document.getElementById('cart-empty'),
    cartCount: document.getElementById('cart-count'),
    cartTotal: document.getElementById('cart-total'),
    cartSection: document.getElementById('cart-section'),
    cartIcon: document.getElementById('cart-icon'),
    checkoutButton: document.getElementById('checkout-button'),
    orderForm: document.getElementById('order-form'),
    userIdInput: document.getElementById('user-id'),
    nameFilter: document.getElementById('name-filter'),
    minPriceFilter: document.getElementById('min-price'),
    maxPriceFilter: document.getElementById('max-price'),
    applyFiltersButton: document.getElementById('apply-filters'),
    resetFiltersButton: document.getElementById('reset-filters'),
    userIdFilter: document.getElementById('user-id-filter'),
    viewOrdersButton: document.getElementById('view-orders'),
    ordersList: document.getElementById('orders-list')
};

document.addEventListener('DOMContentLoaded', () => {
    fetchProducts();

    DOM.applyFiltersButton.addEventListener('click', applyFilters);
    DOM.resetFiltersButton.addEventListener('click', resetFilters);
    DOM.orderForm.addEventListener('submit', placeOrder);
    DOM.viewOrdersButton.addEventListener('click', fetchOrders);
    DOM.userIdInput.addEventListener('input', updateCheckoutButton);

    const savedCart = localStorage.getItem('cart');
    if (savedCart) {
        state.cart = JSON.parse(savedCart);
        updateCart();
    }

    const savedUserId = localStorage.getItem('userId');
    if (savedUserId) {
        DOM.userIdInput.value = savedUserId;
        DOM.userIdFilter.value = savedUserId;
        state.userId = savedUserId;
        updateCheckoutButton();
        fetchOrders();
    }
});

async function fetchProducts() {
    try {
        const url = new URL('/api/products', "http://localhost:8082");
        url.searchParams.append('page', state.currentPage);
        url.searchParams.append('per_page', state.perPage);

        if (state.filters.name) {
            url.searchParams.append('name', state.filters.name);
        }
        if (state.filters.minPrice) {
            url.searchParams.append('min_price', state.filters.minPrice);
        }
        if (state.filters.maxPrice) {
            url.searchParams.append('max_price', state.filters.maxPrice);
        }

        const response = await fetch(url);
        if (!response.ok) {
            throw new Error(`Error fetching products: ${response.statusText}`);
        }

        const data = await response.json();
        state.products = data.products || [];
        state.total = data.total || 0;
        
        renderProducts();
        renderPagination();
    } catch (error) {
        console.error('Error fetching products:', error);
        DOM.productsList.innerHTML = `<div class="main__products-empty">Error loading products: ${error.message}</div>`;
    }
}

function renderProducts() {
    DOM.productsList.innerHTML = '';

    if (!state.products || state.products.length === 0) {
        DOM.productsList.innerHTML = '<div class="main__products-empty">Products not found</div>';
        return;
    }

    state.products.forEach(product => {
        const id = product.ID || product.id;
        const name = product.Name || product.name;
        const price = product.Price !== undefined ? product.Price : 
                      product.price !== undefined ? product.price : 0;
        const stock = product.Stock !== undefined ? product.Stock : 
                      product.stock !== undefined ? product.stock : 0;

        const productElement = document.createElement('div');
        productElement.className = 'main__product-item';
        
        const isInCart = state.cart.some(item => item.id === id);
        const isOutOfStock = stock <= 0;
        
        productElement.innerHTML = `
            <div class="main__product-item-name">${name}</div>
            <div class="main__product-item-price">${parseFloat(price).toFixed(2)} ₸</div>
            <div class="main__product-item-stock">Stock: ${stock}</div>
            <button class="main__product-item-button" 
                    data-id="${id}" 
                    data-name="${name}" 
                    data-price="${price}" 
                    data-stock="${stock}"
                    ${isInCart || isOutOfStock ? 'disabled' : ''}
            >
                ${isInCart ? 'In cart' : isOutOfStock ? 'Out of stock' : 'Add to cart'}
            </button>
        `;
        
        const addButton = productElement.querySelector('button');
        if (!isInCart && !isOutOfStock) {
            addButton.addEventListener('click', addToCart);
        }
        
        DOM.productsList.appendChild(productElement);
    });
}

function renderPagination() {
    DOM.pagination.innerHTML = '';
    
    const totalPages = Math.ceil(state.total / state.perPage);
    if (totalPages <= 1) return;
    
    if (state.currentPage > 1) {
        const prevButton = document.createElement('button');
        prevButton.textContent = 'prev';
        prevButton.addEventListener('click', () => {
            state.currentPage--;
            fetchProducts();
        });
        DOM.pagination.appendChild(prevButton);
    }
    
    for (let i = 1; i <= totalPages; i++) {
        const pageButton = document.createElement('button');
        pageButton.textContent = i;
        pageButton.disabled = i === state.currentPage;
        pageButton.addEventListener('click', () => {
            state.currentPage = i;
            fetchProducts();
        });
        DOM.pagination.appendChild(pageButton);
    }
    
    if (state.currentPage < totalPages) {
        const nextButton = document.createElement('button');
        nextButton.textContent = 'next';
        nextButton.addEventListener('click', () => {
            state.currentPage++;
            fetchProducts();
        });
        DOM.pagination.appendChild(nextButton);
    }
}

function addToCart(event) {
    const button = event.target;
    const id = button.dataset.id;
    const name = button.dataset.name;
    const price = parseFloat(button.dataset.price);
    const stock = parseInt(button.dataset.stock);
    
    const existingItem = state.cart.find(item => item.id === id);
    
    if (existingItem) {
        existingItem.quantity += 1;
    } else {
        state.cart.push({
            id,
            name,
            price,
            stock,
            quantity: 1
        });
    }
    
    button.disabled = true;
    button.textContent = 'In cart';
    
    localStorage.setItem('cart', JSON.stringify(state.cart));
    
    updateCart();
}

function updateCart() {
    const totalItems = state.cart.reduce((sum, item) => sum + item.quantity, 0);
    DOM.cartCount.textContent = totalItems;
    
    if (totalItems === 0) {
        DOM.cartEmpty.style.display = 'block';
        DOM.cartItems.innerHTML = '';
        DOM.cartTotal.textContent = '0 ₸';
        DOM.checkoutButton.disabled = true;
        return;
    }
    
    DOM.cartEmpty.style.display = 'none';
    
    DOM.cartItems.innerHTML = '';
    
    let totalPrice = 0;
    
    state.cart.forEach(item => {
        const itemTotal = item.price * item.quantity;
        totalPrice += itemTotal;
        
        const cartItem = document.createElement('div');
        cartItem.className = 'main__cart-item';
        cartItem.innerHTML = `
            <div class="main__cart-item-details">
                <div class="main__cart-item-name">${item.name}</div>
                <div class="main__cart-item-price">${item.price.toFixed(2)} ₸ x ${item.quantity}</div>
            </div>
            <div class="main__cart-item-quantity">
                <button class="decrease" data-id="${item.id}" ${item.quantity <= 1 ? 'disabled' : ''}>-</button>
                <span>${item.quantity}</span>
                <button class="increase" data-id="${item.id}" ${item.quantity >= item.stock ? 'disabled' : ''}>+</button>
            </div>
            <button class="main__cart-item-remove" data-id="${item.id}">×</button>
        `;
        
        const decreaseButton = cartItem.querySelector('.decrease');
        const increaseButton = cartItem.querySelector('.increase');
        const removeButton = cartItem.querySelector('.main__cart-item-remove');
        
        decreaseButton.addEventListener('click', () => decreaseQuantity(item.id));
        increaseButton.addEventListener('click', () => increaseQuantity(item.id));
        removeButton.addEventListener('click', () => removeFromCart(item.id));
        
        DOM.cartItems.appendChild(cartItem);
    });
    
    DOM.cartTotal.textContent = `${totalPrice.toFixed(2)} ₸`;
    
    updateCheckoutButton();
}

function decreaseQuantity(id) {
    const item = state.cart.find(item => item.id === id);
    if (!item) return;
    
    item.quantity -= 1;
    
    if (item.quantity <= 0) {
        removeFromCart(id);
    } else {
        localStorage.setItem('cart', JSON.stringify(state.cart));
        updateCart();
    }
}

function increaseQuantity(id) {
    const item = state.cart.find(item => item.id === id);
    if (!item || item.quantity >= item.stock) return;
    
    item.quantity += 1;
    localStorage.setItem('cart', JSON.stringify(state.cart));
    updateCart();
}

function removeFromCart(id) {
    state.cart = state.cart.filter(item => item.id !== id);
    localStorage.setItem('cart', JSON.stringify(state.cart));
    
    const productButton = document.querySelector(`button[data-id="${id}"]`);
    if (productButton) {
        productButton.disabled = false;
        productButton.textContent = 'Add to cart';
    }
    
    updateCart();
}

function applyFilters() {
    state.filters.name = DOM.nameFilter.value;
    state.filters.minPrice = DOM.minPriceFilter.value;
    state.filters.maxPrice = DOM.maxPriceFilter.value;
    state.currentPage = 1;
    fetchProducts();
}

function resetFilters() {
    DOM.nameFilter.value = '';
    DOM.minPriceFilter.value = '';
    DOM.maxPriceFilter.value = '';
    state.filters = { name: '', minPrice: '', maxPrice: '' };
    state.currentPage = 1;
    fetchProducts();
}

function updateCheckoutButton() {
    const userId = DOM.userIdInput.value.trim();
    state.userId = userId;
    
    if (userId && state.cart.length > 0) {
        DOM.checkoutButton.disabled = false;
    } else {
        DOM.checkoutButton.disabled = true;
    }
    
    if (userId) {
        localStorage.setItem('userId', userId);
    }
}

async function placeOrder(event) {
    event.preventDefault();
    
    const userId = DOM.userIdInput.value.trim();
    if (!userId || state.cart.length === 0) return;
    
    try {
        const orderData = {
            user_id: userId,
            items: state.cart.map(item => ({
                product_id: item.id,
                quantity: item.quantity
            }))
        };
        
        console.log("Sending order data:", orderData);
        
        const response = await fetch('http://localhost:8093/api/orders', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Accept': 'application/json',
                'Origin': window.location.origin
            },
            body: JSON.stringify(orderData)
        });
        
        if (!response.ok) {
            const errorText = await response.text();
            throw new Error(`Error creating order: ${errorText}`);
        }
        
        const result = await response.json();
        console.log("Order created:", result);
        
        state.cart = [];
        localStorage.setItem('cart', JSON.stringify(state.cart));
        updateCart();
        
        fetchProducts();
        fetchOrders();
        
        alert(`Order created successfully! Order ID: ${result.order_id}`);
        
    } catch (error) {
        console.error('Error creating order:', error);
        alert(`Failed to create order: ${error.message}`);
    }
}

async function fetchOrders() {
    const userId = DOM.userIdFilter.value.trim();
    if (!userId) {
        DOM.ordersList.innerHTML = '<div class="main__orders-empty">Please enter user ID</div>';
        return;
    }
    
    const token = localStorage.getItem('token');
    if (!token) {
        window.location.href = '/login';
        return;
    }
    
    try {
        const url = new URL('/api/orders', 'http://localhost:8093');
        url.searchParams.append('user_id', userId);
        
        console.log("Fetching orders from:", url.toString());
        
        const response = await fetch(url, {
            headers: {
                'Accept': 'application/json',
                'Origin': window.location.origin,
                'Authorization': `Bearer ${token}`
            }
        });
        
        if (response.status === 401 || response.status === 403) {
            localStorage.removeItem('token');
            localStorage.removeItem('userId');
            localStorage.removeItem('username');
            window.location.href = '/login';
            return;
        }
        
        if (!response.ok) {
            throw new Error(`Error fetching orders: ${response.statusText}`);
        }
        
        const data = await response.json();
        console.log("Orders received:", data);
        
        state.orders = data.orders || [];
        
        renderOrders();
    } catch (error) {
        console.error('Error fetching orders:', error);
        DOM.ordersList.innerHTML = `<div class="main__orders-error">Error fetching orders: ${error.message}</div>`;
    }
}

function renderOrders() {
    DOM.ordersList.innerHTML = '';
    
    if (!state.orders || state.orders.length === 0) {
        DOM.ordersList.innerHTML = '<div class="main__orders-empty">No orders found</div>';
        return;
    }
    
    state.orders.forEach(order => {
        const orderElement = document.createElement('div');
        orderElement.className = 'main__order-item';
        
        let createdDate = 'Date unavailable';
        try {
            if (order.created_at) {
                createdDate = new Date(order.created_at).toLocaleString();
            } else if (order.CreatedAt) {
                createdDate = new Date(order.CreatedAt).toLocaleString();
            }
        } catch (e) {
            console.error('Error parsing date:', e);
        }
        
        const id = order.id || order.ID;
        const status = order.status || order.Status || 'unknown';
        const totalPrice = order.total_price || order.TotalPrice || 0;
        const items = order.items || order.Items || [];
        
        orderElement.innerHTML = `
            <div class="main__order-item-header">
                <div class="main__order-item-date">${createdDate}</div>
                <div class="main__order-item-id">ID: ${id}</div>
                <div class="main__order-item-status ${status}">${getStatusText(status)}</div>
            </div>
            ${renderOrderItems(items)}
            <div class="main__order-item-total">
                <span>Total:</span>
                <span>${parseFloat(totalPrice).toFixed(2)} ₸</span>
            </div>
            ${renderOrderActions(order)}
        `;
        
        DOM.ordersList.appendChild(orderElement);
        
        const actionButtons = orderElement.querySelectorAll('.main__order-item-button');
        actionButtons.forEach(button => {
            button.addEventListener('click', () => updateOrderStatus(id, button.dataset.status));
        });
    });
}

function renderOrderItems(items) {
    if (!items || items.length === 0) {
        return '<div class="main__order-item-products-empty">No items in this order</div>';
    }
    
    let html = '<div class="main__order-item-products">';
    
    items.forEach(item => {
        const product = item.product || item.Product;
        const productName = product ? (product.Name || product.name || 'Unknown product') : 'Product details unavailable';
        const quantity = item.quantity || item.Quantity || 0;
        const price = item.price || item.Price || 0;
        
        html += `
            <div class="main__order-item-product">
                <div class="main__order-item-product-name">${productName}</div>
                <div class="main__order-item-product-quantity">Quantity: ${quantity}</div>
                <div class="main__order-item-product-price">${parseFloat(price).toFixed(2)} ₸</div>
            </div>
        `;
    });
    
    html += '</div>';
    return html;
}

function renderOrderActions(order) {
    const status = order.status || order.Status || 'unknown';
    
    if (status === 'completed' || status === 'cancelled') {
        return '';
    }
    
    return `
        <div class="main__order-item-actions">
            ${status === 'pending' ? `
                <button class="main__order-item-button complete" data-status="completed">Complete</button>
                <button class="main__order-item-button cancel" data-status="cancelled">Cancel</button>
            ` : ''}
        </div>
    `;
}

async function updateOrderStatus(orderId, status) {
    try {
        console.log(`Updating order ${orderId} status to ${status}`);
        
        const response = await fetch(`http://localhost:8093/api/orders/${orderId}`, {
            method: 'PATCH',
            headers: {
                'Content-Type': 'application/json',
                'Accept': 'application/json',
                'Origin': window.location.origin
            },
            body: JSON.stringify({ status })
        });
        
        if (!response.ok) {
            const errorText = await response.text();
            throw new Error(`Error updating order: ${errorText}`);
        }
        
        console.log("Order status updated successfully");
        
        fetchOrders();
        
        if (status === 'cancelled') {
            fetchProducts();
        }
        
        alert(`Order status updated to ${getStatusText(status)}`);
    } catch (error) {
        console.error('Error updating order status:', error);
        alert(`Failed to update order: ${error.message}`);
    }
}

function getStatusText(status) {
    switch (status) {
        case 'pending':
            return 'Pending';
        case 'completed':
            return 'Completed';
        case 'cancelled':
            return 'Cancelled';
        default:
            return status;
    }
}