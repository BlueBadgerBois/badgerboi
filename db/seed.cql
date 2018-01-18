create keyspace badgerbois with replication = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };
create table badgerbois.users(timeline text, id UUID, text text, PRIMARY KEY(id));
