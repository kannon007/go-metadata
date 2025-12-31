-- Test: Flink Tumble Window
-- Input SQL
SELECT 
    user_id,
    TUMBLE_START(event_time, INTERVAL '1' HOUR) as window_start,
    COUNT(*) as event_count
FROM events
GROUP BY user_id, TUMBLE(event_time, INTERVAL '1' HOUR);

-- Expected lineage:
-- target: user_id <- sources: [events.user_id], operators: [identity]
-- target: window_start <- sources: [events.event_time], operators: [tumble_start]
-- target: event_count <- sources: [events.*], operators: [count]
