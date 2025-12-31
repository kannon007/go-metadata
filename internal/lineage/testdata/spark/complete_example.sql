-- ============================================
-- Spark SQL Complete Example
-- 包含: Source DDL, Sink DDL, 临时视图, 多个 INSERT INTO
-- ============================================

-- ============================================
-- Source Tables DDL (External Tables)
-- ============================================

-- 用户行为日志表 (Hive External Table - Parquet)
CREATE EXTERNAL TABLE IF NOT EXISTS ods.user_behavior_log (
    user_id BIGINT COMMENT '用户ID',
    session_id STRING COMMENT '会话ID',
    event_type STRING COMMENT '事件类型',
    page_url STRING COMMENT '页面URL',
    referrer_url STRING COMMENT '来源URL',
    device_type STRING COMMENT '设备类型',
    os_type STRING COMMENT '操作系统',
    browser STRING COMMENT '浏览器',
    ip_address STRING COMMENT 'IP地址',
    event_time TIMESTAMP COMMENT '事件时间'
)
PARTITIONED BY (dt STRING COMMENT '日期分区')
STORED AS PARQUET
LOCATION 'hdfs:///data/ods/user_behavior_log'
TBLPROPERTIES ('parquet.compression' = 'SNAPPY');


-- 用户维度表 (Hive Table)
CREATE TABLE IF NOT EXISTS dim.user_dim (
    user_id BIGINT COMMENT '用户ID',
    user_name STRING COMMENT '用户名',
    phone STRING COMMENT '手机号',
    email STRING COMMENT '邮箱',
    gender STRING COMMENT '性别',
    age INT COMMENT '年龄',
    city STRING COMMENT '城市',
    province STRING COMMENT '省份',
    register_time TIMESTAMP COMMENT '注册时间',
    user_level STRING COMMENT '用户等级',
    is_vip BOOLEAN COMMENT '是否VIP'
)
STORED AS ORC
TBLPROPERTIES ('orc.compress' = 'ZLIB');

-- 商品维度表 (Hive Table)
CREATE TABLE IF NOT EXISTS dim.product_dim (
    product_id BIGINT COMMENT '商品ID',
    product_name STRING COMMENT '商品名称',
    category_id BIGINT COMMENT '类目ID',
    category_name STRING COMMENT '类目名称',
    brand_id BIGINT COMMENT '品牌ID',
    brand_name STRING COMMENT '品牌名称',
    price DECIMAL(10, 2) COMMENT '价格',
    cost DECIMAL(10, 2) COMMENT '成本',
    create_time TIMESTAMP COMMENT '创建时间'
)
STORED AS ORC;


-- 订单事实表 (Hive External Table)
CREATE EXTERNAL TABLE IF NOT EXISTS ods.order_fact (
    order_id BIGINT COMMENT '订单ID',
    user_id BIGINT COMMENT '用户ID',
    product_id BIGINT COMMENT '商品ID',
    quantity INT COMMENT '数量',
    unit_price DECIMAL(10, 2) COMMENT '单价',
    total_amount DECIMAL(10, 2) COMMENT '总金额',
    discount_amount DECIMAL(10, 2) COMMENT '优惠金额',
    pay_amount DECIMAL(10, 2) COMMENT '实付金额',
    order_status STRING COMMENT '订单状态',
    pay_type STRING COMMENT '支付方式',
    order_time TIMESTAMP COMMENT '下单时间',
    pay_time TIMESTAMP COMMENT '支付时间'
)
PARTITIONED BY (dt STRING)
STORED AS PARQUET
LOCATION 'hdfs:///data/ods/order_fact';

-- ============================================
-- Sink Tables DDL (Result Tables)
-- ============================================

-- 用户行为汇总表 (DWS层)
CREATE TABLE IF NOT EXISTS dws.user_behavior_daily (
    user_id BIGINT COMMENT '用户ID',
    user_name STRING COMMENT '用户名',
    city STRING COMMENT '城市',
    pv_count BIGINT COMMENT '页面浏览数',
    uv_count BIGINT COMMENT '独立页面数',
    session_count BIGINT COMMENT '会话数',
    avg_session_duration DOUBLE COMMENT '平均会话时长',
    first_visit_time TIMESTAMP COMMENT '首次访问时间',
    last_visit_time TIMESTAMP COMMENT '最后访问时间',
    dt STRING COMMENT '日期'
)
PARTITIONED BY (dt)
STORED AS ORC;


