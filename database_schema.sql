create database labor_db;
use labor_db;
create table requests
(
    id               int                                  auto_increment primary key,
    user_id          bigint                               not null,
    executor_id      bigint                                        default null,
    user_name        varchar(255)                         not null,
    executor_name    varchar(255)                                  default null,
    request_desc     longtext                             not null,
    state            enum ('free', 'in_work', 'complete') not null default 'free',
    price            decimal(7, 2)                        not null,
    creation_time    datetime                             not null default current_timestamp,
    completion_time  text                                 not null
)