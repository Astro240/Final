# Store Images Overview

## Introduction

This document provides an overview of the store images system used in the application. Store logos and banners are essential for creating a professional and branded storefront within the platform.

## Purpose

Store images serve several important functions:
- **Brand Identity**: Logos provide a visual representation of the store's brand and identity.
- **Visual Appeal**: Banners create an attractive hero section that captures customer attention.
- **Professionalism**: High-quality images contribute to the credibility and trustworthiness of the store.
- **Customization**: Store owners can upload custom logos and banners to personalize their storefronts.
- **Marketing**: Banners can be used to promote sales, special offers, or seasonal campaigns.

## Storage

### Store Logos
Store logos are stored with the following specifications:
- **File Type**: png, jpg, jpeg, gif
- **Recommended Resolution**: 200x200 pixels (square format)
- **Maximum File Size**: 2MB
- **Storage Format**: Base64 encoded strings in the database
- **File Location**: `backend/store_images/logos/`

### Store Banners
Store banners are stored with the following specifications:
- **File Type**: png, jpg, jpeg
- **Recommended Resolution**: 1200x400 pixels (3:1 aspect ratio)
- **Maximum File Size**: 5MB
- **Storage Format**: Base64 encoded strings in the database
- **File Location**: `backend/store_images/banners/`

## Access

Store images can be accessed and managed through:
- **Store Creation Process**: Upload during initial store setup (Step 2: Branding)
- **Store Dashboard**: Modify logos and banners after store creation
- **Admin Panel**: Store owners have full control over their store images

Ensure that appropriate permissions are in place for store owners to upload and modify their images.

## Image Guidelines

### Logo Best Practices
- Use transparent backgrounds (PNG format recommended)
- Keep design simple and recognizable at small sizes
- Use high contrast colors for better visibility
- Ensure the logo is centered and properly sized

### Banner Best Practices
- Use high-resolution images for clarity
- Place important content in the center (safe zone)
- Consider text readability if overlaying text
- Optimize file size for faster loading
- Use images that represent your brand or products

## Technical Implementation

Images are processed using the `ValidateImage` function which:
1. Validates file type and size
2. Converts base64 data to image files
3. Stores files in the appropriate directory
4. Returns the saved filename for database storage

## Additional Resources

For more information on image customization options, uploading guidelines, and branding best practices, please refer to the store owner documentation or contact support.