create table if not exists test (
    id serial primary key,
    name varchar(50)
);

insert into test (name) values ('test name 1'), ('test name 2'), ('test name 3');