-- Address(GENESIS): 0x010000000000000000000000000000000000000000000000000000000000000001
-- Address(MINT):    0x020000000000000000000000000000000000000000000000000000000000000002
-- Address(USER):    0x020000000000000000000000000000000000000000000000000000000000000003

-- Create tx block  address Mint transfer 2,000,000 to user
INSERT INTO block (created_at,updated_at,hash,version,block_type,prev,tstamp,merkle,height,mint_tx_id,sig) VALUES (
	'2018-09-06 07:08:24.000',
	'2018-09-06 07:08:24.000',
	0x0000000000000000000000000000000000000000000000000000000000000004,
	1,
	0x0001,
	0x0000000000000000000000000000000000000000000000000000000000000003,
	1536211979,
	'',
	3,
	0x0004000000000000000000000000000000000000000000000000000000000001,
	NULL
);

INSERT INTO tx (created_at,updated_at,hash,version,tx_type,block_hash,block_height,lock_until,amount,fee,send_to,data,sig,sender,`change`,inputs) VALUES (
	'2018-09-06 07:08:24.000',
	'2018-09-06 07:08:24.001',
	0x0004000000000000000000000000000000000000000000000000000000000001,
	1,
	0x0300,
	0x0000000000000000000000000000000000000000000000000000000000000004,
	3,
	0,
	15000100,
	0,
	0x020000000000000000000000000000000000000000000000000000000000000002,
	NULL,
	NULL,
	NULL,
	0,
	NULL
), (
	'2018-09-06 07:08:24.000',
	'2018-09-06 07:08:24.002',
	0x0004000000000000000000000000000000000000000000000000000000000002,
	1,
	0x0000,
	0x0000000000000000000000000000000000000000000000000000000000000004,
	3,
	0,
	2000000,
	100,
	0x020000000000000000000000000000000000000000000000000000000000000003,
	NULL,
	NULL,
	0x020000000000000000000000000000000000000000000000000000000000000002,
	21999900,
	0x000300000000000000000000000000000000000000000000000000000000000201000300000000000000000000000000000000000000000000000000000000000100
);

DELETE from utxo where tx_hash = 0x0003000000000000000000000000000000000000000000000000000000000002 and utxo.out = 1;
DELETE from utxo where tx_hash = 0x0003000000000000000000000000000000000000000000000000000000000001 and utxo.out = 0;

INSERT INTO utxo (created_at,updated_at,tx_hash,destination,amount,block_height,`out`) VALUES (
	'2018-09-06 07:08:24.000',
	'2018-09-06 07:08:24.003',
	0x0004000000000000000000000000000000000000000000000000000000000001,
	0x020000000000000000000000000000000000000000000000000000000000000002,
	15000100,
	3,
	0
);

INSERT INTO utxo (created_at,updated_at,tx_hash,destination,amount,block_height,`out`) VALUES (
	'2018-09-06 07:08:24.000',
	'2018-09-06 07:08:24.000',
	0x0004000000000000000000000000000000000000000000000000000000000002,
	0x020000000000000000000000000000000000000000000000000000000000000003,
	2000000,
	3,
	0
);

INSERT INTO utxo (created_at,updated_at,tx_hash,destination,amount,block_height,`out`) VALUES (
	'2018-09-06 07:08:24.000',
	'2018-09-06 07:08:24.000',
	0x0004000000000000000000000000000000000000000000000000000000000002,
	0x020000000000000000000000000000000000000000000000000000000000000002,
	21999900,
	3,
	1
);
