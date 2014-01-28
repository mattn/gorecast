create table if not exists metrics (
    id           integer primary key,
    service_name varchar(255) not null ,
    section_name varchar(255) not null ,
    graph_name   varchar(255) not null ,
    sort         uint not null default 0,
    meta         text not null,
    created_at   datetime not null,
    updated_at   timestamp not null,
    unique(service_name, section_name, graph_name)
);

create table if not exists data (
    metrics_id    integer primary key,
    datetime      datetime not null,
    number        bigint not null,
    updated_at    timestamp not null,
    primary key (metrics_id, datetime),
    unique(datetime)
);

create table if not exists complex (
    id           integer primary key,
    service_name varchar(255) not null,
    section_name varchar(255) not null,
    graph_name   varchar(255) not null,
    sort         int not null default 0,
    meta         text not null,
    created_at   datetime not null,
    updated_at   timestamp not null,
    primary key (id),
    unique(service_name, section_name, graph_name)
);
