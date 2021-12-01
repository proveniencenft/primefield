# primefield
A golang prime field implementation

This is a clean-room implementation of a field over an arbitrary prime p.
It uses Montgomery multiplication, but otherwise depends on big.Int to to all the arithmetics.
The R factor is selected as the smallest power of 2 bigger than p.
