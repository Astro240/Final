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
document.addEventListener('DOMContentLoaded', function() {
    const form = document.getElementById('addProductForm');
    if (form) {
        form.addEventListener('submit', createProduct);
    }
});