-- Создание типа для статуса заказа
CREATE TYPE order_status_enum AS ENUM('active','closed');

-- Таблица orders с полем customer_name вместо customer_id
CREATE TABLE orders(
    order_id SERIAL PRIMARY KEY, 
    customer_name VARCHAR(255) NOT NULL,  -- добавлено поле для имени клиента
    order_date TIMESTAMPTZ DEFAULT NOW(),
    status order_status_enum NOT NULL,
    total_amount DECIMAL(10,2) NOT NULL CHECK(total_amount>0),
    special_instructions JSONB
);

-- Таблица menu_items
CREATE TABLE menu_items(
    menu_item_id SERIAL PRIMARY KEY,
    description TEXT NOT NULL,
    name VARCHAR(100) NOT NULL UNIQUE,
    price DECIMAL(10,2) NOT NULL CHECK(price>=0),
    tags TEXT[]
);

-- Таблица order_items
CREATE TABLE order_items(
    order_item_id SERIAL PRIMARY KEY,
    menu_item_id INT REFERENCES menu_items(menu_item_id) ON DELETE CASCADE,
    order_id INT REFERENCES orders(order_id) ON DELETE CASCADE,
    customizations JSONB,
    price_at_order_time DECIMAL(10,2) NOT NULL CHECK(price_at_order_time>0),
    quantity INT NOT NULL CHECK (quantity >0)
);

-- Таблица inventory
CREATE TABLE inventory(
    inventory_id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    stock_level DECIMAL(10,2) NOT NULL CHECK(stock_level>0),
    last_updated TIMESTAMPTZ DEFAULT NOW(),
    reorder_level DECIMAL(10,2) NOT NULL CHECK(reorder_level>0)
);

-- Таблица menu_item_ingredients
CREATE TABLE menu_item_ingredients(
    id SERIAL PRIMARY KEY,
    inventory_id INT REFERENCES inventory(inventory_id) ON DELETE CASCADE,
    menu_item_id INT REFERENCES menu_items(menu_item_id) ON DELETE CASCADE,
    quantity DECIMAL(10,2) NOT NULL CHECK(quantity>0)
);

-- Таблица inventory_transactions
CREATE TABLE inventory_transactions(
    transaction_id SERIAL PRIMARY KEY,
    inventory_id INT REFERENCES inventory(inventory_id) ON DELETE CASCADE,
    quantity DECIMAL(10,2) CHECK(quantity>=0),
    transaction_date TIMESTAMPTZ DEFAULT NOW()
);

-- Таблица order_status_history
CREATE TABLE order_status_history(
    id SERIAL PRIMARY KEY,
    order_id INT REFERENCES orders(order_id) ON DELETE CASCADE,
    status order_status_enum NOT NULL,
    changed_at TIMESTAMPTZ DEFAULT NOW()
);

