ALTER TABLE IF EXISTS "accounts" DROP CONSTRAINT IF EXISTS "owner_currency_key";

ALTER TABLE IF EXISTS "accounts" DROP CONSTRAINT IF EXISTS "accounts_owner_fkey"; 
-- names of foreign key constraints, like accounts_owner_fkey" are auto-generated
-- you can easily find them out by inspecting the db with a tool like dbplus or pgadmin 

DROP TABLE IF EXISTS "users";

