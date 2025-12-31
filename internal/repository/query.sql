-- name: GetUserDetails :one
select *
from users_details
where user_uuid = $1
limit 1;

-- name: GetUserContacts :one
select *
from users_contacts
where user_uuid = $1
limit 1;

-- name: CreateUser :one
insert into users (uuid)
values ($1)
returning uuid;

-- name: CreateUserDetails :one
insert into users_details (name, surname, patronymic, group_code, user_uuid)
values ($1, $2, $3, $4, $5)
returning *;

-- name: CreateUserContacts :one
insert into users_contacts (phone_number, email, telegram_id, user_uuid)
values ($1, $2, $3, $4)
returning *;

-- name: UpdateUserName :one
update users_details
set name = $2
where user_uuid = $1
returning *;

-- name: UpdateUserSurname :one
update users_details
set surname = $2
where user_uuid = $1
returning *;

-- name: UpdateUserPatronymic :one
update users_details
set patronymic = $2
where user_uuid = $1
returning *;

-- name: UpdateUserGroupCode :one
update users_details
set group_code = $2
where user_uuid = $1
returning *;

-- name: UpdateUserPhoneNumber :one
update users_contacts
set phone_number = $2
where user_uuid = $1
returning *;

-- name: UpdateUserEmail :one
update users_contacts
set email = $2
where user_uuid = $1
returning *;

-- name: UpdateUserTelegramID :one
update users_contacts
set telegram_id = $2
where user_uuid = $1
returning *;

-- name: DeleteUser :exec
delete
from users
where uuid = $1;