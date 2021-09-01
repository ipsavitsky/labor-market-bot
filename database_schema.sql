create database labor_db;
use labor_db;
create table requests
(
    id               int not null auto_increment,
    user_id          bigint,
    executor_id      bigint                               default null,
    user_name        varchar(255),
    executor_name    varchar(255)                         default null,
    user_chat_id     bigint,
    executor_chat_id bigint                               default null,
    request_desc     longtext,
    state            enum ('free', 'in_work', 'complete') default 'free',
    price            decimal(7, 2),
    creation_time    datetime                             default current_timestamp,
    completion_time  text,
    primary key (id)
)