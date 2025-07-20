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
('b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 'admin', 'Global Administrator Role', NOW(), NOW()),
('c1eebc99-9c0b-4ef8-bb6d-6bb9bd380a33', 'user', 'Standard User Role', NOW(), NOW());

-- Insert Dummy Permissions (Hierarchical)
-- Parent Permissions
INSERT INTO permissions (id, site_id, name, description, parent_id, created_at, updated_at) VALUES
('d1eebc99-9c0b-4ef8-bb6d-6bb9bd380a44', 'a1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'cms:admin', 'Overall CMS Admin access', NULL, NOW(), NOW()),
('41eebc99-9c0b-4ef8-bb6d-6bb9bd380b00', 'a1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'cms:content:read', 'Read any content in CMS', NULL, NOW(), NOW()),
('71eebc99-9c0b-4ef8-bb6d-6bb9bd380b33', '31eebc99-9c0b-4ef8-bb6d-6bb9bd380a99', 'mobile:read', 'Read basic mobile app data', NULL, NOW(), NOW());

-- Child Permissions
INSERT INTO permissions (id, site_id, name, description, parent_id, created_at, updated_at) VALUES
('e1eebc99-9c0b-4ef8-bb6d-6bb9bd380a55', 'a1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'cms:admin:users:manage', 'Manage users in CMS', 'd1eebc99-9c0b-4ef8-bb6d-6bb9bd380a44', NOW(), NOW()),
('f1eebc99-9c0b-4ef8-bb6d-6bb9bd380a66', 'a1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'cms:admin:content:manage', 'Manage all content in CMS', 'd1eebc99-9c0b-4ef8-bb6d-6bb9bd380a44', NOW(), NOW()),
('51eebc99-9c0b-4ef8-bb6d-6bb9bd380b11', 'a1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'cms:content:view_public', 'View public content in CMS', '41eebc99-9c0b-4ef8-bb6d-6bb9bd380b00', NOW(), NOW()),
('61eebc99-9c0b-4ef8-bb6d-6bb9bd380b22', '31eebc99-9c0b-4ef8-bb6d-6bb9bd380a99', 'cms:content:view_drafts', 'View draft content in CMS', '41eebc99-9c0b-4ef8-bb6d-6bb9bd380b00', NOW(), NOW());

-- Insert Dummy Users
INSERT INTO users (id, username, email, password_hash, is_active, created_at, updated_at) VALUES
('10eebc99-9c0b-4ef8-bb6d-6bb9bd380a77', 'admin', 'admin@example.com', '$2a$10$8DTU8kmDKM4lwJ4X6OPBx.4.mTBC4uf7mEb0Pm9/SVwZnPQ3aSDJa', TRUE, NOW(), NOW()),
('20eebc99-9c0b-4ef8-bb6d-6bb9bd380a88', 'testuser', 'testuser@example.com', '$2a$10$8DTU8kmDKM4lwJ4X6OPBx.4.mTBC4uf7mEb0Pm9/SVwZnPQ3aSDJa', TRUE, NOW(), NOW());

-- Assign Roles to Users (Global Roles)
INSERT INTO user_roles (user_id, role_id, created_at) VALUES
('10eebc99-9c0b-4ef8-bb6d-6bb9bd380a77', 'b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', NOW()), -- admin user gets admin role
('20eebc99-9c0b-4ef8-bb6d-6bb9bd380a88', 'c1eebc99-9c0b-4ef8-bb6d-6bb9bd380a33', NOW()); -- testuser gets user role

-- Assign Permissions to Roles (Role-Permission mapping)
-- Admin role gets top-level 'cms:admin' and 'mobile:read' permissions
INSERT INTO role_permissions (role_id, permission_id, created_at) VALUES
('b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 'd1eebc99-9c0b-4ef8-bb6d-6bb9bd380a44', NOW()), -- admin -> cms:admin (this should grant its children)
('b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', '71eebc99-9c0b-4ef8-bb6d-6bb9bd380b33', NOW()); -- admin -> mobile:read

-- User role gets top-level 'cms:content:read' and 'mobile:read' permissions
INSERT INTO role_permissions (role_id, permission_id, created_at) VALUES
('c1eebc99-9c0b-4ef8-bb6d-6bb9bd380a33', '41eebc99-9c0b-4ef8-bb6d-6bb9bd380b00', NOW()), -- user -> cms:content:read (this should grant its children)
('c1eebc99-9c0b-4ef8-bb6d-6bb9bd380a33', '71eebc99-9c0b-4ef8-bb6d-6bb9bd380b33', NOW()); -- user -> mobile:read
