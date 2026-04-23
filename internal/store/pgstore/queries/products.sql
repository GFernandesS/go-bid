-- name: CreateProduct :one
INSERT INTO products(seller_id, product_name, description, base_price, auction_end)
VALUES ($1, $2, $3, $4, $5)
RETURNING id;


-- name: ListProducts :many
SELECT id, product_name, description, base_price, auction_end FROM products;