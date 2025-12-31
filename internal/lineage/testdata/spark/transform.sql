-- Test: Spark SQL Transform
-- Input SQL
SELECT 
    user_id,
    TRANSFORM(items, x -> x.price * x.quantity) as item_totals,
    AGGREGATE(items, 0, (acc, x) -> acc + x.price) as total
FROM orders;

-- Expected lineage:
-- target: user_id <- sources: [orders.user_id], operators: [identity]
-- target: item_totals <- sources: [orders.items], operators: [transform]
-- target: total <- sources: [orders.items], operators: [aggregate]