-- 用户订单汇总表 (DWS层)
CREATE TABLE IF NOT EXISTS dws.user_order_daily (
    user_id BIGINT COMMENT '用户ID',
    user_name STRING COMMENT '用户名',
    user_level STRING COMMENT '用户等级',
    order_count BIGINT COMMENT '订单数',
    total_amount DECIMAL(18, 2) COMMENT '订单总额',
    pay_amount DECIMAL(18, 2) COMMENT '实付总额',
    avg_order_amount DECIMAL(10, 2) COMMENT '平均订单金额',
    max_order_amount DECIMAL(10, 2) COMMENT '最大订单金额',
    dt STRING COMMENT '日期'
)
PARTITIONED BY (dt)
STORED AS ORC;

-- 商品销售汇总表 (DWS层)
CREATE TABLE IF NOT EXISTS dws.product_sales_daily (
    product_id BIGINT COMMENT '商品ID',
    product_name STRING COMMENT '商品名称',
    category_name STRING COMMENT '类目名称',
    brand_name STRING COMMENT '品牌名称',
    sale_quantity BIGINT COMMENT '销售数量',
    sale_amount DECIMAL(18, 2) COMMENT '销售金额',
    order_count BIGINT COMMENT '订单数',
    buyer_count BIGINT COMMENT '购买人数',
    dt STRING COMMENT '日期'
)
PARTITIONED BY (dt)
STORED AS ORC;


-- 用户RFM分析表 (ADS层)
CREATE TABLE IF NOT EXISTS ads.user_rfm_analysis (
    user_id BIGINT COMMENT '用户ID',
    user_name STRING COMMENT '用户名',
    recency_days INT COMMENT '最近购买天数',
    frequency INT COMMENT '购买频次',
    monetary DECIMAL(18, 2) COMMENT '消费金额',
    r_score INT COMMENT 'R评分',
    f_score INT COMMENT 'F评分',
    m_score INT COMMENT 'M评分',
    rfm_score INT COMMENT 'RFM总分',
    user_segment STRING COMMENT '用户分群',
    stat_date STRING COMMENT '统计日期'
)
STORED AS ORC;

-- 销售排行榜 (ADS层)
CREATE TABLE IF NOT EXISTS ads.sales_ranking (
    rank_type STRING COMMENT '排行类型',
    rank_id BIGINT COMMENT '排行对象ID',
    rank_name STRING COMMENT '排行对象名称',
    rank_value DECIMAL(18, 2) COMMENT '排行值',
    rank_num INT COMMENT '排名',
    stat_date STRING COMMENT '统计日期'
)
STORED AS ORC;

-- ============================================
-- Temporary Views (中间转换视图)
-- ============================================

-- 临时视图1: 用户行为明细 (关联用户维度)
CREATE TEMPORARY VIEW tmp_user_behavior_detail AS
SELECT 
    b.user_id,
    u.user_name,
    u.city,
    u.province,
    u.user_level,
    b.session_id,
    b.event_type,
    b.page_url,
    b.device_type,
    b.event_time,
    b.dt
FROM ods.user_behavior_log b
LEFT JOIN dim.user_dim u ON b.user_id = u.user_id
WHERE b.dt = '${bizdate}';


-- 临时视图2: 订单明细 (关联用户和商品维度)
CREATE TEMPORARY VIEW tmp_order_detail AS
SELECT 
    o.order_id,
    o.user_id,
    u.user_name,
    u.city,
    u.user_level,
    u.is_vip,
    o.product_id,
    p.product_name,
    p.category_name,
    p.brand_name,
    o.quantity,
    o.unit_price,
    o.total_amount,
    o.discount_amount,
    o.pay_amount,
    o.order_status,
    o.pay_type,
    o.order_time,
    o.pay_time,
    o.dt
FROM ods.order_fact o
LEFT JOIN dim.user_dim u ON o.user_id = u.user_id
LEFT JOIN dim.product_dim p ON o.product_id = p.product_id
WHERE o.dt = '${bizdate}' AND o.order_status = 'PAID';

-- 临时视图3: 用户行为聚合
CREATE TEMPORARY VIEW tmp_user_behavior_agg AS
SELECT 
    user_id,
    user_name,
    city,
    COUNT(*) AS pv_count,
    COUNT(DISTINCT page_url) AS uv_count,
    COUNT(DISTINCT session_id) AS session_count,
    MIN(event_time) AS first_visit_time,
    MAX(event_time) AS last_visit_time,
    dt
FROM tmp_user_behavior_detail
GROUP BY user_id, user_name, city, dt;


