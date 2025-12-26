// Review Management Script

// Fetch reviewable orders
async function fetchReviewableOrders() {
    try {
        const response = await fetch('/api/reviews/reviewable-orders');
        if (!response.ok) {
            throw new Error('Failed to fetch reviewable orders');
        }
        const data = await response.json();
        return data.orders || [];
    } catch (error) {
        console.error('Error fetching reviewable orders:', error);
        return [];
    }
}

// Display reviewable orders
async function displayReviewableOrders(containerId = 'reviewable-orders-container') {
    const container = document.getElementById(containerId);
    if (!container) return;

    container.innerHTML = '<div class="loading">Loading orders...</div>';

    const orders = await fetchReviewableOrders();

    if (orders.length === 0) {
        container.innerHTML = '<div class="no-orders">No orders available for review.</div>';
        return;
    }

    let html = '';
    orders.forEach(order => {
        html += `
            <div class="reviewable-order">
                <div class="order-header">
                    <h3>Order #${order.order_id}</h3>
                    <span class="order-status">${order.status}</span>
                </div>
                <div class="order-info">
                    <span>Total: BD ${order.total_amount.toFixed(2)}</span>
                    <span>Date: ${new Date(order.created_at).toLocaleDateString()}</span>
                </div>
                <div class="order-products">
                    ${order.products.map(product => `
                        <div class="reviewable-product">
                            <img src="/products_image/${product.product_image}" alt="${product.product_name}" class="product-thumb">
                            <div class="product-info">
                                <h4>${product.product_name}</h4>
                                ${product.has_reviewed ? 
                                    '<span class="reviewed-badge">✓ Reviewed</span>' : 
                                    `<button class="btn-review" onclick="openReviewModal(${order.order_id}, ${product.product_id}, '${product.product_name}')">Write Review</button>`
                                }
                            </div>
                        </div>
                    `).join('')}
                </div>
            </div>
        `;
    });

    container.innerHTML = html;
}

// Open review modal
function openReviewModal(orderId, productId, productName) {
    const modalHTML = `
        <div id="reviewModal" class="review-modal-overlay" onclick="closeReviewModalOnBackground(event)">
            <div class="review-modal">
                <div class="review-modal-header">
                    <h3>Write a Review</h3>
                    <button class="review-modal-close" onclick="closeReviewModal()">×</button>
                </div>
                <div class="review-modal-body">
                    <h4>${productName}</h4>
                    
                    <div class="rating-input">
                        <label>Your Rating:</label>
                        <div class="star-rating-input" id="starRatingInput">
                            <span class="star-input" data-rating="1">☆</span>
                            <span class="star-input" data-rating="2">☆</span>
                            <span class="star-input" data-rating="3">☆</span>
                            <span class="star-input" data-rating="4">☆</span>
                            <span class="star-input" data-rating="5">☆</span>
                        </div>
                        <input type="hidden" id="ratingValue" value="0">
                    </div>

                    <div class="comment-input">
                        <label for="reviewComment">Your Review:</label>
                        <textarea id="reviewComment" rows="5" placeholder="Share your experience with this product..."></textarea>
                    </div>

                    <div class="review-modal-actions">
                        <button class="btn-cancel" onclick="closeReviewModal()">Cancel</button>
                        <button class="btn-submit" onclick="submitReview(${orderId}, ${productId})">Submit Review</button>
                    </div>
                </div>
            </div>
        </div>
    `;

    document.body.insertAdjacentHTML('beforeend', modalHTML);
    initializeStarRating();
    addReviewModalStyles();
}

// Initialize star rating input
function initializeStarRating() {
    const stars = document.querySelectorAll('.star-input');
    const ratingInput = document.getElementById('ratingValue');

    stars.forEach(star => {
        star.addEventListener('click', function() {
            const rating = this.getAttribute('data-rating');
            ratingInput.value = rating;
            updateStarDisplay(rating);
        });

        star.addEventListener('mouseenter', function() {
            const rating = this.getAttribute('data-rating');
            updateStarDisplay(rating);
        });
    });

    document.getElementById('starRatingInput').addEventListener('mouseleave', function() {
        const currentRating = ratingInput.value;
        updateStarDisplay(currentRating);
    });
}

