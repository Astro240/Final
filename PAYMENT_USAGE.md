# Payment System - Quick Start Guide

## How to Use the Payment System

### For New Installations
The payment system will work automatically when you run the application. The database schema includes all necessary columns.

### For Existing Databases
If you already have a database, run the migration script:

```bash
cd backend/database
sqlite3 database.db < migration_payment.sql
```

Or manually add the columns:
```sql
ALTER TABLE orders ADD COLUMN payment_info TEXT;
ALTER TABLE orders ADD COLUMN updated_at DATETIME DEFAULT CURRENT_TIMESTAMP;
```

## Testing the Payment Flow

1. **Start the server:**
   ```bash
   cd backend
   go run main.go
   # or
   ./server.exe
   ```

2. **Navigate to a store** (e.g., `http://localhost/store/[store-name]`)

3. **Add products to cart** - You'll see toast notifications

4. **Go to checkout** - Click the cart icon or checkout button

5. **Fill shipping information:**
   - Full Name
   - Email
   - Address
   - City
   - Country
   - (Optional: Phone, State, ZIP)

6. **Place Order** - Creates order with 'pending' status

7. **Payment Page** (auto-redirected):
   - Enter any cardholder name
   - Card Number: Any 13-19 digits (e.g., 4532 1234 5678 9010)
   - Expiry: MM/YY format (e.g., 12/25)
   - CVV: Any 3-4 digits (e.g., 123)

8. **Payment Success** - Order status changes to 'paid'

9. **Redirect to Home** - After 3 seconds

## Validation Rules

### Card Number
- Must be 13-19 digits
- Spaces are automatically added
- Example: `4532123456789010`

### Expiry Date
- Format: MM/YY
- Auto-formatted as you type
- Example: `12/25`

### CVV
- 3 or 4 digits
- Example: `123`

### All Fields Required
- Cardholder Name
- Card Number
- Expiry Date
- CVV

## Error Messages

All error messages appear in styled divs (no popup alerts):

- **Cart operations**: Toast notifications (top-right corner)
- **Checkout**: Error div at top of page
- **Payment**: Error div at top of page

## API Endpoints

### POST `/api/create-order`
Creates a new order with shipping information.

**Form Data:**
- `full_name`
- `email`
- `phone` (optional)
- `address`
- `city`
- `state` (optional)
- `zip_code` (optional)
- `country`

**Response:**
```json
{
  "success": true,
  "order_id": 123,
  "message": "Order created successfully"
}
```

### GET `/api/pending-order`
Retrieves the most recent pending order for the authenticated user.

**Response:**
```json
{
  "success": true,
  "order": {
    "order_id": 123,
    "total_amount": 99.99,
    "shipping_info": "...",
    "created_at": "2025-12-09 10:30:00"
  }
}
```

### POST `/api/process-payment`
Processes payment for an order (fake payment - always succeeds).

**Form Data:**
- `order_id`
- `card_holder`
- `card_number`
- `expiry`
- `cvv`

**Response:**
```json
{
  "success": true,
  "message": "Payment processed successfully",
  "order_id": 123
}
```

## Security Notes

⚠️ **This is a demonstration payment system**

- Payment always succeeds for valid input
- Only last 4 digits of card are stored
- No actual payment processing occurs
- For production, integrate with:
  - Stripe
  - PayPal
  - Square
  - Other PCI-DSS compliant gateways

## Troubleshooting

### "No pending order found"
- Complete the checkout process first
- Make sure you're logged in
- Try adding items to cart again

### "Payment failed"
- Check all form fields are filled
- Verify card number is 13-19 digits
- Ensure CVV is 3-4 digits
- Check expiry format is MM/YY

### Order still showing 'pending'
- Check browser console for errors
- Verify payment was actually submitted
- Check database manually: `SELECT * FROM orders WHERE status='pending';`

## Features

✅ No alert() popups anywhere
✅ Toast notifications for cart operations  
✅ Styled error messages in divs
✅ Card number auto-formatting
✅ Expiry date auto-formatting
✅ Input validation
✅ Masked card storage (last 4 digits)
✅ Order status tracking
✅ Automatic redirects
✅ Smooth animations
✅ Mobile responsive