-- Таблица price_history
CREATE TABLE price_history (
    id SERIAL PRIMARY KEY,
    menu_item_id INT REFERENCES menu_items(menu_item_id) ON DELETE CASCADE,
    old_price DECIMAL(10,2) CHECK(old_price>0),
    new_price DECIMAL(10,2) CHECK(new_price>0),
    changed_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_order_date ON orders(order_date);

CREATE INDEX idx_menu_items_name_ft ON menu_items USING GIN (to_tsvector('english', name));
CREATE INDEX idx_menu_items_description_ft ON menu_items USING GIN (to_tsvector('english', description));
CREATE INDEX idx_menu_items_tags ON menu_items USING GIN (tags);

CREATE INDEX idx_inventory_name ON inventory(name);
CREATE INDEX idx_inventory_stock_level ON inventory(stock_level);

CREATE INDEX idx_inventory_transactions_inventory_id ON inventory_transactions(inventory_id);
CREATE INDEX idx_inventory_transactions_date ON inventory_transactions(transaction_date);

CREATE INDEX idx_order_status_history_order_id ON order_status_history(order_id);
CREATE INDEX idx_order_status_history_composite ON order_status_history(order_id, changed_at);

CREATE INDEX idx_price_history_menu_item_id ON price_history(menu_item_id);
CREATE INDEX idx_price_history_changed_at ON price_history(changed_at);

-- Вставка данных в orders (теперь с customer_name)
INSERT INTO orders (customer_name, order_date, status, total_amount, special_instructions) VALUES
('John Doe', '2023-11-14 13:15:45', 'active', 17.49, '{"note": "Extra cheese and olives"}'),
('Alice Smith', '2024-02-05 09:02:11', 'closed', 9.25, '{"note": "No onions, add pickles"}'),
('Bob Johnson', '2022-08-07 18:35:50', 'active', 14.00, '{"note": "Gluten-free crust"}'),
('Pizza Davis', '2025-01-10 08:45:30', 'closed', 26.99, '{"note": "Medium rare, side salad"}'),
('John Brown', '2023-06-14 21:10:00', 'active', 6.50, '{"note": "Spicy, extra jalapenos"}'),
('Sophia Wilson', '2024-09-17 12:50:15', 'closed', 8.25, '{"note": "No mayo, extra mustard"}'),
('Daniel Martinez', '2021-03-19 07:30:40', 'active', 11.49, '{"note": "Extra sauce on the side"}'),
('Olivia Taylor', '2022-12-21 20:20:35', 'closed', 7.25, '{"note": "Vegan option, no nuts"}'),
('James Anderson', '2025-05-25 11:55:14', 'active', 4.75, '{"note": "Well-done, no salt"}'),
('Emma Thomas', '2023-11-14 23:15:05', 'closed', 5.50, '{"note": "With sprinkles and syrup"}');

-- Вставка данных в menu_items
INSERT INTO menu_items (name, description, price, tags) VALUES
('Pizza', 'Delicious cheese pizza', 12.99, ARRAY['cheese', 'fast-food']),
('Burger', 'Juicy beef burger', 8.99, ARRAY['beef', 'fast-food']),
('Pasta', 'Creamy Alfredo pasta', 10.99, ARRAY['pasta', 'Italian']),
('Salad', 'Fresh garden salad', 6.99, ARRAY['healthy', 'vegan']),
('Sushi', 'Traditional sushi rolls', 15.99, ARRAY['fish', 'Japanese']),
('Steak', 'Grilled ribeye steak', 24.99, ARRAY['meat', 'gourmet']),
('Soup', 'Hot chicken soup', 5.99, ARRAY['chicken', 'starter']),
('Fries', 'Crispy French fries', 3.99, ARRAY['potato', 'fast-food']),
('Ice Cream', 'Vanilla ice cream', 4.99, ARRAY['dessert', 'sweet']),
('Sandwich', 'Club sandwich', 7.99, ARRAY['bread', 'snack']);

-- Вставка данных в inventory
INSERT INTO inventory (name, stock_level, reorder_level) VALUES
('Cheese', 100,  10),
('Beef', 50,  5),
('Pasta', 200, 20),
('Lettuce', 80, 8),
('Salmon', 60,  6),
('Steak Meat', 40, 4),
('Chicken', 70,  7),
('Potatoes', 150, 15),
('Milk', 90, 9),
('Bread', 120, 12);

-- Вставка данных в menu_item_ingredients
INSERT INTO menu_item_ingredients (menu_item_id, inventory_id, quantity) VALUES
(1, 1, 2),  -- Pizza -> Cheese
(1, 10, 1), -- Pizza -> Bread

(2, 2, 1),  -- Burger -> Beef
(2, 1, 1),  -- Burger -> Cheese
(2, 10, 1), -- Burger -> Bread

(3, 3, 1),  -- Pasta -> Pasta
(3, 9, 1),  -- Pasta -> Milk

(4, 4, 1),  -- Salad -> Lettuce
(4, 8, 1),  -- Salad -> Potatoes

(5, 5, 2),  -- Sushi -> Salmon
(5, 10, 1), -- Sushi -> Bread

(6, 6, 1),  -- Steak -> Steak Meat
(6, 8, 2),  -- Steak -> Potatoes

(7, 7, 1),  -- Soup -> Chicken
(7, 9, 1),  -- Soup -> Milk

(8, 8, 1),  -- Fries -> Potatoes
(8, 9, 1),  -- Fries -> Milk

(9, 9, 2),  -- Ice Cream -> Milk
(9, 1, 1),  -- Ice Cream -> Cheese

(10, 10, 2), -- Sandwich -> Bread
(10, 1, 1),  -- Sandwich -> Cheese
(10, 2, 1);  -- Sandwich -> Beef

-- Вставка данных в order_items
INSERT INTO order_items (menu_item_id, order_id, customizations, price_at_order_time, quantity) VALUES
(6, 1, '{"extra_cheese": true}', 12.99, 5),
(6, 2, '{"no_onions": true}', 8.99, 7),
(5, 3, '{"gluten_free": true}', 10.99, 4),
(4, 4, '{"extra_dressing": true}', 6.99, 2),
(3, 5, '{"spicy": true}', 15.99, 3),
(3, 6, '{"medium_rare": true}', 24.99, 2),
(1, 7, '{"extra_sauce": true}', 5.99, 1),
(2, 8, '{"no_salt": true}', 3.99, 4),
(9, 9, '{"extra_sprinkles": true}', 4.99, 6),
(10, 10, '{"no_mayo": true}', 7.99, 10);

-- Вставка данных в inventory_transactions
INSERT INTO inventory_transactions (inventory_id, quantity, transaction_date) VALUES
(1, 10,NOW()),
(2, 5, NOW()),
(3, 20,NOW()),
(4, 15,NOW()),
(5, 8, NOW()),
(6, 4, NOW()),
(7, 10,NOW()),
(8, 25,NOW()),
(9, 30,NOW()),
(10,20,NOW());

-- Вставка данных в order_status_history
INSERT INTO order_status_history (order_id, status, changed_at) VALUES
(1, 'active', NOW()),
(2, 'closed', NOW()),
(3, 'active', NOW()),
(4, 'closed', NOW()),
(5, 'active', NOW()),
(6, 'closed', NOW()),
(7, 'active', NOW()),
(8, 'closed', NOW()),
(9, 'active', NOW()),
(10, 'closed', NOW());

-- Вставка данных в price_history
INSERT INTO price_history (menu_item_id, old_price, new_price, changed_at) VALUES
(1, 10.99, 12.99, NOW()),
(2, 7.99, 8.99, NOW()),
(3, 9.99, 10.99, NOW()),
(4, 5.99, 6.99, NOW()),
(5, 14.99, 15.99, NOW()),
(6, 22.99, 24.99, NOW()),
(7, 4.99, 5.99, NOW()),
(8, 2.99, 3.99, NOW()),
(9, 3.99, 4.99, NOW()),
(10, 6.99, 7.99, NOW());