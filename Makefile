.PHONY: up down create-table run

up:
	docker compose up -d

down:
	docker compose down

create-table:
	aws dynamodb create-table \
		--endpoint-url http://localhost:8000 \
		--table-name items \
		--attribute-definitions \
			AttributeName=user_id,AttributeType=S \
			AttributeName=created_at,AttributeType=S \
		--key-schema \
			AttributeName=user_id,KeyType=HASH \
			AttributeName=created_at,KeyType=RANGE \
		--billing-mode PAY_PER_REQUEST \
		--region us-east-1

run:
	go run .
