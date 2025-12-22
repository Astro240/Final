let currentStep = 1;
const totalSteps = 7;

// Map error keywords to steps
const errorStepMap = {
    'store title': 1,
    'store name': 1,
    'storetitle': 1,
    'description': 1,
    'logo': 2,
    'banner': 2,
    'branding': 2,
    'template': 2,
    'color': 2,
    'email': 3,
    'phone': 3,
    'contact': 3,
    'address': 3,
    'payment': 4,
    'shipping': 4,
    'shipping cost': 4,
    'shipping threshold': 4,
    'categor': 5,
    'product': 5,
    'social': 6,
};

function getErrorStep(message) {
    const lowerMessage = message.toLowerCase();
    for (const [keyword, step] of Object.entries(errorStepMap)) {
        if (lowerMessage.includes(keyword)) {
            return step;
        }
    }
    return currentStep; // Stay on current step if no match
}

function showError(message, targetStep = null) {
    const errorBox = document.getElementById('errorBox');
    const errorMessage = document.getElementById('errorMessage');
    
    if (errorBox && errorMessage) {
        errorMessage.textContent = message;
        errorBox.style.display = 'block';
        errorBox.scrollIntoView({ behavior: 'smooth', block: 'start' });
        
        // Determine which step the error is from
        const errorStep = targetStep || getErrorStep(message);
        
        // Navigate to the step with the error
        if (errorStep !== currentStep) {
            setTimeout(() => {
                currentStep = errorStep;
                updateProgress();
                showStep(currentStep);
            }, 500);
        }
        
        // Auto-hide after 7 seconds
        setTimeout(() => {
            errorBox.style.display = 'none';
        }, 7000);
    }
}

function hideError() {
    const errorBox = document.getElementById('errorBox');
    if (errorBox) {
        errorBox.style.display = 'none';
    }
}

function updateProgress() {
    const progressFill = document.getElementById('progressFill');
    const percentage = ((currentStep - 1) / (totalSteps - 1)) * 100;
    progressFill.style.width = percentage + '%';

    // Update step indicators
    document.querySelectorAll('.step').forEach((step, index) => {
        const stepNumber = index + 1;
        step.classList.remove('active', 'completed');

        if (stepNumber < currentStep) {
            step.classList.add('completed');
            step.querySelector('.step-circle').innerHTML = '✓';
        } else if (stepNumber === currentStep) {
            step.classList.add('active');
            step.querySelector('.step-circle').innerHTML = stepNumber;
        } else {
            step.querySelector('.step-circle').innerHTML = stepNumber;
        }
    });
}

function showStep(step) {
    // Hide all steps
    document.querySelectorAll('.step-content').forEach(content => {
        content.classList.remove('active');
    });

    // Show current step
    document.getElementById('step' + step).classList.add('active');

    // Update navigation buttons
    const prevBtn = document.getElementById('prevBtn');
    const nextBtn = document.getElementById('nextBtn');

    if (step === 1) {
        prevBtn.style.display = 'none';
    } else {
        prevBtn.style.display = 'block';
    }

    if (step === totalSteps) {
        nextBtn.innerHTML = 'Create Store 🚀';
        nextBtn.onclick = submitForm;
    } else {
        nextBtn.innerHTML = 'Next →';
        nextBtn.onclick = () => changeStep(1);
    }
}

function changeStep(direction) {
    if (direction === 1 && !validateCurrentStep()) {
        return;
    }

    hideError();
    const newStep = currentStep + direction;

    if (newStep >= 1 && newStep <= totalSteps) {
        currentStep = newStep;
        updateProgress();
        showStep(currentStep);
    }
}

