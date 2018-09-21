CREATE or REPLACE FUNCTION create_item_table() RETURNS void AS $item$
    BEGIN
        create table item_table(
			item_id int primary key not null,
			item_name varchar(30) not null,
			item_quantity int not null default 0 CONSTRAINT positive_quantity CHECK (item_quantity >= 0),
			price int not null default 0 CONSTRAINT positive_price CHECK (price >= 0),
			transaction_id text default null
			);
		RAISE NOTICE 'item_table is created';
		insert into item_table values(101,'Milk',5,20),(102,'Bread',3,30),(103,'Cheese',2,120),(104,'Eggs',4,50),(105,'Oats',2,25),(106,'Others',3,40);																	
		RAISE NOTICE 'Values inserted to item_table';
		CREATE TRIGGER send_change_event
			AFTER INSERT OR DELETE OR UPDATE 
			ON item_table
			FOR EACH ROW
			EXECUTE PROCEDURE on_row_change();

    END;
	$item$ LANGUAGE plpgsql;


CREATE or REPLACE FUNCTION create_bank_table() RETURNS void AS $bank$
    BEGIN
        create table bank_table(
			card_no int primary key not null,
			customer_name text,
			balance int default 0 CONSTRAINT positive_balance CHECK (balance >= 0),
			transaction_id text default null
			);
		RAISE NOTICE 'bank_table is created';
		insert into bank_table values(1001,'Anil',500),(1002,'Dheeraj',100),(1003,'Vishu',200);
		RAISE NOTICE 'Values inserted to bank_table';
    END;
	$bank$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION send_message(channel text, routing_key text,message text) RETURNS void AS $send$
	BEGIN
		PERFORM pg_notify(channel, routing_key || '|' || message);
	END;
	$send$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION on_row_change() RETURNS trigger AS $TFC$
	BEGIN
		 DECLARE
			routing_key text;
			row record;
			temp record;
			msg text;
			ct_time text;
			ct_db text;
		 BEGIN
			routing_key := 'row_change';
			ct_db := (SELECT current_database());
			ct_time := now();
			if (TG_OP = 'DELETE') then
				row := old;
				msg := '{"operation":"'::text||TG_OP::text||'","table":"'::text||TG_TABLE_NAME::text||'","database":"'::text||ct_db::text||'","schema":"'::text||TG_TABLE_SCHEMA::text||'","timestamp":"'::text||ct_time::text||'","before":'::text||row_to_json(row)::text||',"after":{"item_id":'::text||'"null"}'::text||'}';
			elsif (TG_OP = 'UPDATE') then
				temp := old;
				row := new;
				msg := '{"operation":"'::text||TG_OP::text||'","table":"'::text||TG_TABLE_NAME::text||'","database":"'::text||ct_db::text||'","schema":"'::text||TG_TABLE_SCHEMA::text||'","timestamp":"'::text||ct_time::text||'","before":'::text||row_to_json(temp)::text||',"after":'::text||row_to_json(row)::text||'}';
			elsif (TG_OP = 'INSERT') then
				row := new;
				msg := '{"operation":"'::text||TG_OP::text||'","table":"'::text||TG_TABLE_NAME::text||'","database":"'::text||ct_db::text||'","schema":"'::text||TG_TABLE_SCHEMA::text||'","timestamp":"'::text||ct_time::text||'","before":{"item_id":'::text||'"null"}'::text||',"after":'::text||row_to_json(row)::text||'}';
			end if;
			-- change 'events' to the desired channel/exchange name
			PERFORM send_message('pgchannel', routing_key, msg);
			return null;
		END;
	END $TFC$ LANGUAGE plpgsql;
	
DO $$
DECLARE
	db_exist bool;
	item_table_exist bool;
	bank_table_exist bool;
BEGIN
	item_table_exist := (SELECT EXISTS ( SELECT 1 FROM  information_schema.tables WHERE  table_schema = 'public' AND table_name = 'item_table'));
	bank_table_exist := (SELECT EXISTS ( SELECT 1 FROM  information_schema.tables WHERE  table_schema = 'public' AND table_name = 'bank_table'));
	IF (item_table_exist = false) THEN
		RAISE NOTICE 'Item table exists : FALSE';
		PERFORM create_item_table();
	ELSE
		RAISE NOTICE 'Item table exists : TRUE';
	END IF;
	IF (bank_table_exist = false) THEN
		RAISE NOTICE 'Bank table exists : FALSE';
		PERFORM create_bank_table();
	ELSE
		RAISE NOTICE 'Bank table exists : TRUE';
	END IF;
END $$;