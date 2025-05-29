proto_gen:
	protoc --proto_path=proto/users --go_out=proto/users --go_opt=paths=source_relative users.proto
	protoc --proto_path=proto/devices --go_out=proto/devices --go_opt=paths=source_relative devices.proto
	protoc --proto_path=proto/notification --go_out=proto/notification --go_opt=paths=source_relative notification.proto


start_service_rebuild:
	docker compose up --build notifycation-service

start_service:
	docker compose up