-- 1. 精简后的用户表（仅保留必要字段）
CREATE TABLE IF NOT EXISTS users(
                                    id INT PRIMARY KEY AUTO_INCREMENT,
                                    username VARCHAR(50) NOT NULL UNIQUE,  -- 用户名（唯一）
    password VARCHAR(100) NOT NULL,       -- 密码（建议存储加密后的值）
    email VARCHAR(100) NOT NULL UNIQUE    -- 邮箱（唯一）
    );

-- 插入用户测试数据
INSERT IGNORE INTO users (username, password, email) VALUES
('user1', 'password123', 'user1@example.com'),
('user2', 'password456', 'user2@example.com'),
('user3', 'password789', 'user3@example.com');

-- 2. 图书表
CREATE TABLE IF NOT EXISTS books(
                                    id INT PRIMARY KEY AUTO_INCREMENT,
                                    title VARCHAR(100) NOT NULL,
    author VARCHAR(100) NOT NULL,
    price DOUBLE(11,2) NOT NULL,
    sales INT NOT NULL,
    stock INT NOT NULL,
    img_path VARCHAR(100)
    );

-- 插入图书测试数据
INSERT IGNORE INTO books (title, author, price, sales, stock, img_path) VALUES
('Go编程实战', 'John Smith', 89.00, 2, 100, '/images/goprogramming.jpg'),
('Python数据分析', 'Jane Doe', 79.00, 2, 80, '/images/pythondata.jpg'),
('MySQL从入门到精通', 'Michael Brown', 69.00, 2, 120, '/images/mysqlguide.jpg'),
('JavaScript高级程序设计', 'David Wilson', 99.00, 2, 70, '/images/jsadv.jpg');

-- 3. 会话表（依赖users表）
CREATE TABLE IF NOT EXISTS sessions(
                                       session_id VARCHAR(100) PRIMARY KEY,
    username VARCHAR(100) NOT NULL,
    user_id INT NOT NULL,
    FOREIGN KEY(user_id) REFERENCES users(id)
    );

-- 插入会话测试数据
INSERT IGNORE INTO sessions (session_id, username, user_id) VALUES
('sess_123456', 'user1', 1),
('sess_789012', 'user2', 2);

-- 4. 购物车表（依赖users表）
CREATE TABLE IF NOT EXISTS carts(
                                    id VARCHAR(100) PRIMARY KEY,
    total_count INT NOT NULL,
    total_amount DOUBLE(11,2) NOT NULL,
    user_id INT NOT NULL,
    FOREIGN KEY(user_id) REFERENCES users(id)
    );

-- 插入购物车测试数据
INSERT IGNORE INTO carts (id, total_count, total_amount, user_id) VALUES
('cart_1001', 2, 168.00, 1),
('cart_1002', 1, 99.00, 2),
('cart_1003', 0, 0.00, 3);

-- 5. 购物项表（依赖books表和carts表）
CREATE TABLE IF NOT EXISTS cart_items(
                                         id INT PRIMARY KEY AUTO_INCREMENT,
                                         count INT NOT NULL,
                                         amount DOUBLE(11,2) NOT NULL,
    book_id INT NOT NULL,
    cart_id VARCHAR(100) NOT NULL,
    FOREIGN KEY(book_id) REFERENCES books(id),
    FOREIGN KEY(cart_id) REFERENCES carts(id)
    );

-- 插入购物项测试数据
INSERT IGNORE INTO cart_items (count, amount, book_id, cart_id) VALUES
(1, 89.00, 1, 'cart_1001'),
(1, 79.00, 2, 'cart_1001'),
(1, 99.00, 4, 'cart_1002');

-- 6. 订单表（依赖users表）
CREATE TABLE IF NOT EXISTS orders(
                                     id VARCHAR(100) PRIMARY KEY,
    total_count INT NOT NULL,
    total_amount DOUBLE(11,2) NOT NULL,
    state INT NOT NULL,
    user_id INT,
    FOREIGN KEY(user_id) REFERENCES users(id)
    );

-- 插入订单测试数据（state: 0-待付款, 1-已付款, 2-已发货, 3-已完成）
INSERT IGNORE INTO orders (id, total_count, total_amount, state, user_id) VALUES
('order_2001', 2, 168.00, 1, 1),
('order_2002', 1, 69.00, 3, 1),
('order_2003', 1, 99.00, 0, 2);

-- 7. 订单项表（依赖orders表）
CREATE TABLE IF NOT EXISTS order_items(
                                          id INT PRIMARY KEY AUTO_INCREMENT,
                                          count INT NOT NULL,
                                          amount DOUBLE(11,2) NOT NULL,
    title VARCHAR(100) NOT NULL,
    author VARCHAR(100) NOT NULL,
    price DOUBLE(11,2) NOT NULL,
    img_path VARCHAR(100) NOT NULL,
    order_id VARCHAR(100) NOT NULL,
    FOREIGN KEY(order_id) REFERENCES orders(id)
    );

-- 插入订单项测试数据
INSERT IGNORE INTO order_items (count, amount, title, author, price, img_path, order_id) VALUES
(1, 89.00, 'Go编程实战', 'John Smith', 89.00, '/images/goprogramming.jpg', 'order_2001'),
(1, 79.00, 'Python数据分析', 'Jane Doe', 79.00, '/images/pythondata.jpg', 'order_2001'),
(1, 69.00, 'MySQL从入门到精通', 'Michael Brown', 69.00, '/images/mysqlguide.jpg', 'order_2002'),
(1, 99.00, 'JavaScript高级程序设计', 'David Wilson', 99.00, '/images/jsadv.jpg', 'order_2003');
