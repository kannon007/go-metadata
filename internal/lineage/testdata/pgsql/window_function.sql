-- Test: Window Function
-- Input SQL
SELECT 
    id,
    name,
    ROW_NUMBER() OVER (PARTITION BY department ORDER BY salary DESC) as rank
FROM employees;

-- Expected lineage:
-- target: id <- sources: [employees.id], operators: [identity]
-- target: name <- sources: [employees.name], operators: [identity]
-- target: rank <- sources: [employees.department, employees.salary], operators: [row_number, over]
