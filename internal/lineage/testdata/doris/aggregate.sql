-- Test: Doris Aggregate
-- Input SQL
SELECT 
    date,
    channel,
    SUM(pv) as total_pv,
    BITMAP_UNION_COUNT(user_bitmap) as uv
FROM page_stats
GROUP BY date, channel;

-- Expected lineage:
-- target: date <- sources: [page_stats.date], operators: [identity]
-- target: channel <- sources: [page_stats.channel], operators: [identity]
-- target: total_pv <- sources: [page_stats.pv], operators: [sum]
-- target: uv <- sources: [page_stats.user_bitmap], operators: [bitmap_union_count]
