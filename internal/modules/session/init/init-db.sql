create database session_db;
create user session_service with password null;
grant all on database session_db to session_service;