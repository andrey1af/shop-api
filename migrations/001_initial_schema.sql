CREATE TABLE addresses (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    country varchar(100) NOT NULL,
    city varchar(100) NOT NULL,
    street varchar(255) NOT NULL,

    CONSTRAINT addresses_country_not_blank CHECK (btrim(country) <> ''),
    CONSTRAINT addresses_city_not_blank CHECK (btrim(city) <> ''),
    CONSTRAINT addresses_street_not_blank CHECK (btrim(street) <> '')
);

CREATE TABLE clients (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    client_name varchar(100) NOT NULL,
    client_surname varchar(100) NOT NULL,
    birthday date NOT NULL,
    gender varchar(50) NOT NULL,
    registration_date date NOT NULL DEFAULT CURRENT_DATE,
    address_id uuid NOT NULL,

    CONSTRAINT clients_name_not_blank CHECK (btrim(client_name) <> ''),
    CONSTRAINT clients_surname_not_blank CHECK (btrim(client_surname) <> ''),
    CONSTRAINT clients_gender_not_blank CHECK (btrim(gender) <> ''),
    CONSTRAINT clients_birthday_before_registration CHECK (birthday <= registration_date),
    CONSTRAINT clients_address_unique UNIQUE (address_id),
    CONSTRAINT clients_address_fk
        FOREIGN KEY (address_id) REFERENCES addresses (id) ON DELETE RESTRICT
);

CREATE INDEX idx_clients_name_surname
    ON clients (client_name, client_surname);

CREATE TABLE suppliers (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name varchar(255) NOT NULL,
    phone_number varchar(30) NOT NULL,
    address_id uuid NOT NULL,

    CONSTRAINT suppliers_name_not_blank CHECK (btrim(name) <> ''),
    CONSTRAINT suppliers_phone_not_blank CHECK (btrim(phone_number) <> ''),
    CONSTRAINT suppliers_address_unique UNIQUE (address_id),
    CONSTRAINT suppliers_address_fk
        FOREIGN KEY (address_id) REFERENCES addresses (id) ON DELETE RESTRICT
);

CREATE TABLE product_categories (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name varchar(100) NOT NULL,

    CONSTRAINT product_categories_name_not_blank CHECK (btrim(name) <> ''),
    CONSTRAINT product_categories_name_unique UNIQUE (name)
);

CREATE TABLE products (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name varchar(255) NOT NULL,
    category_id uuid NOT NULL,
    price numeric(12, 2) NOT NULL,
    available_stock bigint NOT NULL,
    last_update_date date NOT NULL DEFAULT CURRENT_DATE,
    supplier_id uuid NOT NULL,

    CONSTRAINT products_name_not_blank CHECK (btrim(name) <> ''),
    CONSTRAINT products_price_non_negative CHECK (price >= 0),
    CONSTRAINT products_stock_non_negative CHECK (available_stock >= 0),
    CONSTRAINT products_category_fk
        FOREIGN KEY (category_id) REFERENCES product_categories (id) ON DELETE RESTRICT,
    CONSTRAINT products_supplier_fk
        FOREIGN KEY (supplier_id) REFERENCES suppliers (id) ON DELETE RESTRICT
);

CREATE INDEX idx_products_category_id
    ON products (category_id);

CREATE INDEX idx_products_supplier_id
    ON products (supplier_id);

CREATE INDEX idx_products_available
    ON products (id)
    WHERE available_stock > 0;

CREATE TABLE images (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id uuid NOT NULL,
    image bytea NOT NULL,

    CONSTRAINT images_data_not_empty CHECK (octet_length(image) > 0),
    CONSTRAINT images_product_unique UNIQUE (product_id),
    CONSTRAINT images_product_fk
        FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE CASCADE
);
