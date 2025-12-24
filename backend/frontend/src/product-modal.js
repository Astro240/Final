// Product Detail Modal Script
// This script adds a beautiful modal for product details similar to AliExpress/Amazon

(function() {
    // Get product ID from header if it exists (set by backend)
    const productId = document.querySelector('meta[name="product-id"]')?.content;
    const showModal = document.querySelector('meta[name="show-product-modal"]')?.content === 'true';
    
    if (productId && showModal) {
        // Get product data from global scope if available
        if (window.productDetailData) {
            showProductModal(window.productDetailData);
        }
    }
})();

function showProductModal(product) {
    // Create modal HTML
    const modalHTML = `
        <div id="productDetailModal" class="product-modal-overlay">
            <div class="product-modal-container">
                <button class="product-modal-close" onclick="closeProductModal()">×</button>
                
                <div class="product-modal-content">
                    <!-- Image Section -->
                    <div class="product-modal-image-section">
                        <div class="product-modal-image">
                            ${product.Image ? `<img src="/products/${product.Image}" alt="${product.Name}">` : '<div style="font-size: 5rem; display: flex; align-items: center; justify-content: center; height: 100%;">📦</div>'}
                        </div>
                        <div class="product-modal-image-controls">
                            <button class="modal-image-btn" onclick="zoomProductImage()">🔍 Zoom</button>
                            <button class="modal-image-btn" onclick="shareProduct()">📤 Share</button>
                        </div>
                    </div>

                    <!-- Info Section -->
                    <div class="product-modal-info-section">
                        <!-- Title -->
                        <h2 class="product-modal-title">${product.Name}</h2>

                        <!-- Rating -->
                        <div class="product-modal-rating">
                            <span class="stars">★★★★★</span>
                            <span class="rating-text">4.8 (2,345 reviews)</span>
                        </div>

                        <!-- Price Section -->
                        <div class="product-modal-price-section">
                            <div class="product-modal-price">
                                <span class="currency">BD</span>
                                <span class="amount">${product.Price.toFixed(2)}</span>
                            </div>
                            <div class="price-info">
                                ${product.Quantity > 0 ? '<span class="price-badge in-stock">✓ In Stock</span>' : '<span class="price-badge out-stock">Out of Stock</span>'}
                            </div>
                        </div>

                        <!-- Description -->
                        <div class="product-modal-description">
                            <h4>Description</h4>
                            <p>${product.Description || 'Premium quality product. Fast shipping available.'}</p>
                        </div>

                        <!-- Stock Info -->
                        <div class="product-modal-stock">
                            <div class="stock-row">
                                <span class="stock-label">Available Stock:</span>
                                <span class="stock-value">${product.Quantity} units</span>
                            </div>
                            <div class="stock-row">
                                <span class="stock-label">Shipping:</span>
                                <span class="stock-value">Free (Orders over BD 50)</span>
                            </div>
                        </div>

                        <!-- Quantity Selector -->
                        <div class="product-modal-quantity">
                            <span class="qty-label">Quantity:</span>
                            <div class="qty-selector">
                                <button class="qty-btn" onclick="decreaseModalQty()" ${product.Quantity <= 0 ? 'disabled' : ''}>−</button>
                                <input type="number" class="qty-input" id="modalProductQty" value="1" min="1" max="${product.Quantity}" onchange="validateModalQty()" ${product.Quantity <= 0 ? 'disabled' : ''}>
                                <button class="qty-btn" onclick="increaseModalQty()" ${product.Quantity <= 0 ? 'disabled' : ''}>+</button>
                            </div>
                        </div>

                        <!-- Action Buttons -->
                        <div class="product-modal-actions">
                            <button class="btn-add-cart-modal" onclick="addToCartFromModal(${product.ID})" ${product.Quantity <= 0 ? 'disabled' : ''}>
                                <span>🛒</span> Add to Cart
                            </button>
                            <button class="btn-wishlist-modal" onclick="addToWishlistModal()">
                                <span>♡</span> Add to Wishlist
                            </button>
                        </div>

                        <!-- Trust Badges -->
                        <div class="product-modal-trust">
                            <div class="trust-item">
                                <span class="trust-icon">✓</span>
                                <span>Authentic Guarantee</span>
                            </div>
                            <div class="trust-item">
                                <span class="trust-icon">🔒</span>
                                <span>Secure Payment</span>
                            </div>
                            <div class="trust-item">
                                <span class="trust-icon">🚚</span>
                                <span>Fast Delivery</span>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    `;

    // Inject modal into DOM
    document.body.insertAdjacentHTML('beforeend', modalHTML);
    
    // Add styles
    addProductModalStyles();
    
    // Add animation
    setTimeout(() => {
        const modal = document.getElementById('productDetailModal');
        if (modal) {
            modal.classList.add('show');
        }
    }, 100);
}

