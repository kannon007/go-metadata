-- Test: SELECT with JOIN
-- Input SQL
SELECT u.id, u.name, o.amount
FROM users u
INNER JOIN orders o ON u.id = o.user_id;

-- Expected lineage:
-- target: id <- sources: [users.id], operators: [identity]
-- target: name <- sources: [users.name], operators: [identity]
-- target: amount <- sources: [orders.amount], operators: [identity]
