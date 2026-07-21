INSERT INTO tasks (title, assignee, status, created_at, updated_at) VALUES
('Design login page', 'alice', 'done', NOW() - INTERVAL '7 days', NOW() - INTERVAL '6 days'),
('Implement authentication API', 'bob', 'done', NOW() - INTERVAL '6 days', NOW() - INTERVAL '4 days'),
('Write unit tests for auth', 'alice', 'in_progress', NOW() - INTERVAL '4 days', NOW() - INTERVAL '2 days'),
('Set up CI/CD pipeline', 'charlie', 'pending', NOW() - INTERVAL '3 days', NOW() - INTERVAL '3 days'),
('Create database migrations', 'bob', 'done', NOW() - INTERVAL '5 days', NOW() - INTERVAL '5 days'),
('Build task CRUD endpoints', 'alice', 'in_progress', NOW() - INTERVAL '2 days', NOW() - INTERVAL '1 day'),
('Add Redis caching layer', 'charlie', 'pending', NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day'),
('Write API documentation', 'bob', 'pending', NOW(), NOW()),
('Set up monitoring with Prometheus', 'charlie', 'in_progress', NOW() - INTERVAL '1 day', NOW()),
('Deploy to staging environment', 'alice', 'pending', NOW(), NOW());