function addProductModalStyles() {
    if (document.getElementById('productModalStyles')) return;
    
    const styles = `
        <style id="productModalStyles">
            .product-modal-overlay {
                position: fixed;
                top: 0;
                left: 0;
                width: 100%;
                height: 100%;
                background: rgba(0, 0, 0, 0.6);
                backdrop-filter: blur(5px);
                display: flex;
                align-items: center;
                justify-content: center;
                z-index: 10000;
                opacity: 0;
                transition: opacity 0.3s ease;
                animation: fadeIn 0.3s ease;
            }

            .product-modal-overlay.show {
                opacity: 1;
            }

            @keyframes fadeIn {
                from {
                    opacity: 0;
                }
                to {
                    opacity: 1;
                }
            }

            .product-modal-container {
                background: white;
                border-radius: 20px;
                width: 90%;
                max-width: 900px;
                max-height: 90vh;
                overflow-y: auto;
                box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
                position: relative;
                animation: slideUp 0.4s cubic-bezier(0.175, 0.885, 0.32, 1.275);
            }

            @keyframes slideUp {
                from {
                    transform: translateY(50px);
                    opacity: 0;
                }
                to {
                    transform: translateY(0);
                    opacity: 1;
                }
            }

            .product-modal-close {
                position: absolute;
                top: 15px;
                right: 15px;
                background: rgba(0, 0, 0, 0.1);
                border: none;
                width: 40px;
                height: 40px;
                border-radius: 50%;
                font-size: 2rem;
                cursor: pointer;
                transition: all 0.3s;
                z-index: 10001;
                display: flex;
                align-items: center;
                justify-content: center;
                color: #333;
            }

            .product-modal-close:hover {
                background: rgba(0, 0, 0, 0.2);
                transform: rotate(90deg);
            }

            .product-modal-content {
                display: grid;
                grid-template-columns: 1fr 1fr;
                gap: 2rem;
                padding: 2rem;
            }

            .product-modal-image-section {
                display: flex;
                flex-direction: column;
                gap: 1rem;
            }

            .product-modal-image {
                width: 100%;
                height: 400px;
                background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
                border-radius: 15px;
                display: flex;
                align-items: center;
                justify-content: center;
                overflow: hidden;
                box-shadow: 0 5px 20px rgba(0, 0, 0, 0.1);
            }

            .product-modal-image img {
                width: 100%;
                height: 100%;
                object-fit: cover;
            }

            .product-modal-image-controls {
                display: flex;
                gap: 0.75rem;
            }

            .modal-image-btn {
                flex: 1;
                padding: 0.75rem;
                border: 2px solid #e0e7ff;
                background: white;
                border-radius: 10px;
                cursor: pointer;
                font-weight: 600;
                transition: all 0.3s;
                color: #667eea;
            }

            .modal-image-btn:hover {
                border-color: #667eea;
                background: rgba(102, 126, 234, 0.05);
            }

            .product-modal-info-section {
                display: flex;
                flex-direction: column;
                gap: 1.5rem;
            }

            .product-modal-title {
                font-size: 1.8rem;
                font-weight: 700;
                color: #1a202c;
                line-height: 1.3;
            }

            .product-modal-rating {
                display: flex;
                align-items: center;
                gap: 0.75rem;
            }

            .stars {
                font-size: 1.2rem;
                color: #ffc107;
            }

            .rating-text {
                color: #666;
                font-size: 0.9rem;
            }

            .product-modal-price-section {
                display: flex;
                align-items: center;
                gap: 1.5rem;
                padding: 1rem;
                background: rgba(102, 126, 234, 0.05);
                border-radius: 12px;
            }

            .product-modal-price {
                display: flex;
                align-items: baseline;
                gap: 0.5rem;
            }

            .currency {
                font-size: 1rem;
                font-weight: 600;
                color: #667eea;
            }

            .amount {
                font-size: 2rem;
                font-weight: 800;
                color: #667eea;
            }

            .price-info {
                display: flex;
                gap: 0.5rem;
            }

            .price-badge {
                padding: 0.5rem 1rem;
                border-radius: 8px;
                font-weight: 600;
                font-size: 0.85rem;
            }

            .price-badge.in-stock {
                background: rgba(76, 175, 80, 0.15);
                color: #2e7d32;
            }

            .price-badge.out-stock {
                background: rgba(244, 67, 54, 0.15);
                color: #c62828;
            }

            .product-modal-description {
                display: flex;
                flex-direction: column;
                gap: 0.75rem;
                padding: 1rem;
                background: #f8f9fa;
                border-radius: 12px;
            }

            .product-modal-description h4 {
                color: #1a202c;
                font-weight: 700;
                margin: 0;
            }

            .product-modal-description p {
                color: #666;
                margin: 0;
                line-height: 1.6;
            }

            .product-modal-stock {
                display: flex;
                flex-direction: column;
                gap: 0.75rem;
                padding: 1rem;
                border: 2px solid #e0e7ff;
                border-radius: 12px;
            }

            .stock-row {
                display: flex;
                justify-content: space-between;
                align-items: center;
            }

            .stock-label {
                color: #666;
                font-weight: 600;
            }

            .stock-value {
                color: #1a202c;
                font-weight: 700;
                color: #667eea;
            }

            .product-modal-quantity {
                display: flex;
                align-items: center;
                gap: 1rem;
            }

            .qty-label {
                font-weight: 600;
                color: #1a202c;
            }

            .qty-selector {
                display: flex;
                align-items: center;
                gap: 0.75rem;
                background: #f8f9fa;
                padding: 0.5rem;
                border-radius: 10px;
            }

            .qty-btn {
                width: 36px;
                height: 36px;
                border: 2px solid #e0e7ff;
                background: white;
                border-radius: 8px;
                font-size: 1.1rem;
                cursor: pointer;
                font-weight: 700;
                transition: all 0.3s;
                color: #667eea;
            }

            .qty-btn:hover:not(:disabled) {
                border-color: #667eea;
                background: rgba(102, 126, 234, 0.1);
            }

            .qty-btn:disabled {
                opacity: 0.5;
                cursor: not-allowed;
            }

            .qty-input {
                width: 50px;
                text-align: center;
                border: 2px solid #e0e7ff;
                border-radius: 8px;
                padding: 0.5rem;
                font-weight: 700;
                font-size: 1rem;
            }

            .product-modal-actions {
                display: flex;
                gap: 1rem;
            }

            .btn-add-cart-modal {
                flex: 1;
                padding: 1rem;
                background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
                color: white;
                border: none;
                border-radius: 10px;
                font-weight: 700;
                font-size: 1rem;
                cursor: pointer;
                transition: all 0.3s;
                display: flex;
                align-items: center;
                justify-content: center;
                gap: 0.5rem;
                box-shadow: 0 5px 20px rgba(102, 126, 234, 0.3);
            }

            .btn-add-cart-modal:hover:not(:disabled) {
                transform: translateY(-3px);
                box-shadow: 0 8px 25px rgba(102, 126, 234, 0.4);
            }

            .btn-add-cart-modal:disabled {
                background: #ccc;
                cursor: not-allowed;
                box-shadow: none;
            }

            .btn-wishlist-modal {
                flex: 1;
                padding: 1rem;
                background: rgba(255, 64, 129, 0.1);
                color: #d1203f;
                border: 2px solid #d1203f;
                border-radius: 10px;
                font-weight: 700;
                cursor: pointer;
                transition: all 0.3s;
                display: flex;
                align-items: center;
                justify-content: center;
                gap: 0.5rem;
            }

            .btn-wishlist-modal:hover {
                background: #d1203f;
                color: white;
            }

            .product-modal-trust {
                display: grid;
                grid-template-columns: repeat(3, 1fr);
                gap: 1rem;
                padding: 1rem;
                background: #f8f9fa;
                border-radius: 12px;
            }

            .trust-item {
                display: flex;
                flex-direction: column;
                align-items: center;
                gap: 0.5rem;
                text-align: center;
                font-size: 0.9rem;
                color: #666;
            }

            .trust-icon {
                font-size: 1.5rem;
                color: #667eea;
            }

            /* Responsive */
            @media (max-width: 768px) {
                .product-modal-content {
                    grid-template-columns: 1fr;
                    gap: 1rem;
                    padding: 1rem;
                }

                .product-modal-image {
                    height: 300px;
                }

                .product-modal-title {
                    font-size: 1.4rem;
                }

                .amount {
                    font-size: 1.5rem;
                }

                .product-modal-trust {
                    grid-template-columns: 1fr;
                }

                .product-modal-actions {
                    flex-direction: column;
                }

                .product-modal-close {
                    width: 35px;
                    height: 35px;
                    font-size: 1.5rem;
                }
            }
        </style>
    `;
    
    document.head.insertAdjacentHTML('beforeend', styles);
}

