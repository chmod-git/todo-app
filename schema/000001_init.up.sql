CREATE TABLE users
(
    id SERIAL NOT NULL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    username VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL
);

CREATE TABLE todo_list
(
    id SERIAL NOT NULL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description VARCHAR(255) NOT NULL,
    user_id INTEGER NOT NULL,
    CONSTRAINT user_id_fk FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE todo_item
(
    id SERIAL NOT NULL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description VARCHAR(255) NOT NULL,
    done BOOLEAN NOT NULL DEFAULT FALSE,
    list_id INTEGER NOT NULL,
    CONSTRAINT list_id_fk FOREIGN KEY (list_id) REFERENCES todo_list(id) MATCH FULL
);