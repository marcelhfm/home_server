alter table logs add column timestamp TIMESTAMPTZ;
update logs set timestamp = '2024-06-30T17:47:34+00:00';
alter table logs alter column timestamp set not null;
