-- name: CreateRole :one
INSERT INTO roles (role_name, description)
VALUES ($1, $2)
RETURNING id, role_name, description, created_at, updated_at;

-- name: GetRoleByID :one
SELECT id, role_name, description, created_at, updated_at
FROM roles
WHERE id = $1;

-- name: GetRoleByName :one
SELECT id, role_name
FROM roles
WHERE role_name = $1;

-- name: GetAllRoles :many
SELECT id, role_name, description, created_at, updated_at
FROM roles
ORDER BY role_name;

-- name: UpdateRole :one
UPDATE roles
SET role_name = $2, description = $3, updated_at = now()
WHERE id = $1
RETURNING id, role_name, description, created_at, updated_at;

-- name: DeleteRole :exec
DELETE FROM roles
WHERE id = $1;


-- First, ensure the uuid-ossp extension is enabled to generate UUIDs
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Top-level categories (Level 0)
INSERT INTO categories (id, name, parent_id, level)
VALUES
    (uuid_generate_v4(), 'Computing', NULL, 0),
    (uuid_generate_v4(), 'Power Solutions', NULL, 0),
    (uuid_generate_v4(), 'Computer Accessories', NULL, 0),
    (uuid_generate_v4(), 'Security Systems', NULL, 0),
    (uuid_generate_v4(), 'Networking', NULL, 0);
-- Subcategories under 'Computing' (Level 1)
WITH computing_id AS (
    SELECT id FROM categories WHERE name = 'Computing'
)
INSERT INTO categories (id, name, parent_id, level)
VALUES
    (uuid_generate_v4(), 'Desktops', (SELECT id FROM computing_id), 1),
    (uuid_generate_v4(), 'Laptops', (SELECT id FROM computing_id), 1),
    (uuid_generate_v4(), 'Printers', (SELECT id FROM computing_id), 1),
    (uuid_generate_v4(), 'Software', (SELECT id FROM computing_id), 1),
    (uuid_generate_v4(), 'Printer Supplies', (SELECT id FROM computing_id), 1),
    (uuid_generate_v4(), 'Antivirus', (SELECT id FROM computing_id), 1);

-- Subcategories under 'Desktops' (Level 2)
WITH desktops_id AS (
    SELECT id FROM categories WHERE name = 'Desktops'
)
INSERT INTO categories (id, name, parent_id, level)
VALUES
    (uuid_generate_v4(), 'HP', (SELECT id FROM desktops_id), 2),
    (uuid_generate_v4(), 'Lenovo', (SELECT id FROM desktops_id), 2),
    (uuid_generate_v4(), 'DELL', (SELECT id FROM desktops_id), 2);

-- Subcategories under 'Laptops' (Level 2)
WITH laptops_id AS (
    SELECT id FROM categories WHERE name = 'Laptops'
)
INSERT INTO categories (id, name, parent_id, level)
VALUES
    (uuid_generate_v4(), 'HP', (SELECT id FROM laptops_id), 2),
    (uuid_generate_v4(), 'Lenovo', (SELECT id FROM laptops_id), 2),
    (uuid_generate_v4(), 'DELL', (SELECT id FROM laptops_id), 2),
    (uuid_generate_v4(), 'Apple', (SELECT id FROM laptops_id), 2),
    (uuid_generate_v4(), 'ASUS', (SELECT id FROM laptops_id), 2);

-- Subcategories under 'Printers' (Level 2)
WITH printers_id AS (
    SELECT id FROM categories WHERE name = 'Printers'
)
INSERT INTO categories (id, name, parent_id, level)
VALUES
    (uuid_generate_v4(), 'HP', (SELECT id FROM printers_id), 2),
    (uuid_generate_v4(), 'Epson', (SELECT id FROM printers_id), 2),
    (uuid_generate_v4(), 'Point of Sale', (SELECT id FROM printers_id), 2);

-- Subcategories under 'Software' (Level 2)
WITH software_id AS (
    SELECT id FROM categories WHERE name = 'Software'
)
INSERT INTO categories (id, name, parent_id, level)
VALUES
    (uuid_generate_v4(), 'Microsoft', (SELECT id FROM software_id), 2);

-- Subcategories under 'Printer Supplies' (Level 2)
WITH supplies_id AS (
    SELECT id FROM categories WHERE name = 'Printer Supplies'
)
INSERT INTO categories (id, name, parent_id, level)
VALUES
    (uuid_generate_v4(), 'HP Cartridges', (SELECT id FROM supplies_id), 2),
    (uuid_generate_v4(), 'Epson Cartridges', (SELECT id FROM supplies_id), 2),
    (uuid_generate_v4(), 'Brother', (SELECT id FROM supplies_id), 2),
    (uuid_generate_v4(), 'Mercury Toners', (SELECT id FROM supplies_id), 2),
    (uuid_generate_v4(), 'HP Toners', (SELECT id FROM supplies_id), 2);

