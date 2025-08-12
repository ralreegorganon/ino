-- +goose Up
alter table message add raw character varying not null default '';

-- +goose Down
alter table message drop column raw;
