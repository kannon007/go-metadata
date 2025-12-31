-- Test: INSERT...SELECT
-- Input SQL
INSERT INTO report(total_amount, user_count)
SELECT SUM(amount), COUNT(DISTINCT user_id) FROM orders;

-- Expected lineage:
-- target: report.total_amount <- sources: [orders.amount], operators: [sum]
-- target: report.user_count <- sources: [orders.user_id], operators: [count, distinct]
