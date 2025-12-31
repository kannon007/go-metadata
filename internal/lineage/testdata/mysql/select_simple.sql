-- Test: Simple SELECT
-- Input SQL
SELECT id, name FROM users;

-- Expected lineage:
-- target: id <- sources: [users.id], operators: [identity]
-- target: name <- sources: [users.name], operators: [identity]
