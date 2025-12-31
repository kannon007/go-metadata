-- ============================================
-- Flink SQL Complete Example
-- 包含: Source DDL, Sink DDL, 临时视图, 多个 INSERT INTO
-- ============================================

-- ============================================
-- Source Tables DDL
-- ============================================

-- 用户行为事件表 (Kafka Source)
CREATE TABLE user_events (
    user_id BIGINT,
    event_type STRING,
    event_time TIMESTAMP(3),
    page_url STRING,
    device_type STRING,
    WATERMARK FOR event_time AS event_time - INTERVAL '5' SECOND
) WITH (
    'connector' = 'kafka',
    'topic' = 'user_events',
    'properties.bootstrap.servers' = 'localhost:9092',
    'format' = 'json'
);

-- 用户信息表 (MySQL Source)
CREATE TABLE user_info (
    user_id BIGINT,
    user_name STRING,
    age INT,
    gender STRING,
    city STRING,
    register_time TIMESTAMP(3),
    PRIMARY KEY (user_id) NOT ENFORCED
) WITH (
    'connector' = 'jdbc',
    'url' = 'jdbc:mysql://localhost:3306/user_db',
    'table-name' = 'user_info'
);

-- 商品信息表 (MySQL Source)
CREATE TABLE product_info (
    product_id BIGINT,
    product_name STRING,
    category STRING,
    price DECIMAL(10, 2),
    PRIMARY KEY (product_id) NOT ENFORCED
) WITH (
    'connector' = 'jdbc',
    'url' = 'jdbc:mysql://localhost:3306/product_db',
    'table-name' = 'product_info'
);


-- 订单表 (Kafka Source)
CREATE TABLE orders (
    order_id BIGINT,
    user_id BIGINT,
    product_id BIGINT,
    quantity INT,
    amount DECIMAL(10, 2),
    order_time TIMESTAMP(3),
    WATERMARK FOR order_time AS order_time - INTERVAL '10' SECOND
) WITH (
    'connector' = 'kafka',
    'topic' = 'orders',
    'properties.bootstrap.servers' = 'localhost:9092',
    'format' = 'json'
);

-- ============================================
-- Sink Tables DDL
-- ============================================

-- 用户行为统计结果表 (Kafka Sink)
CREATE TABLE user_behavior_stats (
    user_id BIGINT,
    window_start TIMESTAMP(3),
    window_end TIMESTAMP(3),
    pv_count BIGINT,
    uv_count BIGINT,
    PRIMARY KEY (user_id, window_start) NOT ENFORCED
) WITH (
    'connector' = 'upsert-kafka',
    'topic' = 'user_behavior_stats',
    'properties.bootstrap.servers' = 'localhost:9092',
    'key.format' = 'json',
    'value.format' = 'json'
);

-- 用户订单汇总表 (MySQL Sink)
CREATE TABLE user_order_summary (
    user_id BIGINT,
    user_name STRING,
    total_orders BIGINT,
    total_amount DECIMAL(10, 2),
    avg_amount DECIMAL(10, 2),
    last_order_time TIMESTAMP(3),
    PRIMARY KEY (user_id) NOT ENFORCED
) WITH (
    'connector' = 'jdbc',
    'url' = 'jdbc:mysql://localhost:3306/report_db',
    'table-name' = 'user_order_summary'
);


-- 商品销售排行表 (Elasticsearch Sink)
CREATE TABLE product_sales_rank (
    product_id BIGINT,
    product_name STRING,
    category STRING,
    total_quantity BIGINT,
    total_sales DECIMAL(10, 2),
    rank_num BIGINT,
    stat_date STRING,
    PRIMARY KEY (product_id, stat_date) NOT ENFORCED
) WITH (
    'connector' = 'elasticsearch-7',
    'hosts' = 'http://localhost:9200',
    'index' = 'product_sales_rank'
);

-- 实时告警表 (Kafka Sink)
CREATE TABLE realtime_alerts (
    alert_id STRING,
    alert_type STRING,
    user_id BIGINT,
    alert_message STRING,
    alert_time TIMESTAMP(3)
) WITH (
    'connector' = 'kafka',
    'topic' = 'realtime_alerts',
    'properties.bootstrap.servers' = 'localhost:9092',
    'format' = 'json'
);

-- ============================================
-- Temporary Views (中间转换视图)
-- ============================================

-- 临时视图1: 用户行为明细 (关联用户信息)
CREATE TEMPORARY VIEW user_behavior_detail AS
SELECT 
    e.user_id,
    u.user_name,
    u.city,
    e.event_type,
    e.page_url,
    e.device_type,
    e.event_time
FROM user_events e
LEFT JOIN user_info FOR SYSTEM_TIME AS OF e.event_time AS u
    ON e.user_id = u.user_id;


-- 临时视图2: 订单明细 (关联用户和商品信息)
CREATE TEMPORARY VIEW order_detail AS
SELECT 
    o.order_id,
    o.user_id,
    u.user_name,
    u.city,
    o.product_id,
    p.product_name,
    p.category,
    p.price AS unit_price,
    o.quantity,
    o.amount,
    o.order_time
FROM orders o
LEFT JOIN user_info FOR SYSTEM_TIME AS OF o.order_time AS u
    ON o.user_id = u.user_id
