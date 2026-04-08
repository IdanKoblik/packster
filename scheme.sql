CREATE TABLE account (
    id SERIAL PRIMARY KEY,
    display_name VARCHAR(255) NOT NULL,
    last_login TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL
);

CREATE TABLE host (
    id SERIAL PRIMARY KEY,
    url VARCHAR(255) NOT NULL
);

CREATE TABLE repository (
    id SERIAL PRIMARY KEY,
    host INT NOT NULL,
    repository INT NOT NULL,
    owner INT NOT NULL,
    created_at INT NOT NULL,

    FOREIGN KEY (host) REFERENCES host(id),
    FOREIGN KEY (owner) REFERENCES account(id)
);

CREATE INDEX idx_repository_repository ON repository(repository);

CREATE TABLE product (
    id SERIAL PRIMARY KEY,
    external_id INT NOT NULL,
    name VARCHAR(255) NOT NULL,
    group_name VARCHAR(255) NOT NULL,
    repository INT NOT NULL,
    created_at TIMESTAMP NOT NULL,

    FOREIGN KEY (repository) REFERENCES repository(id)
);

CREATE INDEX idx_product_external_id ON product(external_id);

CREATE TABLE token (
    id SERIAL PRIMARY KEY,
    value VARCHAR(255) NOT NULL,
    expiry TIMESTAMP NOT NULL,
    owner INT NOT NULL,
    created_at TIMESTAMP NOT NULL,

    FOREIGN KEY (owner) REFERENCES account(id)
);

CREATE TABLE token_access (
    token INT NOT NULL,
    product INT NOT NULL,

    PRIMARY KEY (token, product),

    FOREIGN KEY (token) REFERENCES token(id),
    FOREIGN KEY (product) REFERENCES product(id)
);

CREATE TABLE auth (
    id INT PRIMARY KEY,
    account INT NOT NULL,
    username VARCHAR(255) NOT NULL,
    sso_id INT NOT NULL,
    host INT NOT NULL,

    FOREIGN KEY (account) REFERENCES account(id),
    FOREIGN KEY (host) REFERENCES host(id)
);

CREATE TABLE permission (
    account INT NOT NULL,
    repository INT NOT NULL,
    can_download BOOLEAN NOT NULL,
    can_upload BOOLEAN NOT NULL,
    can_delete BOOLEAN NOT NULL,

    PRIMARY KEY (account, repository),

    FOREIGN KEY (account) REFERENCES account(id),
    FOREIGN KEY (repository) REFERENCES repository(id)
);

