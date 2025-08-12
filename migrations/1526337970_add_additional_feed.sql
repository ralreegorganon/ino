-- +goose Up
insert into feed (remote_address) values ('153.44.253.27:5631'); 

-- +goose Down
delete from feed where remote_address = '153.44.253.27:5631';