-- 临时视图4: 用户订单聚合
CREATE TEMPORARY VIEW tmp_user_order_agg AS
SELECT 
    user_id,
    user_name,
    user_level,
    COUNT(*) AS order_count,
    SUM(total_amount) AS total_amount,
    SUM(pay_amount) AS pay_amount,
    AVG(pay_amount) AS avg_order_amount,
    MAX(pay_amount) AS max_order_amount,
    dt
FROM tmp_order_detail
GROUP BY user_id, user_name, user_level, dt;

-- 临时视图5: 商品销售聚合
CREATE TEMPORARY VIEW tmp_product_sales_agg AS
SELECT 
    product_id,
    product_name,
    category_name,
    brand_name,
    SUM(quantity) AS sale_quantity,
    SUM(pay_amount) AS sale_amount,
    COUNT(DISTINCT order_id) AS order_count,
    COUNT(DISTINCT user_id) AS buyer_count,
    dt
FROM tmp_order_detail
GROUP BY product_id, product_name, category_name, brand_name, dt;

-- 临时视图6: 用户RFM计算
CREATE TEMPORARY VIEW tmp_user_rfm AS
SELECT 
    user_id,
    user_name,
    DATEDIFF('${bizdate}', MAX(DATE(order_time))) AS recency_days,
    COUNT(DISTINCT order_id) AS frequency,
    SUM(pay_amount) AS monetary
FROM tmp_order_detail
GROUP BY user_id, user_name;


-- 临时视图7: RFM评分
CREATE TEMPORARY VIEW tmp_user_rfm_scored AS
SELECT 
    user_id,
    user_name,
    recency_days,
    frequency,
    monetary,
    CASE 
        WHEN recency_days <= 7 THEN 5
        WHEN recency_days <= 14 THEN 4
        WHEN recency_days <= 30 THEN 3
        WHEN recency_days <= 60 THEN 2
        ELSE 1
    END AS r_score,
    CASE 
        WHEN frequency >= 10 THEN 5
        WHEN frequency >= 5 THEN 4
        WHEN frequency >= 3 THEN 3
        WHEN frequency >= 2 THEN 2
        ELSE 1
    END AS f_score,
    CASE 
        WHEN monetary >= 10000 THEN 5
        WHEN monetary >= 5000 THEN 4
        WHEN monetary >= 1000 THEN 3
        WHEN monetary >= 500 THEN 2
        ELSE 1
    END AS m_score
FROM tmp_user_rfm;

-- ============================================
-- INSERT INTO Statements (多个写入)
-- ============================================

-- INSERT 1: 写入用户行为汇总表
INSERT OVERWRITE TABLE dws.user_behavior_daily PARTITION (dt = '${bizdate}')
SELECT 
    user_id,
    user_name,
    city,
    pv_count,
    uv_count,
    session_count,
    CAST(NULL AS DOUBLE) AS avg_session_duration,
    first_visit_time,
    last_visit_time
FROM tmp_user_behavior_agg;


-- INSERT 2: 写入用户订单汇总表
INSERT OVERWRITE TABLE dws.user_order_daily PARTITION (dt = '${bizdate}')
SELECT 
    user_id,
    user_name,
    user_level,
    order_count,
    total_amount,
    pay_amount,
    avg_order_amount,
    max_order_amount
FROM tmp_user_order_agg;

-- INSERT 3: 写入商品销售汇总表
INSERT OVERWRITE TABLE dws.product_sales_daily PARTITION (dt = '${bizdate}')
SELECT 
    product_id,
    product_name,
    category_name,
    brand_name,
    sale_quantity,
    sale_amount,
    order_count,
    buyer_count
FROM tmp_product_sales_agg;

-- INSERT 4: 写入用户RFM分析表
INSERT OVERWRITE TABLE ads.user_rfm_analysis
SELECT 
    user_id,
    user_name,
    recency_days,
    frequency,
    monetary,
    r_score,
    f_score,
    m_score,
    r_score + f_score + m_score AS rfm_score,
    CASE 
        WHEN r_score >= 4 AND f_score >= 4 AND m_score >= 4 THEN '高价值用户'
        WHEN r_score >= 4 AND f_score >= 4 THEN '重要保持用户'
        WHEN r_score >= 4 AND m_score >= 4 THEN '重要发展用户'
        WHEN f_score >= 4 AND m_score >= 4 THEN '重要挽留用户'
        WHEN r_score >= 4 THEN '新用户'
        WHEN f_score >= 4 THEN '一般保持用户'
        WHEN m_score >= 4 THEN '一般发展用户'
        ELSE '流失用户'
    END AS user_segment,
    '${bizdate}' AS stat_date
FROM tmp_user_rfm_scored;


