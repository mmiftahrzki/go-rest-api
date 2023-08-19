CREATE DATABASE portfolio;

create table seller(
id int auto_increment,
name varchar(100) not null,
primary key(id)
);

create table customer(
id INT auto_increment,
name VARCHAR(100) not null,
primary key(id)
);

create table phone(
id int auto_increment,
customer_id int not null,
number varchar(15) not null,
primary key(id),
index(customer_id),
foreign key(customer_id) references customer(id)
);

create table unit(
id int auto_increment,
name varchar(100) not null,
primary key(id)
);

create table product(
id int auto_increment,
name varchar(100) not null,
netto float unsigned default 0.0 not null,
unit_id int,
price int unsigned default 0 not null,
primary key(id),
index(unit_id),
foreign key(unit_id) references unit(id)
);

CREATE TABLE invoice(
id INT auto_increment,
customer_id INT,
total int default 0 not null,
primary key(id),
index(customer_id),
foreign key (customer_id) references customer(id)
);

create table _order(
	invoice_id int,
    product_id int,
    date datetime not null default now(),
    customer_id int,
    qty int unsigned not null default 1,
    primary key(invoice_id, prod_order_orderuct_id),
    index(invoice_id),
    index(product_id),
    index(customer_id),
    foreign key(invoice_id) references invoice(id),
    foreign key(product_id) references product(id),
    foreign key(customer_id) references customer(id)
);

alter table _order add column price int unsigned not null default 0;

insert into seller(name) values("Miftah");
insert into customer(name) values("A Amir");
insert into phone(customer_id, number) values(2, "8988867411");
insert into unit(name) values("rupiah");
insert into product(name, netto, unit_id, price) values("GoPay", 10000, 1, 10500);
insert into invoice(customer_id) values(1);
insert into _order(invoice_id, product_id, date, customer_id, qty, price) values(2, 2, now(), 2, 10, 105000);

select * from seller;
select * from customer;
select * from phone;
select * from unit;
select * from _order;

DELETE FROM customer WHERE id in (10, 11);

select _order.invoice_id, product.name product_name, _order.date order_date, customer.name customer_name, _order.qty, _order.price
from _order
	join product on product.id=_order.product_id
	join customer on customer.id=_order.customer_id
order by _order.invoice_id asc;

select invoice.id, customer.name customer_name, invoice.total
from invoice
	join customer on customer.id=invoice.customer_id
order by invoice.id asc;

select customer.name, phone.number from customer left join phone on phone.customer_id=customer.id;
select product.id, product.name, product.netto, unit.name, product.price from product left join unit on unit.id=product.unit_id;
select _order.date order_date, customer.name customer_name, product.name product_name, _order.qty from _order left join customer on customer.id=_order.customer_id left join product on product.id=_order.product_id;

select invoice.id invoice_id, invoice_owner.name invoice_owner_name, invoice.total, product.name product_name, _order.date order_date, order_owner.name order_owner_name, _order.qty, _order.price
from invoice
	join _order on _order.invoice_id=invoice.id
	join customer invoice_owner on invoice_owner.id=invoice.customer_id
    join customer order_owner on order_owner.id=_order.customer_id
	join product on product.id=_order.product_id
-- where invoice_owner.id=2
;

update product set name="GoPay" where id=2;
update _order set price=157500 where invoice_id=1 and product_id=1;

delete from unit where id=2;

ALTER TABLE user ADD COLUMN created_at TIMESTAMP DEFAULT NULL;