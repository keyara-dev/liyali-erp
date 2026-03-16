-- Migration 026: Seed a dedicated super_admin user (platform-wide, no org)
-- Password: password (bcrypt hash)

INSERT INTO users (id, email, name, password, role, active, current_organization_id, is_super_admin, created_at, updated_at)
VALUES (
    'user-super-admin-001',
    'superadmin@liyali.com',
    'Super Admin',
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
    'super_admin',
    true,
    NULL,
    true,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
)
ON CONFLICT (id) DO NOTHING;