-- INSERT 5: 写入商品销售排行榜
INSERT INTO TABLE ads.sales_ranking
SELECT 
    'PRODUCT' AS rank_type,
    product_id AS rank_id,
    product_name AS rank_name,
    sale_amount AS rank_value,
    ROW_NUMBER() OVER (ORDER BY sale_amount DESC) AS rank_num,
    '${bizdate}' AS stat_date
FROM tmp_product_sales_agg
ORDER BY sale_amount DESC
LIMIT 100;

-- INSERT 6: 写入品牌销售排行榜
INSERT INTO TABLE ads.sales_ranking
SELECT 
    'BRAND' AS rank_type,
    CAST(ROW_NUMBER() OVER (ORDER BY total_sales DESC) AS BIGINT) AS rank_id,
    brand_name AS rank_name,
    total_sales AS rank_value,
    ROW_NUMBER() OVER (ORDER BY total_sales DESC) AS rank_num,
    '${bizdate}' AS stat_date
FROM (
    SELECT 
        brand_name,
        SUM(sale_amount) AS total_sales
    FROM tmp_product_sales_agg
    GROUP BY brand_name
) brand_sales
ORDER BY total_sales DESC
LIMIT 50;

-- INSERT 7: 写入类目销售排行榜
INSERT INTO TABLE ads.sales_ranking
SELECT 
    'CATEGORY' AS rank_type,
    CAST(ROW_NUMBER() OVER (ORDER BY total_sales DESC) AS BIGINT) AS rank_id,
    category_name AS rank_name,
    total_sales AS rank_value,
    ROW_NUMBER() OVER (ORDER BY total_sales DESC) AS rank_num,
    '${bizdate}' AS stat_date
FROM (
    SELECT 
        category_name,
        SUM(sale_amount) AS total_sales
    FROM tmp_product_sales_agg
    GROUP BY category_name
) category_sales
ORDER BY total_sales DESC
LIMIT 50;


-- ============================================
-- Expected Lineage (血缘关系)
-- ============================================
--
-- dws.user_behavior_daily:
--   user_id <- ods.user_behavior_log.user_id
--   user_name <- dim.user_dim.user_name
--   city <- dim.user_dim.city
--   pv_count <- ods.user_behavior_log.* (COUNT)
--   uv_count <- ods.user_behavior_log.page_url (COUNT DISTINCT)
--   session_count <- ods.user_behavior_log.session_id (COUNT DISTINCT)
--   first_visit_time <- ods.user_behavior_log.event_time (MIN)
--   last_visit_time <- ods.user_behavior_log.event_time (MAX)
--
-- dws.user_order_daily:
--   user_id <- ods.order_fact.user_id
--   user_name <- dim.user_dim.user_name
--   user_level <- dim.user_dim.user_level
--   order_count <- ods.order_fact.* (COUNT)
--   total_amount <- ods.order_fact.total_amount (SUM)
--   pay_amount <- ods.order_fact.pay_amount (SUM)
--   avg_order_amount <- ods.order_fact.pay_amount (AVG)
--   max_order_amount <- ods.order_fact.pay_amount (MAX)
--
-- dws.product_sales_daily:
--   product_id <- ods.order_fact.product_id
--   product_name <- dim.product_dim.product_name
--   category_name <- dim.product_dim.category_name
--   brand_name <- dim.product_dim.brand_name
--   sale_quantity <- ods.order_fact.quantity (SUM)
--   sale_amount <- ods.order_fact.pay_amount (SUM)
--   order_count <- ods.order_fact.order_id (COUNT DISTINCT)
--   buyer_count <- ods.order_fact.user_id (COUNT DISTINCT)
--
-- ads.user_rfm_analysis:
--   user_id <- ods.order_fact.user_id
--   user_name <- dim.user_dim.user_name
--   recency_days <- ods.order_fact.order_time (DATEDIFF, MAX)
--   frequency <- ods.order_fact.order_id (COUNT DISTINCT)
--   monetary <- ods.order_fact.pay_amount (SUM)
--   r_score <- recency_days (CASE)
--   f_score <- frequency (CASE)
--   m_score <- monetary (CASE)
--   rfm_score <- r_score + f_score + m_score
--   user_segment <- rfm_score (CASE)
--
-- ads.sales_ranking (PRODUCT):
--   rank_type <- literal
--   rank_id <- ods.order_fact.product_id
--   rank_name <- dim.product_dim.product_name
--   rank_value <- ods.order_fact.pay_amount (SUM)
--   rank_num <- (ROW_NUMBER window function)
