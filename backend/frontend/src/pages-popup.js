// Pages Popup Module - Only for non-orders pages
const pagesPopup = {
    pages: {
        about: {
            title: 'About Us',
            content: `
                <div class="popup-page-content">
                    <h2>About Our Store</h2>
                    <p>Welcome to our premium shopping experience. We are dedicated to bringing you the finest products curated with excellence and care.</p>
                    <h3>Our Mission</h3>
                    <p>To provide exceptional products and outstanding customer service that exceeds expectations.</p>
                    <h3>Our Values</h3>
                    <ul>
                        <li><strong>Quality:</strong> We only offer products that meet our high standards</li>
                        <li><strong>Innovation:</strong> Constantly evolving to bring you the latest and greatest</li>
                        <li><strong>Integrity:</strong> Honest dealings and transparent practices</li>
                        <li><strong>Customer First:</strong> Your satisfaction is our priority</li>
                    </ul>
                </div>
            `
        },
        contact: {
            title: 'Contact Us',
            content: `
                <div class="popup-page-content">
                    <h2>Get In Touch</h2>
                    <h3>Contact Information</h3>
                    <div class="contact-info">
                        <p><strong>📧 Email:</strong> <span id="contact-email"></span></p>
                        <p><strong>⏰ Hours:</strong> Monday - Friday: 9:00 AM - 6:00 PM (EST)</p>
                    </div>
                </div>
            `
        },
        heritage: {
            title: 'Our Heritage',
            content: `
                <div class="popup-page-content">
                    <h2>A Legacy of Excellence</h2>
                    <p>Our heritage spans decades of commitment to quality and innovation.</p>
                    <h3>Timeline</h3>
                    <ul>
                        <li><strong>2010:</strong> Founded with a simple vision</li>
                        <li><strong>2013:</strong> Expanded to multiple categories</li>
                        <li><strong>2016:</strong> Launched our digital presence</li>
                        <li><strong>2020:</strong> Introduced 3D product viewing</li>
                        <li><strong>2024:</strong> Pioneering AR shopping experience</li>
                    </ul>
                </div>
            `
        },
        faq: {
            title: 'Frequently Asked Questions',
            content: `
                <div class="popup-page-content">
                    <h2>FAQ</h2>
                    <div class="faq-items">
                        <div class="faq-item">
                            <h4>How do I place an order?</h4>
                            <p>Browse our products, add items to your cart, and proceed to checkout.</p>
                        </div>
                        <div class="faq-item">
                            <h4>What payment methods do you accept?</h4>
                            <p>We accept all major credit cards, debit cards, and digital wallets.</p>
                        </div>
                        <div class="faq-item">
                            <h4>How long does shipping take?</h4>
                            <p>Standard shipping typically takes 5-7 business days.</p>
                        </div>
                        <div class="faq-item">
                            <h4>What's your return policy?</h4>
                            <p>We offer 30-day returns for unused items in original condition.</p>
                        </div>
                        <div class="faq-item">
                            <h4>Can I use the 3D viewer?</h4>
                            <p>Our 3D viewer is available for most products. Look for the AR button.</p>
                        </div>
                    </div>
                </div>
            `
        }
    },

    init() {
        this.createPopupStyles();
        window.showPagePopup = (page, ownerEmail) => this.openPage(page, ownerEmail);
    },

    openPage(page, ownerEmail) {
        const pageKey = page.replace('/', '').toLowerCase();
        const pageData = this.pages[pageKey];
        
        if (!pageData) {
            console.warn(`Page ${page} not found`);
            return;
        }

        let content = pageData.content;
        
        // If this is the contact page and we have owner email, inject it
        if (pageKey === 'contact' && ownerEmail) {
            content = pageData.content.replace(
                '<span id="contact-email"></span>',
                `<a href="mailto:${ownerEmail}">${ownerEmail}</a>`
            );
        }

        this.showPopup(pageData.title, content);
    },

    showPopup(title, content) {
        let overlay = document.getElementById('pagesPopupOverlay');
        
        if (!overlay) {
            overlay = document.createElement('div');
            overlay.id = 'pagesPopupOverlay';
            overlay.className = 'pages-popup-overlay';
            document.body.appendChild(overlay);
        }

        overlay.innerHTML = `
            <div class="pages-popup-modal">
                <div class="pages-popup-header">
                    <h2>${title}</h2>
                    <button class="pages-popup-close" onclick="pagesPopup.closePopup()">&times;</button>
                </div>
                <div class="pages-popup-body">
                    ${content}
                </div>
            </div>
        `;

        overlay.classList.add('active');
        overlay.onclick = (e) => {
            if (e.target === overlay) this.closePopup();
        };
    },

    closePopup() {
        const overlay = document.getElementById('pagesPopupOverlay');
        if (overlay) {
            overlay.classList.remove('active');
        }
    },

    createPopupStyles() {
        if (document.getElementById('pagesPopupStyles')) return;

        const style = document.createElement('style');
        style.id = 'pagesPopupStyles';
        style.textContent = `
            .pages-popup-overlay {
                display: none;
                position: fixed;
                top: 0;
                left: 0;
                right: 0;
                bottom: 0;
                background: rgba(0, 0, 0, 0.5);
                backdrop-filter: blur(5px);
                z-index: 10000;
                align-items: center;
                justify-content: center;
                padding: 20px;
            }

            .pages-popup-overlay.active {
                display: flex;
            }

            .pages-popup-modal {
                background: white;
                border-radius: 12px;
                max-width: 600px;
                width: 100%;
                max-height: 80vh;
                overflow-y: auto;
                box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
                animation: popupSlideIn 0.3s ease-out;
            }

            @keyframes popupSlideIn {
                from {
                    opacity: 0;
                    transform: scale(0.95);
                }
                to {
                    opacity: 1;
                    transform: scale(1);
                }
            }

            .pages-popup-header {
                display: flex;
                justify-content: space-between;
                align-items: center;
                padding: 2rem;
                border-bottom: 1px solid #e2e8f0;
                position: sticky;
                top: 0;
                background: white;
                border-radius: 12px 12px 0 0;
            }

            .pages-popup-header h2 {
                margin: 0;
                font-size: 1.8rem;
                color: #2d3748;
            }

            .pages-popup-close {
                background: none;
                border: none;
                font-size: 2.5rem;
                cursor: pointer;
                color: #718096;
                padding: 0;
                width: 40px;
                height: 40px;
                display: flex;
                align-items: center;
                justify-content: center;
                border-radius: 50%;
                transition: all 0.3s;
            }

            .pages-popup-close:hover {
                background: #f7fafc;
                color: #2d3748;
            }

            .pages-popup-body {
                padding: 2rem;
            }

            .popup-page-content h2 {
                color: #2d3748;
                margin-bottom: 1.5rem;
                font-size: 2rem;
            }

            .popup-page-content h3 {
                color: #4a5568;
                margin-top: 1.5rem;
                margin-bottom: 1rem;
                font-size: 1.3rem;
            }

            .popup-page-content h4 {
                color: #4a5568;
                margin-top: 1rem;
                margin-bottom: 0.5rem;
            }

            .popup-page-content p {
                color: #718096;
                line-height: 1.6;
                margin-bottom: 1rem;
            }

            .popup-page-content ul {
                margin: 1rem 0 1rem 2rem;
                color: #718096;
            }

            .popup-page-content ul li {
                margin-bottom: 0.8rem;
                line-height: 1.6;
            }

            .popup-page-content ul li strong {
                color: #2d3748;
            }

            .contact-info {
                background: #f7fafc;
                padding: 1.5rem;
                border-radius: 8px;
                margin: 1.5rem 0;
            }

            .contact-info p {
                margin-bottom: 1rem;
            }

            .contact-info a {
                color: #667eea;
                text-decoration: none;
            }

            .contact-info a:hover {
                text-decoration: underline;
            }

            .faq-items {
                display: flex;
                flex-direction: column;
                gap: 1rem;
            }

            .faq-item {
                background: #f7fafc;
                padding: 1.5rem;
                border-radius: 8px;
                border-left: 4px solid #667eea;
            }

            .faq-item h4 {
                color: #2d3748;
                margin: 0 0 0.5rem 0;
            }

            .faq-item p {
                margin: 0;
            }

            @media (max-width: 768px) {
                .pages-popup-modal {
                    max-width: 90vw;
                    max-height: 90vh;
                }

                .pages-popup-header {
                    padding: 1.5rem;
                }

                .pages-popup-header h2 {
                    font-size: 1.5rem;
                }

                .pages-popup-body {
                    padding: 1.5rem;
                }
            }
        `;
        document.head.appendChild(style);
    }
};

// Initialize on DOM ready
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => pagesPopup.init());
} else {
    pagesPopup.init();
}