function closeProductModal() {
    const modal = document.getElementById('productDetailModal');
    if (modal) {
        modal.style.opacity = '0';
        setTimeout(() => {
            modal.remove();
        }, 300);
    }
}

function increaseModalQty() {
    const qtyInput = document.getElementById('modalProductQty');
    const maxQty = parseInt(qtyInput.max);
    const currentQty = parseInt(qtyInput.value);
    if (currentQty < maxQty) {
        qtyInput.value = currentQty + 1;
    }
}

function decreaseModalQty() {
    const qtyInput = document.getElementById('modalProductQty');
    const currentQty = parseInt(qtyInput.value);
    if (currentQty > 1) {
        qtyInput.value = currentQty - 1;
    }
}

function validateModalQty() {
    const qtyInput = document.getElementById('modalProductQty');
    const maxQty = parseInt(qtyInput.max);
    let currentQty = parseInt(qtyInput.value);

    if (isNaN(currentQty) || currentQty < 1) {
        qtyInput.value = 1;
    } else if (currentQty > maxQty) {
        qtyInput.value = maxQty;
    }
}

async function addToCartFromModal(productId) {
    const quantity = parseInt(document.getElementById('modalProductQty').value);
    const btn = event.target.closest('.btn-add-cart-modal');
    const originalText = btn.innerHTML;

    btn.disabled = true;
    btn.innerHTML = '<span>⏳</span> Adding...';

    try {
        const formData = new FormData();
        formData.append('product_id', productId);
        formData.append('quantity', quantity);

        const response = await fetch('/api/add-to-cart', {
            method: 'POST',
            body: formData,
            headers: {
                'X-Store-Name': window.location.href
            }
        });

        const data = await response.json();

        if (response.ok) {
            btn.innerHTML = '<span>✓</span> Added!';
            btn.style.background = '#4caf50';
            setTimeout(() => {
                closeProductModal();
            }, 1500);
        } else {
            if (response.status === 401) {
                alert('Please login to add items to cart');
                window.location.href = '/login';
            } else {
                alert(data.error || 'Failed to add to cart');
                btn.innerHTML = originalText;
                btn.disabled = false;
            }
        }
    } catch (error) {
        console.error('Error:', error);
        alert('An error occurred. Please try again.');
        btn.innerHTML = originalText;
        btn.disabled = false;
    }
}

function addToWishlistModal() {
    const btn = event.target.closest('.btn-wishlist-modal');
    const icon = btn.querySelector('span');
    
    if (icon.textContent === '♡') {
        icon.textContent = '♥';
        btn.style.background = '#d1203f';
        btn.style.color = 'white';
    } else {
        icon.textContent = '♡';
        btn.style.background = 'rgba(255, 64, 129, 0.1)';
        btn.style.color = '#d1203f';
    }
}

function zoomProductImage() {
    alert('Zoom feature would open a full-size image gallery');
}

function shareProduct() {
    const url = window.location.href;
    if (navigator.share) {
        navigator.share({
            title: 'Check out this product',
            url: url
        });
    } else {
        alert('Share URL: ' + url);
    }
}

// Close modal on background click
document.addEventListener('click', function(event) {
    const modal = document.getElementById('productDetailModal');
    if (modal && event.target === modal) {
        closeProductModal();
    }
});

// Close modal on Escape key
document.addEventListener('keydown', function(event) {
    if (event.key === 'Escape') {
        closeProductModal();
    }
});
