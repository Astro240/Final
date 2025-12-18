function createProduct(event) {
    event.preventDefault();
    const form = document.getElementById('addProductForm');

    const formData = new FormData(form);

    // Handle product image upload as base64
    const imageInput = document.querySelector('input[name="productImage"]');

    let imageProcessed = false;

    // Process product image file
    if (imageInput && imageInput.files[0]) {
        const imageFile = imageInput.files[0];
        const imageReader = new FileReader();
        imageReader.onloadend = function () {
            let base64data = imageReader.result;
            if (base64data === "data:,") {
                base64data = "";
            }
            formData.set('productImage', base64data);
            imageProcessed = true;
            checkAndSubmit();
        };
        imageReader.readAsDataURL(imageFile);
    } else {
        formData.set('productImage', '');
        imageProcessed = true;
        checkAndSubmit();
    }
    const formMessage = document.getElementById('formMessage');
    function checkAndSubmit() {
        if (imageProcessed) {
            fetch('/api/create_product', {
                method: 'POST',
                body: formData,
                credentials: 'include'
            })
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        formMessage.innerText = 'Product created successfully!';
                        formMessage.classList.remove('error');
                        formMessage.classList.add('success');
                        formMessage.style.display = 'block';
                        form.reset();
                    } else {
                        formMessage.innerText = 'Error creating product: ' + data.error;
                        formMessage.classList.remove('success');
                        formMessage.classList.add('error');
                        formMessage.style.display = 'block';
                    }
                })
                .catch(error => {
                    console.error('Error:', error);
                    formMessage.innerText = 'An unexpected error occurred: ' + error;
                    formMessage.classList.remove('success');
                    formMessage.classList.add('error');
                    formMessage.style.display = 'block';
                });
        }
    }
}

// Attach event listener when DOM is loaded
document.addEventListener('DOMContentLoaded', function () {
    const form = document.getElementById('addProductForm');
    if (form) {
        form.addEventListener('submit', createProduct);
    }

    // Add event listeners to all favorite buttons
    const favoriteButtons = document.querySelectorAll('.favorite-btn');
    favoriteButtons.forEach(button => {
        button.addEventListener('click', function (e) {
            e.preventDefault();
            e.stopPropagation();
            const productId = this.getAttribute('data-product-id');
            toggleFavorite(productId, this);
        });
    });
});

function toggleFavorite(productid, buttonElement) {
    const isFavorited = buttonElement.classList.contains('favorited');
    const endpoint = isFavorited ? '/api/unfavorite_product' : '/api/favorite_product';
    
    const formData = new FormData();
    formData.append("product_id", productid);
    
    fetch(endpoint, {
        method: "POST",
        body: formData,
        credentials: 'include'
    }).then(response => response.json())
        .then(data => {
            if (data.success) {
                if (isFavorited) {
                    buttonElement.classList.remove('favorited');
                } else {
                    buttonElement.classList.add('favorited');
                }
            } else {
                if (data.invalid) {
                    // Open the login modal (defined in navbar.js)
                    const modalOverlay = document.getElementById('modalOverlay');
                    if (modalOverlay) {
                        modalOverlay.classList.add('active');
                        document.body.style.overflow = 'hidden';
                        
                        // Switch to login tab
                        const loginTab = document.getElementById('loginTab');
                        const signupTab = document.getElementById('signupTab');
                        const loginContent = document.getElementById('loginContent');
                        const signupContent = document.getElementById('signupContent');
                        
                        if (loginTab && signupTab && loginContent && signupContent) {
                            loginTab.classList.add('active');
                            signupTab.classList.remove('active');
                            loginContent.classList.add('active');
                            signupContent.classList.remove('active');
                        }
                    }
                }
                console.error('Failed to toggle favorite:', data.error);
            }
        }
        ).catch(error => {
            console.error('Error toggling favorite:', error);
        });
}