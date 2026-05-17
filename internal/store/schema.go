package store

const schemaVersion = 2

const schemaSQL = `
create table if not exists institutions (
	id text primary key,
	provider text not null,
	provider_institution_id text not null,
	name text not null,
	country text not null,
	bic text,
	raw_json blob,
	updated_at text not null,
	unique(provider, provider_institution_id)
);

create table if not exists connections (
	id text primary key,
	provider text not null,
	provider_connection_id text not null,
	institution_id text not null,
	status text not null,
	redirect_url text,
	created_at text not null,
	updated_at text not null,
	expires_at text,
	raw_json blob,
	unique(provider, provider_connection_id)
);

create table if not exists accounts (
	id text primary key,
	provider text not null,
	provider_account_id text not null,
	provider_resource_id text,
	institution_id text,
	connection_id text,
	iban text,
	name text,
	currency text,
	owner_name text,
	raw_json blob,
	updated_at text not null,
	unique(provider, provider_account_id)
);

create table if not exists transactions (
	id text primary key,
	dedupe_key text not null unique,
	provider text not null,
	provider_transaction_id text,
	account_id text not null,
	booking_date text not null,
	value_date text,
	amount text not null,
	currency text not null,
	counterparty_name text,
	counterparty_account text,
	description text,
	remittance_info text,
	reference text,
	raw_json blob,
	created_at text not null,
	updated_at text not null
);

create index if not exists idx_transactions_account_date on transactions(account_id, booking_date);
create index if not exists idx_transactions_provider_tx on transactions(provider, provider_transaction_id);
create index if not exists idx_transactions_reference on transactions(provider, account_id, reference);

create table if not exists sync_runs (
	id text primary key,
	provider text not null,
	connection_id text,
	account_id text,
	started_at text not null,
	finished_at text,
	status text not null,
	error text,
	transactions_new integer not null default 0,
	transactions_seen integer not null default 0
);
`
