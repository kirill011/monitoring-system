proto_gen:
	protoc --proto_path=proto/getmail --go_out=proto/getmail --go_opt=paths=source_relative getmail.proto
	protoc --proto_path=proto/getresponsible --go_out=proto/getresponsible --go_opt=paths=source_relative getresponsible.proto
	protoc --proto_path=proto/sendnotify --go_out=proto/sendnotify --go_opt=paths=source_relative sendnotify.proto


start_service_rebuild:
	docker compose up --build notifycation-service

start_service:
	docker compose up