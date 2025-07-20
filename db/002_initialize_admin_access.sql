-- Pastikan ekstensi pgcrypto sudah aktif untuk gen_random_uuid()
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Clear existing data (use TRUNCATE if tables exist and you want to re-insert)
TRUNCATE TABLE users, roles, permissions, sites, user_roles, role_permissions CASCADE;

-- Insert Dummy Sites
INSERT INTO sites (id, site_name, site_slug, created_at, updated_at) VALUES
('a1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'CMS Application', 'cms-app', NOW(), NOW()),
('31eebc99-9c0b-4ef8-bb6d-6bb9bd380a99', 'Mobile Application', 'mobile-app', NOW(), NOW());

-- Insert Dummy Roles
INSERT INTO roles (id, name, description, created_at, updated_at) VALUES
('01eebc99-9c0b-4ef8-bb6d-6bb9bd380a00', 'superadmin', 'Has full control over the system', NOW(), NOW()),
('b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 'admin', 'Manages users and roles globally', NOW(), NOW()),
('c1eebc99-9c0b-4ef8-bb6d-6bb9bd380a33', 'support', 'Can view user and system information', NOW(), NOW()),
('d1eebc99-9c0b-4ef8-bb6d-6bb9bd380a44', 'consumer', 'Standard user with access to their own profile and content', NOW(), NOW());

-- Insert Dummy Permissions (Hierarchical & Global)
-- Site: CMS Application (a1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11)
INSERT INTO permissions (id, site_id, name, description, parent_id, created_at, updated_at) VALUES
('e1eebc99-9c0b-4ef8-bb6d-6bb9bd380a55', 'a1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'user:manage', 'Manage all users (parent)', NULL, NOW(), NOW()),
('f1eebc99-9c0b-4ef8-bb6d-6bb9bd380a66', 'a1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'user:create', 'Create new users', 'e1eebc99-9c0b-4ef8-bb6d-6bb9bd380a55', NOW(), NOW()),
('41eebc99-9c0b-4ef8-bb6d-6bb9bd380b00', 'a1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'user:read_all', 'Read all user profiles', 'e1eebc99-9c0b-4ef8-bb6d-6bb9bd380a55', NOW(), NOW()),
('51eebc99-9c0b-4ef8-bb6d-6bb9bd380b11', 'a1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'user:update_all', 'Update any user profile', 'e1eebc99-9c0b-4ef8-bb6d-6bb9bd380a55', NOW(), NOW()),
('61eebc99-9c0b-4ef8-bb6d-6bb9bd380b22', 'a1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'user:delete', 'Delete users', 'e1eebc99-9c0b-4ef8-bb6d-6bb9bd380a55', NOW(), NOW()),
('71eebc99-9c0b-4ef8-bb6d-6bb9bd380b33', 'a1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'role:assign', 'Assign roles to users', NULL, NOW(), NOW()), -- Not hierarchical under user:manage
('81eebc99-9c0b-4ef8-bb6d-6bb9bd380b44', 'a1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'user:read_own', 'Read own user profile', NULL, NOW(), NOW()),
('91eebc99-9c0b-4ef8-bb6d-6bb9bd380b55', 'a1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'user:update_own', 'Update own user profile', NULL, NOW(), NOW()),
('a2eebc99-9c0b-4ef8-bb6d-6bb9bd380a01', 'a1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'cms:content:read', 'Read content in CMS', NULL, NOW(), NOW());

-- Insert Dummy Users
INSERT INTO users (id, username, email, password_hash, is_active, created_at, updated_at) VALUES
('11eebc99-9c0b-4ef8-bb6d-6bb9bd380a77', 'superadmin', 'superadmin@example.com', '$2a$10$8DTU8kmDKM4lwJ4X6OPBx.4.mTBC4uf7mEb0Pm9/SVwZnPQ3aSDJa', TRUE, NOW(), NOW()),
('21eebc99-9c0b-4ef8-bb6d-6bb9bd380a88', 'admin', 'admin@example.com', '$2a$10$8DTU8kmDKM4lwJ4X6OPBx.4.mTBC4uf7mEb0Pm9/SVwZnPQ3aSDJa', TRUE, NOW(), NOW()),
('31eebc99-9c0b-4ef8-bb6d-6bb9bd380a99', 'support', 'support@example.com', '$2a$10$8DTU8kmDKM4lwJ4X6OPBx.4.mTBC4uf7mEb0Pm9/SVwZnPQ3aSDJa', TRUE, NOW(), NOW()),
('41eebc99-9c0b-4ef8-bb6d-6bb9bd380b00', 'consumer', 'consumer@example.com', '$2a$10$8DTU8kmDKM4lwJ4X6OPBx.4.mTBC4uf7mEb0Pm9/SVwZnPQ3aSDJa', TRUE, NOW(), NOW());


-- Assign Roles to Users (Global Roles)
INSERT INTO user_roles (user_id, role_id, created_at) VALUES
('11eebc99-9c0b-4ef8-bb6d-6bb9bd380a77', '01eebc99-9c0b-4ef8-bb6d-6bb9bd380a00', NOW()), -- superadmin gets superadmin role
('21eebc99-9c0b-4ef8-bb6d-6bb9bd380a88', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', NOW()), -- admin gets admin role
('31eebc99-9c0b-4ef8-bb6d-6bb9bd380a99', 'c1eebc99-9c0b-4ef8-bb6d-6bb9bd380a33', NOW()), -- support gets support role
('41eebc99-9c0b-4ef8-bb6d-6bb9bd380b00', 'd1eebc99-9c0b-4ef8-bb6d-6bb9bd380a44', NOW()); -- consumer gets consumer role

-- Assign Permissions to Roles (Role-Permission mapping)
-- superadmin role: all permissions
INSERT INTO role_permissions (role_id, permission_id, created_at) VALUES
('01eebc99-9c0b-4ef8-bb6d-6bb9bd380a00', 'e1eebc99-9c0b-4ef8-bb6d-6bb9bd380a55', NOW()), -- superadmin -> user:manage (parent)
('01eebc99-9c0b-4ef8-bb6d-6bb9bd380a00', '71eebc99-9c0b-4ef8-bb6d-6bb9bd380b33', NOW()), -- superadmin -> role:assign
('01eebc99-9c0b-4ef8-bb6d-6bb9bd380a00', '81eebc99-9c0b-4ef8-bb6d-6bb9bd380b44', NOW()), -- superadmin -> user:read_own
('01eebc99-9c0b-4ef8-bb6d-6bb9bd380a00', '91eebc99-9c0b-4ef8-bb6d-6bb9bd380b55', NOW()), -- superadmin -> user:update_own
('01eebc99-9c0b-4ef8-bb6d-6bb9bd380a00', 'a2eebc99-9c0b-4ef8-bb6d-6bb9bd380a01', NOW()); -- superadmin -> cms:content:read

-- admin role: user management (excluding superadmin), role assignment
INSERT INTO role_permissions (role_id, permission_id, created_at) VALUES
('b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 'e1eebc99-9c0b-4ef8-bb6d-6bb9bd380a55', NOW()), -- admin -> user:manage (parent)
('b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', '71eebc99-9c0b-4ef8-bb6d-6bb9bd380b33', NOW()); -- admin -> role:assign

-- support role: read all users
INSERT INTO role_permissions (role_id, permission_id, created_at) VALUES
('c1eebc99-9c0b-4ef8-bb6d-6bb9bd380a33', '41eebc99-9c0b-4ef8-bb6d-6bb9bd380b00', NOW()); -- support -> user:read_all

-- consumer role: read/update own profile, read content
INSERT INTO role_permissions (role_id, permission_id, created_at) VALUES
('d1eebc99-9c0b-4ef8-bb6d-6bb9bd380a44', '81eebc99-9c0b-4ef8-bb6d-6bb9bd380b44', NOW()), -- consumer -> user:read_own
('d1eebc99-9c0b-4ef8-bb6d-6bb9bd380a44', '91eebc99-9c0b-4ef8-bb6d-6bb9bd380b55', NOW()), -- consumer -> user:update_own
('d1eebc99-9c0b-4ef8-bb6d-6bb9bd380a44', 'a2eebc99-9c0b-4ef8-bb6d-6bb9bd380a01', NOW()); -- consumer -> cms:content:read