// Update star display
function updateStarDisplay(rating) {
    const stars = document.querySelectorAll('.star-input');
    stars.forEach((star, index) => {
        if (index < rating) {
            star.textContent = '★';
            star.classList.add('active');
        } else {
            star.textContent = '☆';
            star.classList.remove('active');
        }
    });
}

// Submit review
async function submitReview(orderId, productId) {
    const rating = document.getElementById('ratingValue').value;
    const comment = document.getElementById('reviewComment').value.trim();

    if (rating === '0') {
        alert('Please select a rating');
        return;
    }

    const submitBtn = document.querySelector('.btn-submit');
    submitBtn.disabled = true;
    submitBtn.textContent = 'Submitting...';

    try {
        const formData = new FormData();
        formData.append('product_id', productId);
        formData.append('order_id', orderId);
        formData.append('rating', rating);
        formData.append('comment', comment);

        const response = await fetch('/api/reviews/create', {
            method: 'POST',
            body: formData
        });

        const data = await response.json();

        if (response.ok) {
            alert('Review submitted successfully!');
            closeReviewModal();
            // Refresh the reviewable orders list
            displayReviewableOrders();
        } else {
            alert(data.error || 'Failed to submit review');
            submitBtn.disabled = false;
            submitBtn.textContent = 'Submit Review';
        }
    } catch (error) {
        console.error('Error submitting review:', error);
        alert('An error occurred. Please try again.');
        submitBtn.disabled = false;
        submitBtn.textContent = 'Submit Review';
    }
}

// Close review modal
function closeReviewModal() {
    const modal = document.getElementById('reviewModal');
    if (modal) {
        modal.style.opacity = '0';
        setTimeout(() => modal.remove(), 300);
    }
}

// Close modal on background click
function closeReviewModalOnBackground(event) {
    if (event.target.id === 'reviewModal') {
        closeReviewModal();
    }
}

