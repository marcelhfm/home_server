alter table datasources add column status text;
update datasources set status = 'DISCONNECTED';
alter table datasources alter column status set not null;
alter table datasources alter column status set default 'DISCONNECTED';
