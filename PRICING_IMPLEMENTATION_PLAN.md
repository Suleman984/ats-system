# Pricing & Payment Implementation Plan

## Overview

This document outlines the implementation plan for adding pricing and payment features to the ATS system with support for both development and production modes.

## Payment Methods Recommended

### Universal (Worldwide)

1. **Stripe** ⭐ (Primary)

   - 40+ countries
   - Best API and documentation
   - 2.9% + $0.30 per transaction

2. **PayPal** ⭐ (Secondary)
   - 200+ countries (including Pakistan)
   - 3.4% + PKR 35 per transaction (Pakistan)

### Pakistan-Specific

3. **EasyPaisa**

   - Most popular mobile wallet (30M+ users)
   - ~1-2% transaction fee

4. **JazzCash**

   - Second most popular (20M+ users)
   - ~1-2% transaction fee

5. **Bank Transfer (IBFT)**
   - Manual verification
   - Universal fallback

## Implementation Status

### ✅ Completed

- Environment mode system (dev/prod)
- Database schema for subscriptions and payments
- Payment service structure
- Configuration system

### 🔄 In Progress

- Stripe integration
- PayPal integration
- EasyPaisa integration
- JazzCash integration
- Subscription management
- Pricing UI

## Next Steps

### Phase 1: Stripe Integration (Week 1)

1. Install Stripe Go SDK: `go get github.com/stripe/stripe-go/v76`
2. Implement Stripe payment processing
3. Add webhook handling for subscription events
4. Test with Stripe test keys

### Phase 2: PayPal Integration (Week 2)

1. Install PayPal SDK or use REST API
2. Implement PayPal payment processing
3. Add webhook handling
4. Test with PayPal sandbox

### Phase 3: Pakistani Gateways (Week 3-4)

1. Apply for EasyPaisa API access
2. Apply for JazzCash API access
3. Implement both integrations
4. Test thoroughly

### Phase 4: Frontend UI (Week 5)

1. Create pricing page
2. Create subscription management page
3. Add payment method selection
4. Add subscription status display

## Environment Variables Needed

```env
# Stripe
STRIPE_SECRET_KEY=sk_test_...
STRIPE_PUBLISHABLE_KEY=pk_test_...

# PayPal
PAYPAL_CLIENT_ID=...
PAYPAL_SECRET=...
PAYPAL_MODE=sandbox

# EasyPaisa
EASYPAISA_MERCHANT_ID=...
EASYPAISA_PASSWORD=...
EASYPAISA_STORE_ID=...

# JazzCash
JAZZCASH_MERCHANT_ID=...
JAZZCASH_PASSWORD=...
JAZZCASH_INTEGRITY_SALT=...
```

## Database Tables Created

1. **subscription_plans** - Available pricing plans
2. **subscriptions** - Company subscriptions
3. **payments** - Payment transactions

See `supabase/schema.sql` for full schema.
