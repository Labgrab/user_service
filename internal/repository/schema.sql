create table public.users
(
    uuid uuid not null,
    constraint users_pk primary key (uuid)
);

create table public.users_details
(
    name       text not null,
    surname    text not null,
    patronymic text,
    group_code text not null,
    user_uuid  uuid not null,
    constraint name_check check ((name ~ '^[\p{L}\_\-\. ]+$'::text)),
    constraint surname_check check ((surname ~ '^[\p{L}\_\-\. ]+$'::text)),
    constraint patronymic_check check ((patronymic ~ '^[\p{L}\_\-\. ]+$'::text)),
    constraint group_code_check check ((group_code ~ '^\p{L}{2,3}\-[0-9]{1,2}\-[0-9]{1,2}$'::text)),
    constraint users_details_pk primary key (user_uuid)
);

create table public.users_contacts
(
    phone_number text not null,
    email        text,
    telegram_id  bigint,
    user_uuid    uuid not null,
    constraint phone_number_check check ((phone_number ~ '^\+[1-9]\d{1,14}$'::text)),
    constraint telegram_id_check check ((telegram_id > 0)),
    constraint user_contacts_pk primary key (user_uuid)
);

alter table public.users_details
    add constraint user_uuid foreign key (user_uuid)
        references public.users (uuid) match simple
        on delete cascade on update cascade;

alter table public.users_contacts
    add constraint user_uuid foreign key (user_uuid)
        references public.users (uuid) match simple
        on delete cascade on update cascade;