alter table datasources add column type TEXT;
update datasources set type = 'UNKNOWN';
alter table datasources alter column type set not null;
alter table datasources alter column type set default 'UNKNOWN';
