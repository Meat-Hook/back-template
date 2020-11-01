create database session_db;
create user service_session with password null;
grant all on database session_db to service_session;