truncate log_items,stock_holdings,transaction_nums,transactions,triggers,users;
alter sequence log_items_id_seq restart with 1;
alter sequence stock_holdings_id_seq restart with 1;
alter sequence transaction_nums_id_seq restart with 1;
alter sequence transaction_nums_transaction_id_seq restart with 1;
alter sequence transactions_id_seq restart with 1;
alter sequence triggers_id_seq restart with 1;
alter sequence users_id_seq restart with 1;
