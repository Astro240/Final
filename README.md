# Welcome to Our Marketplace Platform

A full-featured e-commerce marketplace where store owners can set up shop, manage inventory, and customers can browse, purchase, and leave reviews. Built with Go on the backend and vanilla JavaScript on the frontend.

## What's Inside

### Backend (Go)
The backend handles all the heavy lifting - user authentication, product management, order processing, and real-time features. Everything is built with Go for speed and reliability.

- **User Management**: Registration, login, profiles, and avatars
- **Store System**: Create and manage multiple stores with customizable templates
- **Products & Inventory**: Add products with descriptions, pricing, and 3D previews
- **Shopping Cart & Checkout**: Full cart management with secure payment processing
- **Orders**: Track orders, manage fulfillment, and keep purchase history
- **Reviews & Ratings**: Let customers share feedback on products
- **Admin Dashboard**: Tools for managing the platform and store operations
- **3D Visualization**: Interactive 3D product previews using Three.js
- **Email Notifications**: Automated emails for registrations, orders, and updates
- **WebSocket Support**: Real-time features like live notifications

### Frontend
The frontend is clean and straightforward - just HTML, CSS, and JavaScript. No heavy frameworks, just solid vanilla JS that does the job.

- **Responsive Design**: Works great on desktop, tablet, and mobile
- **Interactive Pages**: Product browsing, store front, shopping cart, checkout, and user dashboard
- **3D Globe**: Geographic visualization of your marketplace
- **Dynamic Content**: Real-time UI updates without page reloads

## Getting Started

### Requirements
- Go 1.16 or higher
- A modern web browser

### Installation

1. Clone the repo
```bash
git clone <repository-url>
cd Final
```

2. Navigate to the backend folder
```bash
cd backend
```

3. Install Go dependencies
```bash
go mod download
```

4. Configure your environment
Create a `.env` file in the backend folder with your settings:
```
GMAIL_KEY=MY_GMAIL_KEY (This will only work with my key, refer to the readme doc in the submission)
```

5. Run the backend server
```bash
go run main.go
```

The server will start and you can access the frontend by opening your browser to `https://localhost`

## Project Structure

```
/backend       - Go API server and business logic
/frontend      - HTML, CSS, and JavaScript files
/store_images  - Uploaded store logos, banners, and product images
/avatars       - User profile pictures
```

## Features Worth Noting

- **Custom Store Templates**: Different visual themes to choose from
- **3D Product Previews**: Show off products in interactive 3D
- **Real-time Updates**: WebSocket connections for live data
- **Secure Authentication**: JWT-based user sessions
- **Email Integration**: Automated communications with customers

## Development

Want to contribute? Check out the code structure:
- API endpoints are organized by feature (users, products, orders, etc.)
- Frontend templates use a simple naming convention
- Database queries are centralized for easier maintenance

## Support

If you run into issues or have questions, feel free to reach out or check the documentation in individual folders.

Enjoy building!