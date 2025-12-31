-- Test: SQL Server TOP clause
-- Input SQL
SELECT TOP 10 id, name, email
FROM users
ORDER BY created_at DESC;

-- Expected lineage:
-- target: id <- sources: [users.id], operators: [identity]
-- target: name <- sources: [users.name], operators: [identity]
-- target: email <- sources: [users.email], operators: [identity]
