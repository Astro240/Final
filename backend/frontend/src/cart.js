// Create notification container if it doesn't exist
function createNotificationContainer() {
    if (!document.getElementById('notificationContainer')) {
        const container = document.createElement('div');
        container.id = 'notificationContainer';
        container.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            z-index: 10000;
            display: flex;
            flex-direction: column;
            gap: 10px;
            max-width: 400px;
        `;
        document.body.appendChild(container);
    }
}

// Show notification toast
function showNotification(message, type = 'info') {
    createNotificationContainer();
    
    const notification = document.createElement('div');
    notification.style.cssText = `
        padding: 16px 20px;
        border-radius: 8px;
        box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
        color: white;
        font-weight: 500;
        animation: slideIn 0.3s ease-out;
        backdrop-filter: blur(10px);
        display: flex;
        align-items: center;
        gap: 10px;
    `;
    
    const colors = {
        success: 'linear-gradient(135deg, #48bb78, #38a169)',
        error: 'linear-gradient(135deg, #f56565, #e53e3e)',
        warning: 'linear-gradient(135deg, #ed8936, #dd6b20)',
        info: 'linear-gradient(135deg, #4299e1, #3182ce)'
    };
    
    const icons = {
        success: '✓',
        error: '✕',
        warning: '⚠',
        info: 'ℹ'
    };
    
    notification.style.background = colors[type] || colors.info;
    notification.innerHTML = `<span style="font-size: 1.2em;">${icons[type] || icons.info}</span> ${message}`;
    
    const style = document.createElement('style');
    style.textContent = `
        @keyframes slideIn {
            from {
                transform: translateX(400px);
                opacity: 0;
            }
            to {
                transform: translateX(0);
                opacity: 1;
            }
        }
        @keyframes slideOut {
            from {
                transform: translateX(0);
                opacity: 1;
            }
            to {
                transform: translateX(400px);
                opacity: 0;
            }
        }
    `;
    
    if (!document.getElementById('notificationStyles')) {
        style.id = 'notificationStyles';
        document.head.appendChild(style);
    }
    
    document.getElementById('notificationContainer').appendChild(notification);
    
    // Auto remove after 4 seconds
    setTimeout(() => {
        notification.style.animation = 'slideOut 0.3s ease-in';
        setTimeout(() => notification.remove(), 300);
    }, 4000);
}

// Function to update cart total display
async function updateCartTotal() {
    try {
        const fullPath = window.location.href;
        const response = await fetch('/api/get-cart', {
            method: 'GET',
            headers: {
                'X-Store-Name': fullPath
            }
        });
        
        if (response.ok) {
            const data = await response.json();
            const cartTotalElements = document.querySelectorAll('#cart-total');
            cartTotalElements.forEach(element => {
                element.textContent = data.total_items || 0;
            });
        }
    } catch (error) {
        console.error('Failed to update cart total:', error);
    }
}

// Handle Add to Cart functionality
document.addEventListener('DOMContentLoaded', () => {
    // Initialize cart total on page load
    updateCartTotal();
    
    const addToCartButtons = document.querySelectorAll('.add-to-cart');
    
    addToCartButtons.forEach(button => {
        button.addEventListener('click', async (e) => {
            e.preventDefault();
            
            const productId = button.dataset.productId;
            const originalText = button.textContent;
            
            button.disabled = true;
            button.textContent = 'Adding...';
            
            const formData = new FormData();
            formData.append('product_id', productId);
            formData.append('quantity', '1');
            
            try {
                const fullPath = window.location.href;
                const response = await fetch('/api/add-to-cart', {
                    method: 'POST',
                    body: formData,
                    headers: {
                        'X-Store-Name': fullPath
                    }
                });
                
                const data = await response.json();
                
                if (response.ok) {
                    button.textContent = '✓ Added!';
                    button.style.background = '#48bb78';
                    showNotification('Item added to cart successfully!', 'success');
                    
                    // Update cart total
                    updateCartTotal();
                    
                    setTimeout(() => {
                        button.textContent = originalText;
                        button.style.background = '';
                        button.disabled = false;
                    }, 2000);
                } else {
                    if (response.status === 401) {
                        showNotification('Please login to add items to cart', 'warning');
                        button.textContent = originalText;
                        button.style.background = '';
                        button.disabled = false;
                        openModal();
                    } else {
                        showNotification(data.error || 'Failed to add to cart', 'error');
                        button.textContent = originalText;
                        button.disabled = false;
                    }
                }
            } catch (error) {
                showNotification('An error occurred. Please try again.', 'error');
                button.textContent = originalText;
                button.disabled = false;
            }
        });
    });
});

function goToCheckout(){
    window.location.href = window.location.href+"/checkout";
}

function goToPage(text) {
    window.location.href = window.location.href+text;
}

// Standalone addToCart function for programmatic calls
async function addToCart(productId, quantity = 1) {
    const formData = new FormData();
    formData.append('product_id', productId);
    formData.append('quantity', quantity.toString());
    
    try {
        const fullPath = window.location.href;
        const response = await fetch('/api/add-to-cart', {
            method: 'POST',
            body: formData,
            headers: {
                'X-Store-Name': fullPath
            }
        });
        
        const data = await response.json();
        
        if (response.ok) {
            showNotification(`${quantity} item${quantity > 1 ? 's' : ''} added to cart successfully!`, 'success');
            
            // Update cart total
            updateCartTotal();
            
            return true;
        } else {
            if (response.status === 401) {
                showNotification('Please login to add items to cart', 'warning');
                if (typeof openModal === 'function') {
                    openModal();
                }
            } else {
                showNotification(data.error || 'Failed to add to cart', 'error');
            }
            return false;
        }
    } catch (error) {
        showNotification('An error occurred. Please try again.', 'error');
        return false;
    }
}