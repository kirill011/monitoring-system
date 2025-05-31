CREATE OR REPLACE FUNCTION bd_tr_tags()
 RETURNS trigger
 LANGUAGE plpgsql
AS $function$
	BEGIN
		update tags
		set deleted_at = current_timestamp
		where id = old.id;

		return NULL;
	END;
$function$
;

create or replace trigger tr_bd_tags before
delete
	on
	tags for each row execute function bd_tr_tags();