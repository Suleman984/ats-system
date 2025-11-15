# How to Add a Super Admin via Supabase

Since super admin registration is disabled for security, you need to add super admins directly in the database.

## Method 1: Using Supabase SQL Editor

1. Go to your Supabase project dashboard
2. Navigate to **SQL Editor**
3. Run this SQL (replace values with your details):

```sql
-- Generate password hash (you'll need to do this in Go or use online bcrypt tool)
-- For now, use this Go code to generate hash, or use: https://bcrypt-generator.com/

-- Example: Password "yourpassword123" hashed with bcrypt
-- Copy the hash from bcrypt-generator.com (cost 10)

INSERT INTO super_admin (name, email, password_hash, created_at)
VALUES (
  'Your Name',
  'your-email@example.com',
  '$2a$10$YourBcryptHashHere', -- Replace with actual bcrypt hash
  NOW()
);
```

## Method 2: Using Go to Generate Hash

1. Create a temporary Go file:

```go
package main

import (
    "fmt"
    "golang.org/x/crypto/bcrypt"
)

func main() {
    password := "yourpassword123"
    hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    fmt.Println(string(hash))
}
```

2. Run it: `go run hash.go`
3. Copy the hash and use it in the SQL above

## Method 3: Using Backend API (One-time)

If you need to add super admins programmatically, you can temporarily enable the register endpoint, add the admin, then disable it again.

## Security Note

- Super admin accounts have full platform access
- Only add trusted administrators
- Keep passwords secure
- Consider using strong, unique passwords
