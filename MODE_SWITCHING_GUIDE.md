# Quick Guide: Switching Between Development and Production Modes

## Backend

### Development Mode (Default)

```env
# backend/.env
APP_MODE=development
```

### Production Mode

```env
# backend/.env
APP_MODE=production
```

**That's it!** Just change the value and restart the backend server.

---

## Frontend

### Development Mode (Default)

```env
# frontend/.env.local
NEXT_PUBLIC_APP_MODE=development
```

### Production Mode

```env
# frontend/.env.local
NEXT_PUBLIC_APP_MODE=production
```

**Restart the frontend server** after changing.

---

## Quick Switch

1. **Edit `.env` files** (backend and frontend)
2. **Change `APP_MODE`** to `development` or `production`
3. **Restart servers**
4. **Check the indicator** on the frontend to confirm mode

The frontend will show a badge indicating which mode is active!
