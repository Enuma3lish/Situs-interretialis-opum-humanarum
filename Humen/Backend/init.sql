CREATE TABLE company (
    id   SERIAL PRIMARY KEY,
    name VARCHAR(128) NOT NULL
);

CREATE TABLE job (
    id         SERIAL PRIMARY KEY,
    company_id INTEGER NOT NULL REFERENCES company(id),
    title      VARCHAR(128) NOT NULL,
    salary_min INTEGER NOT NULL,
    salary_max INTEGER NOT NULL
);

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(64) UNIQUE NOT NULL,
    password_hash VARCHAR(128) NOT NULL,
    is_admin BOOLEAN NOT NULL DEFAULT FALSE
);

-- 密碼 root，這裡請貼上你剛產生的 hash
INSERT INTO users (username, password_hash, is_admin) VALUES
('admin', '$2a$10$kYVrMYYOxI36/1H4vSxQ1OFFSjJwLn9gaNU6LhElXOdj2.gzc7k9.', TRUE);

-- 插入公司
INSERT INTO company (name) VALUES
('台積電'),
('鴻海'),
('聯發科'),
('華碩'),
('宏碁');

-- 插入職缺 ...（原本內容不用變）
INSERT INTO job (company_id, title, salary_min, salary_max) VALUES
(1, '製程工程師', 80000, 120000),
(1, '技術員', 30000, 36000),
(1, '自動化工程師', 65000, 90000),
(1, '設備工程師', 70000, 110000),
(1, '工安工程師', 50000, 65000),
(1, '品保工程師', 70000, 105000),
(2, '工廠技術員', 34000, 40000),
(2, '品保工程師', 53000, 70000),
(2, '自動化工程師', 60000, 85000),
(2, '研發工程師', 75000, 95000),
(2, '製造工程師', 90000, 120000),
(3, 'IC設計工程師', 130000, 200000),
(3, '測試工程師', 70000, 95000),
(3, '韌體工程師', 75000, 130000),
(4, '軟體工程師', 75000, 120000),
(4, '硬體工程師', 60000, 95000),
(4, '產品經理', 85000, 130000),
(5, '資安工程師', 80000, 120000),
(5, '系統維運', 50000, 85000),
(5, '產品企劃', 70000, 100000);