LEFT JOIN product_info FOR SYSTEM_TIME AS OF o.order_time AS p
    ON o.product_id = p.product_id;

-- 临时视图3: 窗口聚合 - 用户行为统计
CREATE TEMPORARY VIEW user_behavior_window AS
SELECT 
    user_id,
    TUMBLE_START(event_time, INTERVAL '1' HOUR) AS window_start,
    TUMBLE_END(event_time, INTERVAL '1' HOUR) AS window_end,
    COUNT(*) AS pv_count,
    COUNT(DISTINCT page_url) AS uv_count
FROM user_events
GROUP BY user_id, TUMBLE(event_time, INTERVAL '1' HOUR);

-- 临时视图4: 用户订单汇总
CREATE TEMPORARY VIEW user_order_agg AS
SELECT 
    user_id,
    user_name,
    COUNT(*) AS total_orders,
    SUM(amount) AS total_amount,
    AVG(amount) AS avg_amount,
    MAX(order_time) AS last_order_time
FROM order_detail
GROUP BY user_id, user_name;


-- 临时视图5: 商品销售统计
CREATE TEMPORARY VIEW product_sales_agg AS
SELECT 
    product_id,
    product_name,
    category,
    SUM(quantity) AS total_quantity,
    SUM(amount) AS total_sales,
    DATE_FORMAT(order_time, 'yyyy-MM-dd') AS stat_date
FROM order_detail
GROUP BY product_id, product_name, category, DATE_FORMAT(order_time, 'yyyy-MM-dd');

-- ============================================
-- INSERT INTO Statements (多个写入)
-- ============================================

-- INSERT 1: 写入用户行为统计
INSERT INTO user_behavior_stats
SELECT 
    user_id,
    window_start,
    window_end,
    pv_count,
    uv_count
FROM user_behavior_window;

-- INSERT 2: 写入用户订单汇总
INSERT INTO user_order_summary
SELECT 
    user_id,
    user_name,
    total_orders,
    total_amount,
    avg_amount,
    last_order_time
FROM user_order_agg;


-- INSERT 3: 写入商品销售排行 (带窗口排名)
INSERT INTO product_sales_rank
SELECT 
    product_id,
    product_name,
    category,
    total_quantity,
    total_sales,
    ROW_NUMBER() OVER (PARTITION BY stat_date ORDER BY total_sales DESC) AS rank_num,
    stat_date
FROM product_sales_agg;

-- INSERT 4: 写入实时告警 (大额订单告警)
INSERT INTO realtime_alerts
SELECT 
    CONCAT('ALERT_', CAST(order_id AS STRING)) AS alert_id,
    'HIGH_VALUE_ORDER' AS alert_type,
    user_id,
    CONCAT('用户 ', user_name, ' 下单金额 ', CAST(amount AS STRING), ' 元，超过阈值') AS alert_message,
    order_time AS alert_time
FROM order_detail
WHERE amount > 10000;

-- INSERT 5: 写入实时告警 (异常行为告警)
INSERT INTO realtime_alerts
SELECT 
    CONCAT('ALERT_', CAST(user_id AS STRING), '_', CAST(window_start AS STRING)) AS alert_id,
    'ABNORMAL_BEHAVIOR' AS alert_type,
    user_id,
    CONCAT('用户 ', CAST(user_id AS STRING), ' 在1小时内访问 ', CAST(pv_count AS STRING), ' 次，疑似异常') AS alert_message,
    window_end AS alert_time
FROM user_behavior_window
WHERE pv_count > 1000;


-- ============================================
-- Expected Lineage (血缘关系)
-- ============================================
-- 
-- user_behavior_stats:
--   user_id <- user_events.user_id
--   window_start <- user_events.event_time (TUMBLE_START)
--   window_end <- user_events.event_time (TUMBLE_END)
--   pv_count <- user_events.* (COUNT)
--   uv_count <- user_events.page_url (COUNT DISTINCT)
--
-- user_order_summary:
--   user_id <- orders.user_id
--   user_name <- user_info.user_name
--   total_orders <- orders.* (COUNT)
--   total_amount <- orders.amount (SUM)
--   avg_amount <- orders.amount (AVG)
--   last_order_time <- orders.order_time (MAX)
--
-- product_sales_rank:
--   product_id <- orders.product_id
--   product_name <- product_info.product_name
--   category <- product_info.category
--   total_quantity <- orders.quantity (SUM)
--   total_sales <- orders.amount (SUM)
--   rank_num <- (ROW_NUMBER window function)
--   stat_date <- orders.order_time (DATE_FORMAT)
--
-- realtime_alerts (HIGH_VALUE_ORDER):
--   alert_id <- orders.order_id (CONCAT)
--   alert_type <- literal
--   user_id <- orders.user_id
--   alert_message <- user_info.user_name, orders.amount (CONCAT)
--   alert_time <- orders.order_time
--
-- realtime_alerts (ABNORMAL_BEHAVIOR):
--   alert_id <- user_events.user_id, user_events.event_time (CONCAT)
--   alert_type <- literal
--   user_id <- user_events.user_id
--   alert_message <- user_events.user_id, COUNT(*) (CONCAT)
--   alert_time <- user_events.event_time (TUMBLE_END)
