proto_gen:
	protoc --proto_path=proto --go_out=proto --go_opt=paths=source_relative getmail.proto

start_service_rebuild:
	docker compose up --build auth-service

start_service:
	docker compose up