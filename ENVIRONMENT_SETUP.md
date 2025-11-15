# Environment Setup Guide

## Development vs Production Modes

The application supports two modes:

### Development Mode (Default)

- **Purpose**: Local development and testing
- **Features**:
  - Debug logging enabled
  - Detailed error messages
  - No payment processing (or test mode)
  - Relaxed security
  - Hot reload support

### Production Mode

- **Purpose**: Live application
- **Features**:
  - Optimized performance
  - Payment processing enabled
  - Enhanced security
  - Production logging
  - Error tracking

## Setting Environment Mode

### Backend (.env)

```env
# Application Mode
APP_MODE=development  # or "production"

# Development Mode Settings
DATABASE_URL=your_database_url
JWT_SECRET=your_jwt_secret
PORT=8080

# Production Mode Settings (when APP_MODE=production)
# Payment Gateways
STRIPE_SECRET_KEY=sk_live_...
STRIPE_PUBLISHABLE_KEY=pk_live_...
PAYPAL_CLIENT_ID=your_paypal_client_id
PAYPAL_SECRET=your_paypal_secret
PAYPAL_MODE=live  # or "sandbox" for testing

# Pakistani Payment Gateways
EASYPAISA_MERCHANT_ID=your_merchant_id
EASYPAISA_PASSWORD=your_password
EASYPAISA_STORE_ID=your_store_id

JAZZCASH_MERCHANT_ID=your_merchant_id
JAZZCASH_PASSWORD=your_password
JAZZCASH_INTEGRITY_SALT=your_salt
```

### Frontend (.env.local)

```env
# Application Mode
NEXT_PUBLIC_APP_MODE=development  # or "production"

# API Configuration
NEXT_PUBLIC_API_URL=http://localhost:8080  # Development
# NEXT_PUBLIC_API_URL=https://api.yourdomain.com  # Production

# Payment Configuration (Production only)
NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY=pk_test_...  # Development
# NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY=pk_live_...  # Production
```

## Mode Detection

The application automatically detects the mode:

**Backend:**

- Checks `APP_MODE` environment variable
- Defaults to `development` if not set
- Sets Gin framework mode accordingly

**Frontend:**

- Checks `NEXT_PUBLIC_APP_MODE` environment variable
- Can show/hide features based on mode
- Can use different API endpoints

## Switching Modes

### Development → Production

1. Update `.env` files:

   ```env
   APP_MODE=production
   ```

2. Set production environment variables:

   - Payment gateway keys
   - Production database URL
   - Production email settings

3. Restart application

### Production → Development

1. Update `.env` files:

   ```env
   APP_MODE=development
   ```

2. Use test/development credentials
3. Restart application

## Best Practices

1. **Never commit `.env` files** - Use `.env.example` instead
2. **Use different databases** for dev and prod
3. **Test payment flows** in development with test keys
4. **Monitor logs** differently in each mode
5. **Use feature flags** to enable/disable features by mode

## Feature Flags

You can check mode in code:

**Backend (Go):**

```go
import "ats-backend/config"

if config.IsProduction() {
    // Production-only code
} else {
    // Development code
}
```

**Frontend (TypeScript):**

```typescript
const isProduction = process.env.NEXT_PUBLIC_APP_MODE === "production";

if (isProduction) {
  // Production-only code
}
```