// Add review modal styles
function addReviewModalStyles() {
    if (document.getElementById('reviewModalStyles')) return;

    const styles = `
        <style id="reviewModalStyles">
            .review-modal-overlay {
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
                opacity: 1;
                transition: opacity 0.3s ease;
            }

            .review-modal {
                background: white;
                border-radius: 15px;
                width: 90%;
                max-width: 500px;
                box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
                animation: slideUp 0.3s ease;
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

            .review-modal-header {
                display: flex;
                justify-content: space-between;
                align-items: center;
                padding: 1.5rem;
                border-bottom: 2px solid #e0e7ff;
            }

            .review-modal-header h3 {
                margin: 0;
                color: #1a202c;
                font-size: 1.5rem;
            }

            .review-modal-close {
                background: none;
                border: none;
                font-size: 2rem;
                cursor: pointer;
                color: #666;
                width: 35px;
                height: 35px;
                display: flex;
                align-items: center;
                justify-content: center;
                border-radius: 50%;
                transition: all 0.3s;
            }

            .review-modal-close:hover {
                background: rgba(0, 0, 0, 0.1);
                transform: rotate(90deg);
            }

            .review-modal-body {
                padding: 1.5rem;
            }

            .review-modal-body h4 {
                margin: 0 0 1.5rem 0;
                color: #667eea;
                font-size: 1.2rem;
            }

            .rating-input {
                margin-bottom: 1.5rem;
            }

            .rating-input label {
                display: block;
                margin-bottom: 0.5rem;
                font-weight: 600;
                color: #1a202c;
            }

            .star-rating-input {
                display: flex;
                gap: 0.5rem;
                font-size: 2rem;
            }

            .star-input {
                cursor: pointer;
                color: #ddd;
                transition: all 0.2s;
            }

            .star-input:hover,
            .star-input.active {
                color: #ffc107;
                transform: scale(1.2);
            }

            .comment-input {
                margin-bottom: 1.5rem;
            }

            .comment-input label {
                display: block;
                margin-bottom: 0.5rem;
                font-weight: 600;
                color: #1a202c;
            }

            .comment-input textarea {
                width: 100%;
                padding: 0.75rem;
                border: 2px solid #e0e7ff;
                border-radius: 10px;
                font-family: inherit;
                font-size: 1rem;
                resize: vertical;
                transition: border-color 0.3s;
            }

            .comment-input textarea:focus {
                outline: none;
                border-color: #667eea;
            }

            .review-modal-actions {
                display: flex;
                gap: 1rem;
                justify-content: flex-end;
            }

            .btn-cancel {
                padding: 0.75rem 1.5rem;
                border: 2px solid #e0e7ff;
                background: white;
                color: #666;
                border-radius: 10px;
                font-weight: 600;
                cursor: pointer;
                transition: all 0.3s;
            }

            .btn-cancel:hover {
                background: #f8f9fa;
                border-color: #667eea;
            }

            .btn-submit {
                padding: 0.75rem 1.5rem;
                border: none;
                background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
                color: white;
                border-radius: 10px;
                font-weight: 600;
                cursor: pointer;
                transition: all 0.3s;
            }

            .btn-submit:hover:not(:disabled) {
                transform: translateY(-2px);
                box-shadow: 0 5px 15px rgba(102, 126, 234, 0.4);
            }

            .btn-submit:disabled {
                opacity: 0.6;
                cursor: not-allowed;
            }

            /* Reviewable orders styles */
            .reviewable-order {
                background: white;
                border: 2px solid #e0e7ff;
                border-radius: 12px;
                padding: 1.5rem;
                margin-bottom: 1.5rem;
            }

            .order-header {
                display: flex;
                justify-content: space-between;
                align-items: center;
                margin-bottom: 1rem;
            }

            .order-header h3 {
                margin: 0;
                color: #1a202c;
            }

            .order-status {
                padding: 0.5rem 1rem;
                background: rgba(102, 126, 234, 0.1);
                color: #667eea;
                border-radius: 8px;
                font-weight: 600;
                font-size: 0.9rem;
                text-transform: capitalize;
            }

            .order-info {
                display: flex;
                gap: 2rem;
                margin-bottom: 1rem;
                color: #666;
                font-size: 0.9rem;
            }

            .order-products {
                display: flex;
                flex-direction: column;
                gap: 1rem;
            }

            .reviewable-product {
                display: flex;
                align-items: center;
                gap: 1rem;
                padding: 1rem;
                background: #f8f9fa;
                border-radius: 10px;
            }

            .product-thumb {
                width: 60px;
                height: 60px;
                object-fit: cover;
                border-radius: 8px;
            }

            .product-info {
                flex: 1;
                display: flex;
                justify-content: space-between;
                align-items: center;
            }

            .product-info h4 {
                margin: 0;
                color: #1a202c;
                font-size: 1rem;
            }

            .btn-review {
                padding: 0.5rem 1rem;
                background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
                color: white;
                border: none;
                border-radius: 8px;
                font-weight: 600;
                cursor: pointer;
                transition: all 0.3s;
            }

            .btn-review:hover {
                transform: translateY(-2px);
                box-shadow: 0 5px 15px rgba(102, 126, 234, 0.4);
            }

            .reviewed-badge {
                padding: 0.5rem 1rem;
                background: rgba(76, 175, 80, 0.15);
                color: #2e7d32;
                border-radius: 8px;
                font-weight: 600;
                font-size: 0.9rem;
            }

            .loading, .no-orders {
                text-align: center;
                padding: 2rem;
                color: #666;
                font-style: italic;
            }

            @media (max-width: 768px) {
                .review-modal {
                    width: 95%;
                }

                .order-header {
                    flex-direction: column;
                    align-items: flex-start;
                    gap: 0.5rem;
                }

                .order-info {
                    flex-direction: column;
                    gap: 0.5rem;
                }

                .reviewable-product {
                    flex-direction: column;
                    text-align: center;
                }

                .product-info {
                    flex-direction: column;
                    gap: 0.75rem;
                    width: 100%;
                }
            }
        </style>
    `;

    document.head.insertAdjacentHTML('beforeend', styles);
}

// Auto-initialize if on orders page
document.addEventListener('DOMContentLoaded', function() {
    const reviewContainer = document.getElementById('reviewable-orders-container');
    if (reviewContainer) {
        displayReviewableOrders();
    }
});
