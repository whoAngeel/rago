CREATE TABLE roles (
    id SERIAL PRIMARY KEY,
    name varchar(50) UNIQUE NOT NULL,
    description TEXT, 
    created_at TIMESTAMP DEFAULT NOW()
);

INSERT INTO roles (name, description) VALUES 
('admin', 'Can do everything'),
('editor', 'Can ask and ingest'),
('viewer', 'Can only ask');