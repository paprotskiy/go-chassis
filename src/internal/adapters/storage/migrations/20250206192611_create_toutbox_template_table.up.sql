create table toutbox.toutbox_template
(
    "idempotence_key"  uuid                     not null primary key,
    "data"             jsonb                    not null,
    "created"          timestamp with time zone not null,
    "transferred"      timestamp with time zone 
);

create index toutbox_example_created_index on toutbox.toutbox_template (created);
create index toutbox_example_transferred_index on toutbox.toutbox_template (transferred);
