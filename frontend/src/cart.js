// Handle Add to Cart functionality
document.addEventListener('DOMContentLoaded', () => {
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
                const response = await fetch('/api/add-to-cart', {
                    method: 'POST',
                    body: formData
                });
                
                const data = await response.json();
                
                if (response.ok) {
                    button.textContent = '✓ Added!';
                    button.style.background = '#48bb78';
                    
                    setTimeout(() => {
                        button.textContent = originalText;
                        button.style.background = '';
                        button.disabled = false;
                    }, 2000);
                } else {
                    if (response.status === 401) {
                        alert('Please login to add items to cart');
                        window.location.href = '/login';
                    } else {
                        alert(data.error || 'Failed to add to cart');
                        button.textContent = originalText;
                        button.disabled = false;
                    }
                }
            } catch (error) {
                alert('An error occurred. Please try again.');
                button.textContent = originalText;
                button.disabled = false;
            }
        });
    });
});

function goToCheckout(){
    window.location.href = window.location.href+"/checkout";
}