-- Test: CTE (Common Table Expression)
-- Input SQL
WITH active_users AS (
    SELECT id, name FROM users WHERE status = 'active'
)
SELECT id, name FROM active_users;

-- Expected lineage:
-- target: id <- sources: [users.id], operators: [identity]
-- target: name <- sources: [users.name], operators: [identity]