function validateCurrentStep() {
    const currentStepElement = document.getElementById('step' + currentStep);
    const requiredFields = currentStepElement.querySelectorAll('input[required], textarea[required]');

    for (let field of requiredFields) {
        if (!field.value.trim()) {
            field.focus();
            field.style.borderColor = '#d9534f';
            field.style.boxShadow = '0 0 0 3px rgba(217, 83, 79, 0.2)';
            
            // Get field label for better error message
            const label = currentStepElement.querySelector(`label[for="${field.id}"]`);
            const fieldName = label ? label.textContent : field.id;
            
            showError(`Please fill in the "${fieldName}" field.`);
            return false;
        } else {
            // Remove error styling
            field.style.borderColor = '';
            field.style.boxShadow = '';
        }
    }

    // Special validation for checkboxes
    if (currentStep === 4) {
        const paymentMethods = document.querySelectorAll('input[name="paymentMethods"]:checked');
        if (paymentMethods.length === 0) {
            showError('Please select at least one payment method.');
            return false;
        }
        
        // Validate IBAN
        const ibanField = document.getElementById('ibanNumber');
        if (!ibanField.value.trim()) {
            ibanField.focus();
            ibanField.style.borderColor = '#d9534f';
            ibanField.style.boxShadow = '0 0 0 3px rgba(217, 83, 79, 0.2)';
            showError('Please fill in the IBAN Number.');
            return false;
        } else {
            ibanField.style.borderColor = '';
            ibanField.style.boxShadow = '';
        }
    }

    if (currentStep === 5) {
        const categories = document.querySelectorAll('input[name="categories"]:checked');
        if (categories.length === 0) {
            showError('Please select at least one product category.');
            return false;
        }
    }

    return true;
}

function previewFile(input, previewId) {
    const file = input.files[0];
    const preview = document.getElementById(previewId);

    if (file) {
        const reader = new FileReader();
        reader.onload = function (e) {
            preview.innerHTML = `<img src="${e.target.result}" alt="Preview">`;
            preview.style.display = 'block';
        };
        reader.readAsDataURL(file);
    } else {
        preview.style.display = 'none';
    }
}

function submitForm() {
    const formData = new FormData(document.getElementById('storeCreationForm'));
    
    // Handle logo file upload as base64
    const logoInput = document.querySelector('input[name="storeLogo"]');
    const bannerInput = document.querySelector('input[name="storeBanner"]');
    
    let logoProcessed = false;
    let bannerProcessed = false;
    
    // Process logo file
    if (logoInput && logoInput.files[0]) {
        const logoFile = logoInput.files[0];
        const logoReader = new FileReader();
        logoReader.onloadend = function () {
            let base64data = logoReader.result;
            if (base64data === "data:,") {
                base64data = "";
            }
            formData.set('storeLogo', base64data);
            logoProcessed = true;
            checkAndSubmit();
        };
        logoReader.readAsDataURL(logoFile);
    } else {
        formData.set('storeLogo', '');
        logoProcessed = true;
    }
    
    // Process banner file
    if (bannerInput && bannerInput.files[0]) {
        const bannerFile = bannerInput.files[0];
        const bannerReader = new FileReader();
        bannerReader.onloadend = function () {
            let base64data = bannerReader.result;
            if (base64data === "data:,") {
                base64data = "";
            }
            formData.set('storeBanner', base64data);
            bannerProcessed = true;
            checkAndSubmit();
        };
        bannerReader.readAsDataURL(bannerFile);
    } else {
        formData.set('storeBanner', '');
        bannerProcessed = true;
    }
    
    function checkAndSubmit() {
        if (logoProcessed && bannerProcessed) {
            
            fetch('/api/create_store', {
                method: 'POST',
                body: formData,
                credentials: 'include'
            })
            .then(response => {
                if (response.ok) {
                    window.location.href = '/'+formData.get('storeTitle')+".com";
                } else {
                    return response.text().then(text => {
                        try {
                            const json = JSON.parse(text);
                            const errorMessage = json.error || 'Failed to create store';
                            showError(errorMessage);
                        } catch (e) {
                            showError(text || 'Failed to create store');
                        }
                    });
                }
            })
            .catch(error => {
                console.error('Fetch error:', error);
                showError('An unexpected error occurred: ' + error.message);
            });
        }
    }
    
    // If no files to process, submit immediately
    if (!logoInput?.files[0] && !bannerInput?.files[0]) {
        checkAndSubmit();
    }
}

// Initialize the form
updateProgress();
showStep(1);

// Get template from URL params
const urlParams = new URLSearchParams(window.location.search);
const template = urlParams.get('template');
if (template) {
    const templateNames = {
        'modern': 'Modern Minimalist',
        'vibrant': 'Vibrant Creative',
        'luxury': 'Luxury Premium'
    };

    const templatePreview = document.querySelector('.template-preview h4');
    if (templatePreview && templateNames[template]) {
        templatePreview.textContent = `Selected Template: ${templateNames[template]}`;
    }
}