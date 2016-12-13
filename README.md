btcscryer
=========
Record attempts to double spend Bitcoin.

First run the (old) btcd fork located [here](https://github.com/kcking/btcd).
Then run `btcscryer` and connect it to `btcd`.
Attempts at double spending (multiple transactions spending the same coins) will be logged at `{btcd_data_dir}/doublespends.log`.