-- Subcategories under 'Antivirus' (Level 2)
WITH antivirus_id AS (
    SELECT id FROM categories WHERE name = 'Antivirus'
)
INSERT INTO categories (id, name, parent_id, level)
VALUES
    (uuid_generate_v4(), 'Quickheal', (SELECT id FROM antivirus_id), 2),
    (uuid_generate_v4(), 'Kaspersky', (SELECT id FROM antivirus_id), 2),
    (uuid_generate_v4(), 'Seqrite', (SELECT id FROM antivirus_id), 2);

-- Subcategories under 'Power Solutions' (Level 1)
WITH power_id AS (
    SELECT id FROM categories WHERE name = 'Power Solutions'
)
INSERT INTO categories (id, name, parent_id, level)
VALUES
    (uuid_generate_v4(), 'UPS', (SELECT id FROM power_id), 1),
    (uuid_generate_v4(), 'Ups Batteries', (SELECT id FROM power_id), 1),
    (uuid_generate_v4(), 'Inverters and Solar', (SELECT id FROM power_id), 1),
    (uuid_generate_v4(), 'Sollatek', (SELECT id FROM power_id), 1);

-- Subcategories under 'UPS' (Level 2)
WITH ups_id AS (
    SELECT id FROM categories WHERE name = 'UPS'
)
INSERT INTO categories (id, name, parent_id, level)
VALUES
    (uuid_generate_v4(), 'Mecer', (SELECT id FROM ups_id), 2),
    (uuid_generate_v4(), 'APC', (SELECT id FROM ups_id), 2),
    (uuid_generate_v4(), 'Mercury', (SELECT id FROM ups_id), 2);

-- Subcategories under 'Ups Batteries' (Level 2)
WITH batteries_id AS (
    SELECT id FROM categories WHERE name = 'Ups Batteries'
)
INSERT INTO categories (id, name, parent_id, level)
VALUES
    (uuid_generate_v4(), 'Mercury Elite', (SELECT id FROM batteries_id), 2);

-- Subcategories under 'Sollatek' (Level 2)
WITH sollatek_id AS (
    SELECT id FROM categories WHERE name = 'Sollatek'
)
INSERT INTO categories (id, name, parent_id, level)
VALUES
    (uuid_generate_v4(), 'AVS', (SELECT id FROM sollatek_id), 2),
    (uuid_generate_v4(), 'AVR', (SELECT id FROM sollatek_id), 2),
    (uuid_generate_v4(), 'SVS', (SELECT id FROM sollatek_id), 2),
    (uuid_generate_v4(), 'Power Suppressors', (SELECT id FROM sollatek_id), 2),
    (uuid_generate_v4(), 'Solar Panels', (SELECT id FROM sollatek_id), 2);

-- Subcategories under 'Computer Accessories' (Level 1)
WITH accessories_id AS (
    SELECT id FROM categories WHERE name = 'Computer Accessories'
)
INSERT INTO categories (id, name, parent_id, level)
VALUES
    (uuid_generate_v4(), 'Monitors', (SELECT id FROM accessories_id), 1),
    (uuid_generate_v4(), 'Projectors', (SELECT id FROM accessories_id), 1),
    (uuid_generate_v4(), 'Storage Devices', (SELECT id FROM accessories_id), 1),
    (uuid_generate_v4(), 'Accessories', (SELECT id FROM accessories_id), 1);

-- Subcategories under 'Monitors' (Level 2)
WITH monitors_id AS (
    SELECT id FROM categories WHERE name = 'Monitors'
)
INSERT INTO categories (id, name, parent_id, level)
VALUES
    (uuid_generate_v4(), 'HP', (SELECT id FROM monitors_id), 2),
    (uuid_generate_v4(), 'Dell', (SELECT id FROM monitors_id), 2),
    (uuid_generate_v4(), 'Lenovo', (SELECT id FROM monitors_id), 2);

-- Subcategories under 'Projectors' (Level 2)
WITH projectors_id AS (
    SELECT id FROM categories WHERE name = 'Projectors'
)
INSERT INTO categories (id, name, parent_id, level)
VALUES
    (uuid_generate_v4(), 'Epson', (SELECT id FROM projectors_id), 2),
    (uuid_generate_v4(), 'Sony', (SELECT id FROM projectors_id), 2);

