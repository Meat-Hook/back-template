create database user_db;
create user service_user with password null;
grant all on database user_db to service_user;
