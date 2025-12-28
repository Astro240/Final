const mobileMenuToggle = document.getElementById('mobileMenuToggle');
const mobileNav = document.getElementById('mobileNav');
const mobileNavClose = document.getElementById('mobileNavClose');
const mobileOverlay = document.getElementById('mobileOverlay');

function openMobileMenu() {
    mobileNav.classList.add('active');
    mobileOverlay.classList.add('active');
    mobileMenuToggle.classList.add('active');
    document.body.style.overflow = 'hidden';
}

function closeMobileMenu() {
    mobileNav.classList.remove('active');
    mobileOverlay.classList.remove('active');
    mobileMenuToggle.classList.remove('active');
    document.body.style.overflow = '';
}

mobileMenuToggle.addEventListener('click', openMobileMenu);
mobileNavClose.addEventListener('click', closeMobileMenu);
mobileOverlay.addEventListener('click', closeMobileMenu);

// Close menu when clicking on a link
document.querySelectorAll('.mobile-nav-links a').forEach(link => {
    link.addEventListener('click', closeMobileMenu);
});

const loginBtn = document.getElementById('loginBtn');
const mobileLoginBtn = document.getElementById('mobileLoginBtn');
const modalOverlay = document.getElementById('modalOverlay');
const modalClose = document.getElementById('modalClose');
const loginTab = document.getElementById('loginTab');
const signupTab = document.getElementById('signupTab');
const loginContent = document.getElementById('loginContent');
const signupContent = document.getElementById('signupContent');
const switchToSignup = document.getElementById('switchToSignup');
const switchToLogin = document.getElementById('switchToLogin');

function openModal() {
    modalOverlay.classList.add('active');
    document.body.style.overflow = 'hidden';
}

function closeModal() {
    modalOverlay.classList.remove('active');
    document.body.style.overflow = '';
}

loginBtn.addEventListener('click', openModal);
mobileLoginBtn.addEventListener('click', () => {
    closeMobileMenu();
    openModal();
});
modalClose.addEventListener('click', closeModal);
modalOverlay.addEventListener('click', (e) => {
    if (e.target === modalOverlay) closeModal();
});

loginTab.addEventListener('click', () => {
    loginTab.classList.add('active');
    signupTab.classList.remove('active');
    loginContent.classList.add('active');
    signupContent.classList.remove('active');
});

signupTab.addEventListener('click', () => {
    signupTab.classList.add('active');
    loginTab.classList.remove('active');
    signupContent.classList.add('active');
    loginContent.classList.remove('active');
});

switchToSignup.addEventListener('click', (e) => {
    e.preventDefault();
    signupTab.click();
});

switchToLogin.addEventListener('click', (e) => {
    e.preventDefault();
    loginTab.click();
});

// Form submissions (placeholder - integrate with backend)
document.getElementById('loginForm').addEventListener('submit', (e) => {
    e.preventDefault();
    const loginError = document.getElementById('loginError');
    loginError.style.display = 'none';
    
    const formData = new FormData(document.getElementById('loginForm'));
    const fullPath = window.location.href;
    fetch('/api/store_login', {
        method: 'POST',
        credentials: 'include',
        body: formData,
        headers: {
            'X-Store-Name': fullPath
        }
    }).then(response => {
        if (response.ok) {
            return response.json().then(data => {
                // Show 2FA verification modal
                closeModal();
                openVerifyModal(data.user_id, data.store_id, data.email);
            });
        } else {
            return response.json().then(data => {
                loginError.textContent = data.error || 'Login failed. Please check your credentials.';
                loginError.style.display = 'block';
            });
        }
    }).catch(error => {
        loginError.textContent = 'An error occurred. Please try again.';
        loginError.style.display = 'block';
    });
});

document.getElementById('signupForm').addEventListener('submit', (e) => {
    e.preventDefault();
    const signupError = document.getElementById('signupError');
    signupError.style.display = 'none';
    
    const formData = new FormData(document.getElementById('signupForm'));
    const fullPath = window.location.href;
    fetch('/api/store_register', {
        method: 'POST',
        credentials: 'include',
        body: formData,
        headers: {
            'X-Store-Name': fullPath
        }
    }).then(response => {
        if (response.ok) {
            return response.json().then(data => {
                // Show 2FA verification modal
                closeModal();
                openVerifyModal(data.user_id, data.store_id, data.email);
            });
        } else {
            return response.json().then(data => {
                signupError.textContent = data.error || 'Signup failed. Please try again.';
                signupError.style.display = 'block';
            });
        }
    }).catch(error => {
        signupError.textContent = 'An error occurred. Please try again.';
        signupError.style.display = 'block';
    });
});

// 2FA Verification Modal
const verifyModalOverlay = document.getElementById('verifyModalOverlay');
const verifyModalClose = document.getElementById('verifyModalClose');

function openVerifyModal(userId, storeId, email) {
    document.getElementById('verifyUserId').value = userId;
    document.getElementById('verifyStoreId').value = storeId;
    verifyModalOverlay.classList.add('active');
    document.body.style.overflow = 'hidden';
}

function closeVerifyModal() {
    verifyModalOverlay.classList.remove('active');
    document.body.style.overflow = '';
    document.getElementById('verifyForm').reset();
    document.getElementById('verifyError').style.display = 'none';
}

verifyModalClose.addEventListener('click', closeVerifyModal);
verifyModalOverlay.addEventListener('click', (e) => {
    if (e.target === verifyModalOverlay) {
        closeVerifyModal();
    }
});

document.getElementById('verifyForm').addEventListener('submit', (e) => {
    e.preventDefault();
    const verifyError = document.getElementById('verifyError');
    verifyError.style.display = 'none';
    
    const formData = new FormData(document.getElementById('verifyForm'));
    
    fetch('/api/store_verify', {
        method: 'POST',
        credentials: 'include',
        body: formData
    }).then(response => {
        if (response.ok) {
            closeVerifyModal();
            window.location.reload();
        } else {
            return response.json().then(data => {
                verifyError.textContent = data.error || 'Verification failed. Please check your code.';
                verifyError.style.display = 'block';
            });
        }
    }).catch(error => {
        verifyError.textContent = 'An error occurred. Please try again.';
        verifyError.style.display = 'block';
    });
});