-- Subcategories under 'Storage Devices' (Level 2)
WITH storage_id AS (
    SELECT id FROM categories WHERE name = 'Storage Devices'
)
INSERT INTO categories (id, name, parent_id, level)
VALUES
    (uuid_generate_v4(), 'Hard Drives', (SELECT id FROM storage_id), 2),
    (uuid_generate_v4(), 'Flash Drives', (SELECT id FROM storage_id), 2),
    (uuid_generate_v4(), 'Memory Cards', (SELECT id FROM storage_id), 2);

-- Subcategories under 'Accessories' (Level 2)
WITH sub_accessories_id AS (
    SELECT id FROM categories WHERE name = 'Accessories' AND parent_id = (SELECT id FROM categories WHERE name = 'Computer Accessories')
)
INSERT INTO categories (id, name, parent_id, level)
VALUES
    (uuid_generate_v4(), 'Logitech', (SELECT id FROM sub_accessories_id), 2),
    (uuid_generate_v4(), 'Kingsons', (SELECT id FROM sub_accessories_id), 2),
    (uuid_generate_v4(), 'Cables', (SELECT id FROM sub_accessories_id), 2),
    (uuid_generate_v4(), 'Keyboards and Mouse', (SELECT id FROM sub_accessories_id), 2);

-- Subcategories under 'Security Systems' (Level 1)
WITH security_id AS (
    SELECT id FROM categories WHERE name = 'Security Systems'
)
INSERT INTO categories (id, name, parent_id, level)
VALUES
    (uuid_generate_v4(), 'Hikvision', (SELECT id FROM security_id), 1),
    (uuid_generate_v4(), 'Dahua', (SELECT id FROM security_id), 1),
    (uuid_generate_v4(), 'Zkteco', (SELECT id FROM security_id), 1);

-- Subcategories under 'Hikvision' (Level 2)
WITH hikvision_id AS (
    SELECT id FROM categories WHERE name = 'Hikvision'
)
INSERT INTO categories (id, name, parent_id, level)
VALUES
    (uuid_generate_v4(), 'Analog Cameras', (SELECT id FROM hikvision_id), 2),
    (uuid_generate_v4(), 'IP & PTZ Cameras', (SELECT id FROM hikvision_id), 2),
    (uuid_generate_v4(), 'NVR and DVR', (SELECT id FROM hikvision_id), 2),
    (uuid_generate_v4(), 'Surveillance Hard Disks', (SELECT id FROM hikvision_id), 2);

-- Subcategories under 'Zkteco' (Level 2)
WITH zkteco_id AS (
    SELECT id FROM categories WHERE name = 'Zkteco'
)
INSERT INTO categories (id, name, parent_id, level)
VALUES
    (uuid_generate_v4(), 'Access Control', (SELECT id FROM zkteco_id), 2);

-- Subcategories under 'Made in Kenya' (Level 1)
WITH kenya_id AS (
    SELECT id FROM categories WHERE name = 'Made in Kenya'
)
INSERT INTO categories (id, name, parent_id, level)
VALUES
    (uuid_generate_v4(), 'Zuriby KLI', (SELECT id FROM kenya_id), 1);

-- Subcategories under 'Zuriby KLI' (Level 2)
WITH zuriby_id AS (
    SELECT id FROM categories WHERE name = 'Zuriby KLI'
)
INSERT INTO categories (id, name, parent_id, level)
VALUES
    (uuid_generate_v4(), 'Cocktail Coasters', (SELECT id FROM zuriby_id), 2),
    (uuid_generate_v4(), 'Wine & Spirit Holders', (SELECT id FROM zuriby_id), 2),
    (uuid_generate_v4(), 'Mobile Phone Holders', (SELECT id FROM zuriby_id), 2),
    (uuid_generate_v4(), 'Laptop Risers', (SELECT id FROM zuriby_id), 2),
    (uuid_generate_v4(), 'Tissue Box Covers', (SELECT id FROM zuriby_id), 2),
    (uuid_generate_v4(), 'Cup Covers', (SELECT id FROM zuriby_id), 2),
    (uuid_generate_v4(), 'Bookmarks', (SELECT id FROM zuriby_id), 2),
    (uuid_generate_v4(), 'Serviette Holders', (SELECT id FROM zuriby_id), 2),
    (uuid_generate_v4(), 'Wall Art', (SELECT id FROM zuriby_id), 2),
    (uuid_generate_v4(), 'Planters', (SELECT id FROM zuriby_id), 2);
