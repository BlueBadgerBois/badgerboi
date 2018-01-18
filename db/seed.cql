create keyspace badgerboi with replication = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };
create table badgerboi.users(timeline text, id UUID, text text, PRIMARY KEY(id));
